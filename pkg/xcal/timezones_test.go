package xcal

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_ZoneMap(t *testing.T) {
	for _, zon := range Zones {
		// --- When ---
		_, err := zon.Loc()

		// --- Then ---
		assert.NoError(t, err, zon.Name)
	}
}
