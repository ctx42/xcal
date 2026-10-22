package tzgen

import (
	"cmp"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/ctx42/xcal/pkg/xcal"
)

// Updater refreshes a zone registry from the IANA tz database release that a
// Go toolchain embeds.
type Updater struct {
	// Client downloads the release; nil means [http.DefaultClient].
	Client *http.Client

	// URL is the release URL format; empty means [ReleaseURL].
	URL string

	// GoRoot is the root of the Go toolchain.
	GoRoot string

	// DataDir holds [ZonesFile] and [TZDataDir].
	DataDir string

	// Year is the year zone offsets are taken from; zero means this year.
	Year int
}

// Summary describes a registry update.
type Summary struct {
	Version string   // Version is the tz database release used.
	Added   []string // Added names the zones that got new IDs.
	Changed []string // Changed names the zones whose details changed.
}

// Run fetches the release files into the tzdata directory unless it already
// holds that release, merges the release into the registry, and writes the
// registry back.
func (upd *Updater) Run(ctx context.Context) (Summary, error) {
	ver, err := GoTZVersion(upd.GoRoot)
	if err != nil {
		return Summary{}, err
	}

	dir := filepath.Join(upd.DataDir, TZDataDir)
	if !haveRelease(dir, ver) {
		cli := cmp.Or(upd.Client, http.DefaultClient)
		url := cmp.Or(upd.URL, ReleaseURL)
		if err = Fetch(ctx, cli, url, ver, dir); err != nil {
			return Summary{}, err
		}
	}

	zip := filepath.Join(upd.GoRoot, "lib", "time", "zoneinfo.zip")
	year := cmp.Or(upd.Year, time.Now().Year())
	ups, err := ReadUpstream(dir, zip, year)
	if err != nil {
		return Summary{}, err
	}
	pth := filepath.Join(upd.DataDir, ZonesFile)
	prev, err := DecodeZones(pth)
	if err != nil {
		return Summary{}, err
	}
	zs, err := Merge(prev, ups)
	if err != nil {
		return Summary{}, err
	}
	if err = WriteZones(pth, zs); err != nil {
		return Summary{}, err
	}
	return summarize(ver, prev, zs), nil
}

// haveRelease reports whether dir holds every file [Fetch] extracts from the
// release ver.
func haveRelease(dir, ver string) bool {
	data, err := os.ReadFile(filepath.Join(dir, VersionFile))
	if err != nil || strings.TrimSpace(string(data)) != ver {
		return false
	}
	for _, name := range []string{LinksFile, ZoneTabFile} {
		if _, err = os.Stat(filepath.Join(dir, name)); err != nil {
			return false
		}
	}
	return true
}

// summarize describes the change from the prev zones to zs.
func summarize(ver string, prev, zs []*xcal.Zone) Summary {
	byID := make(map[int]*xcal.Zone, len(prev))
	for _, zon := range prev {
		byID[zon.Id] = zon
	}
	sum := Summary{Version: ver}
	for _, zon := range zs {
		old, ok := byID[zon.Id]
		switch {
		case !ok:
			sum.Added = append(sum.Added, zon.Name)
		case !reflect.DeepEqual(old, zon):
			sum.Changed = append(sum.Changed, zon.Name)
		}
	}
	return sum
}
