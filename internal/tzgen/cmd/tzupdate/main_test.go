package main

import (
	"bytes"
	"io"
	"path/filepath"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"

	"github.com/ctx42/xcal/internal/tzgen"
)

func Test_run(t *testing.T) {
	// --- Given ---
	root := newGoRoot(t)
	dir := t.TempDir()
	for _, name := range []string{
		tzgen.VersionFile,
		tzgen.LinksFile,
		tzgen.ZoneTabFile,
	} {
		src := filepath.Join(testdata, "tzdata", name)
		copyFile(t, src, filepath.Join(dir, tzgen.TZDataDir, name))
	}
	src := filepath.Join(testdata, "timezones.json")
	copyFile(t, src, filepath.Join(dir, tzgen.ZonesFile))
	args := []string{"-goroot", root, "-data", dir, "-year", "2026"}
	var out bytes.Buffer

	// --- When ---
	err := run(t.Context(), args, &out)

	// --- Then ---
	assert.NoError(t, err)
	want := "" +
		"tz database 2026c: 3 zones added, 1 changed\n" +
		"  added:   Etc/UTC\n" +
		"  added:   Europe/Kiev\n" +
		"  added:   Europe/Kyiv\n" +
		"  changed: Europe/Warsaw\n" +
		"run \"go generate ./pkg/xcal\" to regenerate the zone tables\n"
	assert.Equal(t, want, out.String())

	pth := filepath.Join(dir, tzgen.ZonesFile)
	zs := must.Value(tzgen.ReadZones(pth))
	assert.Len(t, 5, zs)
}

func Test_run_checkCurrent(t *testing.T) {
	// --- Given ---
	root := newGoRoot(t)
	dir := t.TempDir()
	for _, name := range []string{
		tzgen.VersionFile,
		tzgen.LinksFile,
		tzgen.ZoneTabFile,
	} {
		src := filepath.Join(testdata, "tzdata", name)
		copyFile(t, src, filepath.Join(dir, tzgen.TZDataDir, name))
	}
	args := []string{"-check", "-goroot", root, "-data", dir}
	var out bytes.Buffer

	// --- When ---
	err := run(t.Context(), args, &out)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, "zone registry is current\n", out.String())
}

func Test_run_checkOutdated(t *testing.T) {
	// --- Given ---
	args := []string{"-check", "-goroot", newGoRoot(t), "-data", t.TempDir()}
	var out bytes.Buffer

	// --- When ---
	err := run(t.Context(), args, &out)

	// --- Then ---
	want := "" +
		"zone registry outdated: Go embeds tz database 2026c, the registry " +
		"has none; run \"go run ./internal/tzgen/cmd/tzupdate\" and " +
		"\"go generate ./pkg/xcal\""
	assert.ErrorEqual(t, want, err)
	assert.Equal(t, "", out.String())
}

func Test_run_errors_tabular(t *testing.T) {
	tt := []struct {
		testN string

		args []string
		want string
	}{
		{
			"error - unknown flag",
			[]string{"-bad"},
			"flag provided but not defined: -bad",
		},
		{
			"error - no Go root",
			[]string{"-goroot", "not_existing"},
			"update.bash: no such file or directory",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := run(t.Context(), tc.args, io.Discard)

			// --- Then ---
			assert.ErrorContain(t, tc.want, err)
		})
	}
}
