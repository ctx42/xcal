package tzgen

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"

	"github.com/ctx42/xcal/pkg/xcal"
)

func Test_Updater_Run(t *testing.T) {
	// --- Given ---
	var reqs atomic.Int32
	srv := newRelease(t, &reqs)
	upd := &Updater{
		Client:  srv.Client(),
		URL:     srv.URL + "/%s",
		GoRoot:  newGoRoot(t, "2026c"),
		DataDir: newDataDir(t),
		Year:    2026,
	}

	// --- When ---
	have, err := upd.Run(t.Context())

	// --- Then ---
	assert.NoError(t, err)
	want := Summary{
		Version: "2026c",
		Added:   []string{"Etc/UTC", "Europe/Kiev", "Europe/Kyiv"},
		Changed: []string{"Europe/Warsaw"},
	}
	assert.Equal(t, want, have)
	assert.Equal(t, int32(1), reqs.Load())

	zs := must.Value(ReadZones(filepath.Join(upd.DataDir, ZonesFile)))
	assert.Len(t, 5, zs)
	assert.Equal(t, 418, zs[3].Id)
	assert.Equal(t, "Europe/Kiev", zs[3].Name)

	ver := filepath.Join(upd.DataDir, TZDataDir, VersionFile)
	assert.Equal(t, "2026c\n", string(must.Value(os.ReadFile(ver))))
}

func Test_Updater_Run_current(t *testing.T) {
	// --- Given ---
	var reqs atomic.Int32
	srv := newRelease(t, &reqs)
	upd := &Updater{
		Client:  srv.Client(),
		URL:     srv.URL + "/%s",
		GoRoot:  newGoRoot(t, "2026c"),
		DataDir: newDataDir(t),
		Year:    2026,
	}
	must.Value(upd.Run(t.Context()))

	// --- When ---
	have, err := upd.Run(t.Context())

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, Summary{Version: "2026c"}, have)
	assert.Equal(t, int32(1), reqs.Load())
}

func Test_Updater_Run_zeroYear(t *testing.T) {
	// --- Given ---
	dir := newDataDir(t)
	for _, name := range []string{VersionFile, LinksFile, ZoneTabFile} {
		src := filepath.Join("testdata", "tzdata", name)
		dst := filepath.Join(dir, TZDataDir, name)
		must.Nil(os.MkdirAll(filepath.Dir(dst), 0755))
		must.Nil(os.WriteFile(dst, must.Value(os.ReadFile(src)), 0644))
	}
	upd := &Updater{GoRoot: newGoRoot(t, "2026c"), DataDir: dir}

	// --- When ---
	_, err := upd.Run(t.Context())

	// --- Then ---
	assert.NoError(t, err)

	zs := must.Value(ReadZones(filepath.Join(dir, ZonesFile)))
	assert.Equal(t, "Europe/Warsaw", zs[1].Name)
	assert.Equal(t, 60, zs[1].Offset)
}

func Test_Updater_Run_noGoRoot(t *testing.T) {
	// --- Given ---
	upd := &Updater{GoRoot: t.TempDir(), DataDir: newDataDir(t)}

	// --- When ---
	have, err := upd.Run(t.Context())

	// --- Then ---
	assert.ErrorContain(t, "update.bash: no such file or directory", err)
	assert.Equal(t, Summary{}, have)
}

func Test_Updater_Run_fetchError(t *testing.T) {
	// --- Given ---
	srv := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(srv.Close)
	upd := &Updater{
		Client:  srv.Client(),
		URL:     srv.URL + "/%s",
		GoRoot:  newGoRoot(t, "2026c"),
		DataDir: newDataDir(t),
	}

	// --- When ---
	have, err := upd.Run(t.Context())

	// --- Then ---
	assert.ErrorContain(t, "/2026c: 404 Not Found", err)
	assert.Equal(t, Summary{}, have)
}

func Test_Updater_Run_noZoneInfo(t *testing.T) {
	// --- Given ---
	var reqs atomic.Int32
	srv := newRelease(t, &reqs)
	root := newGoRoot(t, "2026c")
	must.Nil(os.Remove(filepath.Join(root, "lib", "time", "zoneinfo.zip")))
	upd := &Updater{
		Client:  srv.Client(),
		URL:     srv.URL + "/%s",
		GoRoot:  root,
		DataDir: newDataDir(t),
	}

	// --- When ---
	have, err := upd.Run(t.Context())

	// --- Then ---
	assert.ErrorContain(t, "zoneinfo.zip: no such file or directory", err)
	assert.Equal(t, Summary{}, have)
}

func Test_Updater_Run_noRegistry(t *testing.T) {
	// --- Given ---
	var reqs atomic.Int32
	srv := newRelease(t, &reqs)
	upd := &Updater{
		Client:  srv.Client(),
		URL:     srv.URL + "/%s",
		GoRoot:  newGoRoot(t, "2026c"),
		DataDir: t.TempDir(),
	}

	// --- When ---
	have, err := upd.Run(t.Context())

	// --- Then ---
	assert.ErrorContain(t, "timezones.json: no such file or directory", err)
	assert.Equal(t, Summary{}, have)
}

func Test_Updater_Run_mergeError(t *testing.T) {
	// --- Given ---
	var reqs atomic.Int32
	srv := newRelease(t, &reqs)
	dir := t.TempDir()
	zs := []*xcal.Zone{{Id: 0, Name: "Europe/Paris", Status: xcal.Canonical}}
	must.Nil(WriteZones(filepath.Join(dir, ZonesFile), zs))
	upd := &Updater{
		Client:  srv.Client(),
		URL:     srv.URL + "/%s",
		GoRoot:  newGoRoot(t, "2026c"),
		DataDir: dir,
	}

	// --- When ---
	have, err := upd.Run(t.Context())

	// --- Then ---
	assert.ErrorContain(t, "zones missing upstream", err)
	assert.Equal(t, Summary{}, have)
}

func Test_Updater_Run_missingReleaseFile(t *testing.T) {
	// --- Given ---
	var reqs atomic.Int32
	srv := newRelease(t, &reqs)
	upd := &Updater{
		Client:  srv.Client(),
		URL:     srv.URL + "/%s",
		GoRoot:  newGoRoot(t, "2026c"),
		DataDir: newDataDir(t),
		Year:    2026,
	}
	must.Value(upd.Run(t.Context()))
	must.Nil(os.Remove(filepath.Join(upd.DataDir, TZDataDir, LinksFile)))

	// --- When ---
	_, err := upd.Run(t.Context())

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, int32(2), reqs.Load())
}

func Test_haveRelease_tabular(t *testing.T) {
	all := map[string]string{
		VersionFile: "2026c\n",
		LinksFile:   "",
		ZoneTabFile: "",
	}

	tt := []struct {
		testN string

		files map[string]string
		want  bool
	}{
		{"complete", all, true},
		{"no files", nil, false},
		{
			"other version",
			map[string]string{
				VersionFile: "2026b\n",
				LinksFile:   "",
				ZoneTabFile: "",
			},
			false,
		},
		{
			"missing links",
			map[string]string{VersionFile: "2026c\n", ZoneTabFile: ""},
			false,
		},
		{
			"missing zone.tab",
			map[string]string{VersionFile: "2026c\n", LinksFile: ""},
			false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			dir := t.TempDir()
			for name, data := range tc.files {
				pth := filepath.Join(dir, name)
				must.Nil(os.WriteFile(pth, []byte(data), 0644))
			}

			// --- When ---
			have := haveRelease(dir, "2026c")

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_summarize(t *testing.T) {
	// --- Given ---
	prev := []*xcal.Zone{
		{Id: 0, Name: "UTC"},
		{Id: 1, Name: "Europe/Warsaw", Offset: 120},
	}
	zs := []*xcal.Zone{
		{Id: 0, Name: "UTC"},
		{Id: 1, Name: "Europe/Warsaw", Offset: 60},
		{Id: 2, Name: "Europe/Kyiv"},
	}

	// --- When ---
	have := summarize("2026c", prev, zs)

	// --- Then ---
	want := Summary{
		Version: "2026c",
		Added:   []string{"Europe/Kyiv"},
		Changed: []string{"Europe/Warsaw"},
	}
	assert.Equal(t, want, have)
}
