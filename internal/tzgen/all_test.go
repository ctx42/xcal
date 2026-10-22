package tzgen

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"sync/atomic"

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

// newGoRoot creates a Go root in a temporary directory, embedding the 2026c
// tz database release and the zones of testdata/zoneinfo.zip, and returns its
// path.
func newGoRoot(t tester.T) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "lib", "time")
	must.Nil(os.MkdirAll(dir, 0755))
	script := "CODE=2026c\nDATA=2026c\n"
	pth := filepath.Join(dir, "update.bash")
	must.Nil(os.WriteFile(pth, []byte(script), 0644))
	data := must.Value(os.ReadFile("testdata/zoneinfo.zip"))
	pth = filepath.Join(dir, "zoneinfo.zip")
	must.Nil(os.WriteFile(pth, data, 0644))
	return root
}

// canceledContext returns a canceled context derived from t.Context().
func canceledContext(t tester.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	return ctx
}

// zipEntry returns the entry of testdata/zoneinfo.zip with the given name.
func zipEntry(t tester.T, name string) *zip.File {
	t.Helper()
	zr := must.Value(zip.OpenReader("testdata/zoneinfo.zip"))
	t.Cleanup(func() { _ = zr.Close() })
	for _, fil := range zr.File {
		if fil.Name == name {
			return fil
		}
	}
	t.Fatalf("no %s in testdata/zoneinfo.zip", name)
	return nil
}

// newRelease starts a server returning the testdata/tzdata release and
// counting its requests.
func newRelease(t tester.T, reqs *atomic.Int32) *httptest.Server {
	t.Helper()
	read := func(name string) string {
		pth := filepath.Join("testdata", "tzdata", name)
		return string(must.Value(os.ReadFile(pth)))
	}
	body := tarGz(t, map[string]string{
		"version":  read(VersionFile),
		"zone.tab": read(ZoneTabFile),
		"backward": read(LinksFile),
	})
	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			reqs.Add(1)
			_, _ = w.Write(body)
		},
	))
	t.Cleanup(srv.Close)
	return srv
}

// newDataDir creates a data directory holding testdata/timezones.json and
// returns its path.
func newDataDir(t tester.T) string {
	t.Helper()
	dir := t.TempDir()
	data := must.Value(os.ReadFile("testdata/timezones.json"))
	must.Nil(os.WriteFile(filepath.Join(dir, ZonesFile), data, 0644))
	return dir
}
