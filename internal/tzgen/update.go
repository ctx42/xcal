package tzgen

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/ctx42/xcal/pkg/xcal"
)

// Updater refreshes a zone registry from the IANA tz database release that a
// Go toolchain embeds.
type Updater struct {
	Client  *http.Client // Client downloads the release.
	URL     string       // URL is the release URL format, see [ReleaseURL].
	GoRoot  string       // GoRoot is the root of the Go toolchain.
	DataDir string       // DataDir holds [ZonesFile] and [TZDataDir].
	Year    int          // Year is the year zone offsets are taken from.
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
	have, _ := os.ReadFile(filepath.Join(dir, VersionFile))
	if strings.TrimSpace(string(have)) != ver {
		if err = Fetch(ctx, upd.Client, upd.URL, ver, dir); err != nil {
			return Summary{}, err
		}
	}

	zip := filepath.Join(upd.GoRoot, "lib", "time", "zoneinfo.zip")
	ups, err := ReadUpstream(dir, zip, upd.Year)
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
