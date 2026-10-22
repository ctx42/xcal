//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/ctx42/xcal/internal/tzgen"
)

// main generates a Go source file with timezones defined in a JSON file.
func main() {
	src := "data/timezones.json"
	tzGo := "timezones.go"
	pkg := os.Getenv("GOPACKAGE")
	if pkg == "" {
		_, _ = fmt.Fprintln(os.Stderr, "GOPACKAGE not set: run via go generate")
		os.Exit(1)
	}
	if err := tzgen.GenerateTZ(src, tzGo, pkg); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
