package xcal

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_ZoneMap(t *testing.T) {
	for _, zon := range Zones {
		_, err := zon.Loc()
		assert.NoError(t, err)
	}
}
