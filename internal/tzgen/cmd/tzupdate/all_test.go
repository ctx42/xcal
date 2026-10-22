package main

import (
	"os"
	"path/filepath"

	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testing/pkg/tester"
)

// testdata is the tzgen package testdata directory.
const testdata = "../../testdata"

// copyFile copies the file at src to dst.
func copyFile(t tester.T, src, dst string) {
	t.Helper()
	must.Nil(os.MkdirAll(filepath.Dir(dst), 0755))
	must.Nil(os.WriteFile(dst, must.Value(os.ReadFile(src)), 0644))
}

// newGoRoot creates a Go root embedding the 2026c tz database release with
// the zones of the tzgen testdata zoneinfo.zip, and returns its path.
func newGoRoot(t tester.T) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "lib", "time")
	must.Nil(os.MkdirAll(dir, 0755))
	script := []byte("DATA=2026c\n")
	must.Nil(os.WriteFile(filepath.Join(dir, "update.bash"), script, 0644))
	src := filepath.Join(testdata, "zoneinfo.zip")
	copyFile(t, src, filepath.Join(dir, "zoneinfo.zip"))
	return root
}
