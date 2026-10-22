package tzgen

import (
	"archive/zip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"

	"github.com/ctx42/xcal/pkg/xcal"
)

func Test_GoTZVersion(t *testing.T) {
	// --- Given ---
	root := newGoRoot(t, "2026c")

	// --- When ---
	have, err := GoTZVersion(root)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, "2026c", have)
}

func Test_GoTZVersion_noScript(t *testing.T) {
	// --- When ---
	have, err := GoTZVersion(t.TempDir())

	// --- Then ---
	assert.ErrorContain(t, "update.bash: no such file or directory", err)
	assert.Equal(t, "", have)
}

func Test_GoTZVersion_noDataLine(t *testing.T) {
	// --- Given ---
	root := t.TempDir()
	dir := filepath.Join(root, "lib", "time")
	must.Nil(os.MkdirAll(dir, 0755))
	pth := filepath.Join(dir, "update.bash")
	must.Nil(os.WriteFile(pth, []byte("CODE=2026c\n"), 0644))

	// --- When ---
	have, err := GoTZVersion(root)

	// --- Then ---
	assert.ErrorContain(t, "no DATA line in", err)
	assert.Equal(t, "", have)
}

func Test_Fetch(t *testing.T) {
	// --- Given ---
	body := tarGz(t, map[string]string{
		"version":  "2026c\n",
		"zone.tab": "PL\t+5215+02100\tEurope/Warsaw\n",
		"europe": "" +
			"Zone\tEurope/Kyiv\t2:00\tEU\tEE%sT\n" +
			"Link\tEurope/Kyiv\tEurope/Zaporozhye # merged\n",
		"backward":  "Link\tEurope/Kyiv\t\tEurope/Kiev\n",
		"backzone":  "Link\tEurope/Kyiv\tEurope/Simferopol\n",
		"README":    "Link\tA\tB\n",
		"tzdata.zi": "L Europe/Kyiv Europe/Uzhgorod\n",
	})
	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(body)
		},
	))
	t.Cleanup(srv.Close)
	dir := filepath.Join(t.TempDir(), "tzdata")

	// --- When ---
	err := Fetch(t.Context(), srv.Client(), srv.URL+"/%s", "2026c", dir)

	// --- Then ---
	assert.NoError(t, err)

	read := func(name string) string {
		return string(must.Value(os.ReadFile(filepath.Join(dir, name))))
	}
	assert.Equal(t, "2026c\n", read(VersionFile))
	want := "" +
		"Link\tEurope/Kyiv\tEurope/Kiev\n" +
		"Link\tEurope/Kyiv\tEurope/Zaporozhye\n"
	assert.Equal(t, want, read(LinksFile))
	assert.Equal(t, "PL\t+5215+02100\tEurope/Warsaw\n", read(ZoneTabFile))
}

func Test_Fetch_errors_tabular(t *testing.T) {
	valid := map[string]string{"version": "2026c\n", "zone.tab": ""}

	tt := []struct {
		testN string

		status int
		body   []byte
		want   string
	}{
		{
			"error - not found",
			http.StatusNotFound,
			nil,
			"/2026c: 404 Not Found",
		},
		{
			"error - not gzip",
			http.StatusOK,
			[]byte("abc"),
			"/2026c: unexpected EOF",
		},
		{
			"error - other version",
			http.StatusOK,
			tarGz(t, map[string]string{"version": "2026b\n"}),
			`holds version "2026b", want "2026c"`,
		},
		{
			"error - no zone.tab",
			http.StatusOK,
			tarGz(t, map[string]string{"version": "2026c\n"}),
			"/2026c has no zone.tab",
		},
		{
			"error - truncated archive",
			http.StatusOK,
			tarGz(t, valid)[:40],
			"/2026c: unexpected EOF",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			srv := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.status)
					_, _ = w.Write(tc.body)
				},
			))
			t.Cleanup(srv.Close)
			dir := t.TempDir()
			url := srv.URL + "/%s"

			// --- When ---
			err := Fetch(t.Context(), srv.Client(), url, "2026c", dir)

			// --- Then ---
			assert.ErrorContain(t, tc.want, err)
			assert.NoFileExist(t, filepath.Join(dir, VersionFile))
		})
	}
}

func Test_Fetch_requestErrors_tabular(t *testing.T) {
	tt := []struct {
		testN string

		ctx  context.Context
		url  string
		want string
	}{
		{
			"error - invalid URL",
			context.Background(),
			"://%s",
			"missing protocol scheme",
		},
		{
			"error - canceled context",
			canceledContext(),
			"http://127.0.0.1:1/%s",
			"context canceled",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := Fetch(tc.ctx, http.DefaultClient, tc.url, "2026c", "")

			// --- Then ---
			assert.ErrorContain(t, tc.want, err)
		})
	}
}

func Test_isDataFile_tabular(t *testing.T) {
	tt := []struct {
		testN string

		name string
		want bool
	}{
		{"region file", "europe", true},
		{"backward links", "backward", true},
		{"backzone", "backzone", false},
		{"compiled data", "tzdata.zi", false},
		{"table", "zone.tab", false},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := isDataFile(tc.name)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_linkLines(t *testing.T) {
	// --- Given ---
	data := []byte("" +
		"# Link\tA\tB\n" +
		"Zone\tEurope/Kyiv\t2:00\tEU\tEE%sT\n" +
		"Link  Europe/Kyiv\t\tEurope/Kiev  # renamed\n" +
		"Link\tbroken\n",
	)

	// --- When ---
	have := linkLines(data)

	// --- Then ---
	assert.Equal(t, []string{"Link\tEurope/Kyiv\tEurope/Kiev"}, have)
}

func Test_lastField(t *testing.T) {
	// --- When ---
	have := lastField("Link\tEurope/Kyiv\tEurope/Kiev")

	// --- Then ---
	assert.Equal(t, "Europe/Kiev", have)
}

func Test_ReadUpstream(t *testing.T) {
	// --- When ---
	have, err := ReadUpstream("testdata/tzdata", "testdata/zoneinfo.zip", 2026)

	// --- Then ---
	assert.NoError(t, err)
	want := []*xcal.Zone{
		{Name: "Etc/UTC", Status: xcal.Canonical},
		{
			Name:      "Europe/Kiev",
			Status:    xcal.Alias,
			Offset:    120,
			OffsetDST: 180,
			Link:      "Europe/Kyiv",
		},
		{
			Name:        "Europe/Kyiv",
			Status:      xcal.Canonical,
			Offset:      120,
			OffsetDST:   180,
			Notes:       "most of Ukraine",
			CountryCode: "UA",
		},
		{
			Name:        "Europe/Warsaw",
			Status:      xcal.Canonical,
			Offset:      60,
			OffsetDST:   120,
			CountryCode: "PL",
		},
		{Name: "UTC", Status: xcal.Alias, Link: "Etc/UTC"},
	}
	assert.Equal(t, want, have)
}

func Test_ReadUpstream_errors_tabular(t *testing.T) {
	tt := []struct {
		testN string

		dir  string
		zip  string
		want string
	}{
		{
			"error - no links",
			"testdata/not_existing",
			"testdata/zoneinfo.zip",
			"not_existing/links: no such file or directory",
		},
		{
			"error - no zip",
			"testdata/tzdata",
			"testdata/not_existing.zip",
			"not_existing.zip: no such file or directory",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, err := ReadUpstream(tc.dir, tc.zip, 2026)

			// --- Then ---
			assert.ErrorContain(t, tc.want, err)
			assert.Nil(t, have)
		})
	}
}

func Test_ReadUpstream_noZoneTab(t *testing.T) {
	// --- Given ---
	dir := t.TempDir()
	links := must.Value(os.ReadFile("testdata/tzdata/links"))
	must.Nil(os.WriteFile(filepath.Join(dir, LinksFile), links, 0644))

	// --- When ---
	have, err := ReadUpstream(dir, "testdata/zoneinfo.zip", 2026)

	// --- Then ---
	assert.ErrorContain(t, "zone.tab: no such file or directory", err)
	assert.Nil(t, have)
}

func Test_ReadUpstream_invalidZone(t *testing.T) {
	// --- Given ---
	pth := writeZip(t, map[string][]byte{"Europe/Bad": []byte("abc")})

	// --- When ---
	have, err := ReadUpstream("testdata/tzdata", pth, 2026)

	// --- Then ---
	assert.ErrorContain(t, "zoneinfo.zip: malformed time zone information", err)
	assert.Nil(t, have)
}

func Test_loadZipLocation(t *testing.T) {
	// --- Given ---
	zr := must.Value(zip.OpenReader("testdata/zoneinfo.zip"))
	t.Cleanup(func() { _ = zr.Close() })

	// --- When ---
	have, err := loadZipLocation(zr.File[3])

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, "Europe/Warsaw", have.String())
}

func Test_readLinks(t *testing.T) {
	// --- When ---
	have, err := readLinks("testdata/tzdata/links")

	// --- Then ---
	assert.NoError(t, err)
	want := map[string]string{"UTC": "Etc/UTC", "Europe/Kiev": "Europe/Kyiv"}
	assert.Equal(t, want, have)
}

func Test_readZoneTab(t *testing.T) {
	// --- When ---
	have, err := readZoneTab("testdata/tzdata/zone.tab")

	// --- Then ---
	assert.NoError(t, err)
	want := map[string]zoneTab{
		"Europe/Warsaw": {country: "PL"},
		"Europe/Kyiv":   {country: "UA", comment: "most of Ukraine"},
	}
	assert.Equal(t, want, have)
}

func Test_readZoneTab_invalidLine(t *testing.T) {
	// --- Given ---
	pth := filepath.Join(t.TempDir(), "zone.tab")
	must.Nil(os.WriteFile(pth, []byte("PL\tEurope/Warsaw\n"), 0644))

	// --- When ---
	have, err := readZoneTab(pth)

	// --- Then ---
	assert.ErrorContain(t, `invalid line "PL\tEurope/Warsaw"`, err)
	assert.Nil(t, have)
}

func Test_offsets_tabular(t *testing.T) {
	tt := []struct {
		testN string

		name string
		wStd int
		wDST int
	}{
		{"no DST", "Etc/UTC", 0, 0},
		{"DST", "Europe/Warsaw", 60, 120},
		{"DST east", "Europe/Kyiv", 120, 180},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			loc := must.Value(loadZipLocation(zipEntry(t, tc.name)))

			// --- When ---
			hStd, hDST := offsets(loc, 2026)

			// --- Then ---
			assert.Equal(t, tc.wStd, hStd)
			assert.Equal(t, tc.wDST, hDST)
		})
	}
}
