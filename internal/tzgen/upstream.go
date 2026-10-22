package tzgen

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/ctx42/xcal/pkg/xcal"
)

// GoTZVersion returns the IANA tz database release the Go toolchain rooted at
// goroot embeds, read from the DATA line of its lib/time/update.bash.
func GoTZVersion(goroot string) (string, error) {
	pth := filepath.Join(goroot, "lib", "time", "update.bash")
	data, err := os.ReadFile(pth) //nolint:gosec // Caller-supplied path.
	if err != nil {
		return "", err
	}
	for line := range strings.Lines(string(data)) {
		line = strings.TrimSpace(line)
		if ver, ok := strings.CutPrefix(line, "DATA="); ok && ver != "" {
			return ver, nil
		}
	}
	return "", fmt.Errorf("no DATA line in %s", pth)
}

// Fetch downloads the IANA tz database release version from url, a
// [ReleaseURL] format, and writes its [VersionFile], [LinksFile], and
// [ZoneTabFile] into dir.
func Fetch(
	ctx context.Context,
	cli *http.Client,
	url string,
	version string,
	dir string,
) error {

	url = fmt.Sprintf(url, version)
	body := http.NoBody
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, body)
	if err != nil {
		return err
	}
	rsp, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = rsp.Body.Close() }()
	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: %s", url, rsp.Status)
	}

	rel, err := readRelease(rsp.Body)
	if err != nil {
		return fmt.Errorf("read %s: %w", url, err)
	}
	if rel.version != version {
		format := "%s holds version %q, want %q"
		return fmt.Errorf(format, url, rel.version, version)
	}
	if rel.zoneTab == nil {
		return fmt.Errorf("%s has no %s", url, ZoneTabFile)
	}
	return rel.write(dir)
}

// release holds the files [Fetch] keeps from a tz database release.
type release struct {
	version string   // Release version.
	zoneTab []byte   // The zone.tab file.
	links   []string // Link lines of every data file, sorted by link name.
}

// readRelease reads a gzip-compressed tar archive of a tz database release.
func readRelease(r io.Reader) (*release, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	rel := &release{}
	tr := tar.NewReader(gz)
	var hdr *tar.Header
	for {
		hdr, err = tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		var data []byte
		if data, err = io.ReadAll(tr); err != nil {
			return nil, err
		}
		rel.add(hdr.Name, data)
	}

	slices.SortFunc(rel.links, func(a, b string) int {
		return strings.Compare(lastField(a), lastField(b))
	})
	return rel, nil
}

// add keeps the parts of the release file name the release needs.
func (rel *release) add(name string, data []byte) {
	switch {
	case name == VersionFile:
		rel.version = string(bytes.TrimSpace(data))
	case name == ZoneTabFile:
		rel.zoneTab = data
	case isDataFile(name):
		rel.links = append(rel.links, linkLines(data)...)
	}
}

// write writes the release files into dir. The version file goes last, so
// it never names a release whose files are missing.
func (rel *release) write(dir string) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	files := []struct {
		name string
		data []byte
	}{
		{LinksFile, []byte(strings.Join(rel.links, "\n") + "\n")},
		{ZoneTabFile, rel.zoneTab},
		{VersionFile, []byte(rel.version + "\n")},
	}
	for _, fil := range files {
		pth := filepath.Join(dir, fil.name)
		if err := os.WriteFile(pth, fil.data, 0o600); err != nil {
			return err
		}
	}
	return nil
}

// isDataFile reports whether name is a tz database source file compiled into
// the Go zoneinfo.zip.
func isDataFile(name string) bool {
	switch name {
	case "africa", "antarctica", "asia", "australasia", "europe",
		"northamerica", "southamerica", "etcetera", "factory", "backward":
		return true
	}
	return false
}

// linkLines returns the Link lines of a tz database source file, each as
// "Link<TAB>target<TAB>name".
func linkLines(data []byte) []string {
	var lines []string
	for line := range strings.Lines(string(data)) {
		line, _, _ = strings.Cut(line, "#")
		fields := strings.Fields(line)
		if len(fields) == 3 && fields[0] == "Link" {
			lines = append(lines, strings.Join(fields, "\t"))
		}
	}
	return lines
}

// lastField returns the last tab-separated field of line.
func lastField(line string) string {
	return line[strings.LastIndexByte(line, '\t')+1:]
}

// ReadUpstream returns a zone for every location in the Go zoneinfo.zip at
// zipPath, sorted by name, described by the [LinksFile] and [ZoneTabFile] in
// dir. A linked name is an [xcal.Alias] of the link target, every other name
// is [xcal.Canonical]. Offsets are the standard and DST offsets in effect in
// the given year. The zones have no IDs.
func ReadUpstream(dir, zipPath string, year int) ([]*xcal.Zone, error) {
	links, err := readLinks(filepath.Join(dir, LinksFile))
	if err != nil {
		return nil, err
	}
	tabs, err := readZoneTab(filepath.Join(dir, ZoneTabFile))
	if err != nil {
		return nil, err
	}

	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = zr.Close() }()

	var zs []*xcal.Zone
	for _, fil := range zr.File {
		if fil.FileInfo().IsDir() {
			continue
		}
		loc, err := loadZipLocation(fil)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", zipPath, err)
		}
		zon := &xcal.Zone{Name: fil.Name, Status: xcal.Canonical}
		if target, ok := links[fil.Name]; ok {
			zon.Status = xcal.Alias
			zon.Link = target
		}
		zon.Offset, zon.OffsetDST = offsets(loc, year)
		if tab, ok := tabs[fil.Name]; ok {
			zon.CountryCode = tab.country
			zon.Notes = tab.comment
		}
		zs = append(zs, zon)
	}

	slices.SortFunc(zs, func(a, b *xcal.Zone) int {
		return strings.Compare(a.Name, b.Name)
	})
	return zs, nil
}

// loadZipLocation loads the location stored in a zoneinfo.zip entry.
func loadZipLocation(fil *zip.File) (*time.Location, error) {
	rc, err := fil.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = rc.Close() }()
	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	return time.LoadLocationFromTZData(fil.Name, data)
}

// readLinks reads a [LinksFile] into a map of link names to their targets.
func readLinks(pth string) (map[string]string, error) {
	data, err := os.ReadFile(pth) //nolint:gosec // Caller-supplied path.
	if err != nil {
		return nil, err
	}
	links := make(map[string]string)
	for _, line := range linkLines(data) {
		fields := strings.Split(line, "\t")
		links[fields[2]] = fields[1]
	}
	return links, nil
}

// zoneTab is a zone.tab entry.
type zoneTab struct {
	country string // ISO 3166 country code.
	comment string // Comment telling zones of the country apart.
}

// readZoneTab reads a zone.tab file into a map of zone names to entries.
func readZoneTab(pth string) (map[string]zoneTab, error) {
	data, err := os.ReadFile(pth) //nolint:gosec // Caller-supplied path.
	if err != nil {
		return nil, err
	}
	tabs := make(map[string]zoneTab)
	for line := range strings.Lines(string(data)) {
		line = strings.TrimRight(line, "\r\n")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 3 {
			return nil, fmt.Errorf("%s: invalid line %q", pth, line)
		}
		tab := zoneTab{country: fields[0]}
		if len(fields) > 3 {
			tab.comment = fields[3]
		}
		tabs[fields[2]] = tab
	}
	return tabs, nil
}

// offsets returns the standard and DST offsets of loc in minutes, as the
// first offsets of each kind in effect at noon UTC on a day of year. A zone
// without DST that year returns its standard offset twice.
func offsets(loc *time.Location, year int) (int, int) {
	var std, dst int
	var haveStd, haveDST bool
	tim := time.Date(year, time.January, 1, 12, 0, 0, 0, time.UTC)
	for ; tim.Year() == year; tim = tim.AddDate(0, 0, 1) {
		local := tim.In(loc)
		_, off := local.Zone()
		if local.IsDST() {
			if !haveDST {
				dst, haveDST = off/60, true
			}
		} else if !haveStd {
			std, haveStd = off/60, true
		}
	}
	if !haveStd {
		std = dst
	}
	if !haveDST {
		dst = std
	}
	return std, dst
}
