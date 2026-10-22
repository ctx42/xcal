package xcal

import (
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_DefaultLocations_tabular(t *testing.T) {
	tt := []struct {
		testN string

		want string
		loc  *time.Location
	}{
		{"WAW", "Europe/Warsaw", WAW},
		{"ZRH", "Europe/Zurich", ZRH},
		{"MOW", "Europe/Moscow", MOW},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := tc.loc.String()

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}
