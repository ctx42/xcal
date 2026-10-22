package main

import (
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

	// --- When ---
	err := run(t.Context(), args)

	// --- Then ---
	assert.NoError(t, err)

	pth := filepath.Join(dir, tzgen.ZonesFile)
	zs := must.Value(tzgen.ReadZones(pth))
	assert.Len(t, 5, zs)
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
			err := run(t.Context(), tc.args)

			// --- Then ---
			assert.ErrorContain(t, tc.want, err)
		})
	}
}
