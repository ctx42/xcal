package tzgen

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/ctx42/xcal/pkg/xcal"
)

// Merge returns the zones of prev, a registry sorted by ID, updated with the
// upstream zones from [ReadUpstream], sorted by ID. A zone keeps its ID, and
// new zones get IDs above every ID in prev, in name order. A country code or
// note upstream lacks is kept from prev. A zone of prev missing upstream must
// already have the [xcal.Removed] status, and is kept as it is.
func Merge(prev, upstream []*xcal.Zone) ([]*xcal.Zone, error) {
	byName := make(map[string]*xcal.Zone, len(prev))
	next := 0
	for _, zon := range prev {
		byName[zon.Name] = zon
		next = max(next, zon.ID+1)
	}

	byNameAsc := func(a, b *xcal.Zone) int {
		return strings.Compare(a.Name, b.Name)
	}
	ups := slices.SortedFunc(slices.Values(upstream), byNameAsc)
	seen := make(map[string]bool, len(ups))
	zs := make([]*xcal.Zone, 0, len(ups))
	for _, up := range ups {
		zon := *up
		if old, ok := byName[zon.Name]; ok {
			zon.ID = old.ID
			zon.CountryCode = cmp.Or(zon.CountryCode, old.CountryCode)
			zon.Notes = cmp.Or(zon.Notes, old.Notes)
		} else {
			zon.ID = next
			next++
		}
		seen[zon.Name] = true
		zs = append(zs, &zon)
	}

	var missing []string
	for _, old := range prev {
		if seen[old.Name] {
			continue
		}
		if old.Status != xcal.Removed {
			missing = append(missing, old.Name)
			continue
		}
		zon := *old
		zs = append(zs, &zon)
	}
	if len(missing) > 0 {
		format := "" +
			"zones missing upstream, mark them removed (%q) with a link to " +
			"their replacement: %s"
		names := strings.Join(missing, ", ")
		return nil, fmt.Errorf(format, xcal.Removed, names)
	}

	slices.SortFunc(zs, func(a, b *xcal.Zone) int {
		return cmp.Compare(a.ID, b.ID)
	})
	if err := CheckStable(byName, zs); err != nil {
		return nil, err
	}
	for i, zon := range zs {
		if err := zon.Check(); err != nil {
			return nil, fmt.Errorf("at index %d: %w", i, err)
		}
	}
	return zs, nil
}

// WriteZones writes zs to a timezones JSON file at pth.
func WriteZones(pth string, zs []*xcal.Zone) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", " ")
	if err := enc.Encode(zs); err != nil {
		return err
	}
	return os.WriteFile(pth, buf.Bytes(), 0o600)
}
