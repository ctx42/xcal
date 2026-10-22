package tzgen

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testing/pkg/tester"

	"github.com/ctx42/xcal/pkg/xcal"
)

// testPrev returns the zones of testdata/timezones.json by name.
func testPrev(t tester.T) map[string]*xcal.Zone {
	t.Helper()
	zs := must.Value(ReadZones("testdata/timezones.json"))
	prev := make(map[string]*xcal.Zone, len(zs))
	for _, zon := range zs {
		prev[zon.Name] = zon
	}
	return prev
}

// tarGz returns a gzip-compressed tar archive of files, by name.
func tarGz(t tester.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, name := range slices.Sorted(maps.Keys(files)) {
		hdr := &tar.Header{
			Name:     name,
			Mode:     0644,
			Size:     int64(len(files[name])),
			Typeflag: tar.TypeReg,
		}
		must.Nil(tw.WriteHeader(hdr))
		must.Value(tw.Write([]byte(files[name])))
	}
	must.Nil(tw.Close())
	must.Nil(gz.Close())
	return buf.Bytes()
}

// writeZip writes a zip archive of files, by name, to a temporary directory
// and returns its path.
func writeZip(t tester.T, files map[string][]byte) string {
	t.Helper()
	pth := filepath.Join(t.TempDir(), "zoneinfo.zip")
	fil := must.Value(os.Create(pth))
	zw := zip.NewWriter(fil)
	for _, name := range slices.Sorted(maps.Keys(files)) {
		must.Value(must.Value(zw.Create(name)).Write(files[name]))
	}
	must.Nil(zw.Close())
	must.Nil(fil.Close())
	return pth
}

// newGoRoot creates a Go root in a temporary directory, embedding the tz
// database release version and the zones of testdata/zoneinfo.zip, and
// returns its path.
func newGoRoot(t tester.T, version string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "lib", "time")
	must.Nil(os.MkdirAll(dir, 0755))
	script := "CODE=" + version + "\nDATA=" + version + "\n"
	pth := filepath.Join(dir, "update.bash")
	must.Nil(os.WriteFile(pth, []byte(script), 0644))
	data := must.Value(os.ReadFile("testdata/zoneinfo.zip"))
	pth = filepath.Join(dir, "zoneinfo.zip")
	must.Nil(os.WriteFile(pth, data, 0644))
	return root
}

// canceledContext returns a context that is already canceled.
func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}
