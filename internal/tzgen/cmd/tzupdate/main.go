// Command tzupdate refreshes the xcal zone registry from the IANA tz database
// release the Go toolchain embeds. Run it from the module root, then run
// "go generate ./pkg/xcal" to regenerate the zone tables.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ctx42/xcal/internal/tzgen"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run updates the registry as the command-line args direct.
func run(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("tzupdate", flag.ContinueOnError)
	upd := &tzgen.Updater{Client: http.DefaultClient, URL: tzgen.ReleaseURL}
	fs.StringVar(&upd.DataDir, "data", "pkg/xcal/data", "registry directory")
	fs.StringVar(&upd.GoRoot, "goroot", "", "Go root (default: go env GOROOT)")
	fs.IntVar(&upd.Year, "year", time.Now().Year(), "year of zone offsets")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if upd.GoRoot == "" {
		out, err := exec.CommandContext(ctx, "go", "env", "GOROOT").Output()
		if err != nil {
			return fmt.Errorf("go env GOROOT: %w", err)
		}
		upd.GoRoot = strings.TrimSpace(string(out))
	}

	sum, err := upd.Run(ctx)
	if err != nil {
		return err
	}
	format := "" +
		"tz database %s: %d zones added, %d changed\n" +
		"run \"go generate ./pkg/xcal\" to regenerate the zone tables\n"
	_, _ = fmt.Printf(format, sum.Version, len(sum.Added), len(sum.Changed))
	return nil
}
