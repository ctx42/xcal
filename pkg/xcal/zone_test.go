package xcal

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_Zone_Loc(t *testing.T) {
	// --- When ---
	loc, err := Zones["Europe/Warsaw"].Loc()

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, "Europe/Warsaw", loc.String())
}

func Test_Zone_Loc_unknownTimezone(t *testing.T) {
	// --- Given ---
	z := &Zone{
		Name: "Unknown/Unknown",
	}

	// --- When ---
	loc, err := z.Loc()

	// --- Then ---
	assert.ErrorIs(t, ErrInvTimezone, err)
	assert.Nil(t, loc)
}

func Test_Zone_Check(t *testing.T) {
	validZone := func() *Zone {
		return &Zone{
			Id:          416,
			CountryCode: "PL",
			Name:        "Europe/Warsaw",
			Offset:      120,
			OffsetDST:   180,
			Status:      "C",
			Link:        "",
			Notes:       "",
		}
	}

	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		z := validZone()

		// --- When ---
		err := z.Check()

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("negative idx", func(t *testing.T) {
		// --- Given ---
		z := validZone()
		z.Id = -1

		// --- When ---
		err := z.Check()

		// --- Then ---
		assert.ErrorEqual(t, "invalid ZoneID -1", err)
	})

	t.Run("unknown timezone name", func(t *testing.T) {
		// --- Given ---
		z := validZone()
		z.Name = "Unknown/Unknown"

		// --- When ---
		err := z.Check()

		// --- Then ---
		assert.ErrorIs(t, ErrInvTimezone, err)
		assert.ErrorEqual(t, "invalid timezone: Unknown/Unknown", err)
	})

	t.Run("invalid offset", func(t *testing.T) {
		// --- Given ---
		z := validZone()
		z.Offset = -1000

		// --- When ---
		err := z.Check()

		// --- Then ---
		assert.ErrorEqual(t, "invalid Zone offset -1000", err)
	})

	t.Run("invalid DST offset", func(t *testing.T) {
		// --- Given ---
		z := validZone()
		z.OffsetDST = -1000

		// --- When ---
		err := z.Check()

		// --- Then ---
		assert.ErrorEqual(t, "invalid Zone DST offset -1000", err)
	})

	t.Run("alias without link", func(t *testing.T) {
		// --- Given ---
		z := validZone()
		z.Status = Alias

		// --- When ---
		err := z.Check()

		// --- Then ---
		assert.ErrorEqual(t, "alias without link for Europe/Warsaw", err)
	})

	t.Run("alias with invalid link", func(t *testing.T) {
		// --- Given ---
		z := validZone()
		z.Status = Alias
		z.Link = "Unknown/Unknown"

		// --- When ---
		err := z.Check()

		// --- Then ---
		assert.ErrorEqual(t, "invalid alias link Unknown/Unknown", err)
	})
}

func Test_Zone_Validate(t *testing.T) {
	validZone := func() *Zone {
		return &Zone{
			Id:          416,
			CountryCode: "PL",
			Name:        "Europe/Warsaw",
			Offset:      120,
			OffsetDST:   180,
			Status:      "C",
			Link:        "",
			Notes:       "",
		}
	}

	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		z := validZone()

		// --- When ---
		err := z.Validate()

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("negative idx", func(t *testing.T) {
		// --- Given ---
		z := validZone()
		z.Id = -1

		// --- When ---
		err := z.Validate()

		// --- Then ---
		assert.ErrorEqual(t, "invalid ZoneID -1", err)
	})

	t.Run("not in zone table", func(t *testing.T) {
		// --- Given ---
		z := validZone()
		z.Name = "Europe/Kyiv"
		z.CountryCode = "UA"

		// --- When ---
		err := z.Validate()

		// --- Then ---
		assert.ErrorIs(t, ErrInvTimezone, err)
	})
}
