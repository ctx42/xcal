// Command tzupdate refreshes the xcal zone registry from the IANA tz database
// release the Go toolchain embeds. Run it from the module root, then run
// "go generate ./pkg/xcal" to regenerate the zone tables.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ctx42/xcal/internal/tzgen"
)

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run updates the registry as the command-line args direct and reports the
// changes to out.
func run(ctx context.Context, args []string, out io.Writer) error {
	fs := flag.NewFlagSet("tzupdate", flag.ContinueOnError)
	upd := &tzgen.Updater{Client: http.DefaultClient, URL: tzgen.ReleaseURL}
	fs.StringVar(&upd.DataDir, "data", "pkg/xcal/data", "registry directory")
	fs.StringVar(&upd.GoRoot, "goroot", "", "Go root (default: go env GOROOT)")
	fs.IntVar(&upd.Year, "year", time.Now().Year(), "year of zone offsets")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if upd.GoRoot == "" {
		root, err := exec.CommandContext(ctx, "go", "env", "GOROOT").Output()
		if err != nil {
			return fmt.Errorf("go env GOROOT: %w", err)
		}
		upd.GoRoot = strings.TrimSpace(string(root))
	}

	sum, err := upd.Run(ctx)
	if err != nil {
		return err
	}
	added, changed := len(sum.Added), len(sum.Changed)
	format := "tz database %s: %d zones added, %d changed\n"
	_, _ = fmt.Fprintf(out, format, sum.Version, added, changed)
	for _, name := range sum.Added {
		_, _ = fmt.Fprintf(out, "  added:   %s\n", name)
	}
	for _, name := range sum.Changed {
		_, _ = fmt.Fprintf(out, "  changed: %s\n", name)
	}
	hint := "run \"go generate ./pkg/xcal\" to regenerate the zone tables\n"
	_, _ = fmt.Fprint(out, hint)
	return nil
}
