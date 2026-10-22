package xcal

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_ZoneMap(t *testing.T) {
	for _, z := range Zones {
		_, err := z.Loc()
		assert.NoError(t, err)
	}
}
