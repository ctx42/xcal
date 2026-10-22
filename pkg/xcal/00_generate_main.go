//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/ctx42/xcal/internal/tzgen"
)

// main generates go source file with timezones defined in a JSON file.
func main() {
	src := "data/timezones.json"
	tzGo := "timezones.go"
	pkg := os.Getenv("GOPACKAGE")
	if err := tzgen.GenerateTZ(src, tzGo, pkg); err != nil {
		_, _ = fmt.Fprint(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
}
