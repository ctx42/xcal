package xcal

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

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
		zon := validZone()

		// --- When ---
		err := zon.Check()

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - negative ID", func(t *testing.T) {
		// --- Given ---
		zon := validZone()
		zon.Id = -1

		// --- When ---
		err := zon.Check()

		// --- Then ---
		assert.ErrorEqual(t, "invalid ZoneID -1", err)
	})

	t.Run("error - unknown timezone name", func(t *testing.T) {
		// --- Given ---
		zon := validZone()
		zon.Name = "Unknown/Unknown"

		// --- When ---
		err := zon.Check()

		// --- Then ---
		assert.ErrorIs(t, ErrInvTimezone, err)
		assert.ErrorEqual(t, "invalid timezone: Unknown/Unknown", err)
	})

	t.Run("error - invalid offset", func(t *testing.T) {
		// --- Given ---
		zon := validZone()
		zon.Offset = -1000

		// --- When ---
		err := zon.Check()

		// --- Then ---
		assert.ErrorEqual(t, "invalid Zone offset -1000", err)
	})

	t.Run("error - invalid DST offset", func(t *testing.T) {
		// --- Given ---
		zon := validZone()
		zon.OffsetDST = -1000

		// --- When ---
		err := zon.Check()

		// --- Then ---
		assert.ErrorEqual(t, "invalid Zone DST offset -1000", err)
	})

	t.Run("error - alias without link", func(t *testing.T) {
		// --- Given ---
		zon := validZone()
		zon.Status = Alias

		// --- When ---
		err := zon.Check()

		// --- Then ---
		assert.ErrorEqual(t, "alias without link for Europe/Warsaw", err)
	})

	t.Run("error - alias with invalid link", func(t *testing.T) {
		// --- Given ---
		zon := validZone()
		zon.Status = Alias
		zon.Link = "Unknown/Unknown"

		// --- When ---
		err := zon.Check()

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
		zon := validZone()

		// --- When ---
		err := zon.Validate()

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - negative ID", func(t *testing.T) {
		// --- Given ---
		zon := validZone()
		zon.Id = -1

		// --- When ---
		err := zon.Validate()

		// --- Then ---
		assert.ErrorEqual(t, "invalid ZoneID -1", err)
	})

	t.Run("error - not in zone table", func(t *testing.T) {
		// --- Given ---
		zon := validZone()
		save := Zones[zon.Name]
		delete(Zones, zon.Name)
		t.Cleanup(func() { Zones[save.Name] = save })

		// --- When ---
		err := zon.Validate()

		// --- Then ---
		assert.ErrorIs(t, ErrInvTimezone, err)
	})
}

func Test_Zone_Loc(t *testing.T) {
	// --- When ---
	have, err := Zones["Europe/Warsaw"].Loc()

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, "Europe/Warsaw", have.String())
}

func Test_Zone_Loc_unknownTimezone(t *testing.T) {
	// --- Given ---
	zon := &Zone{
		Name: "Unknown/Unknown",
	}

	// --- When ---
	have, err := zon.Loc()

	// --- Then ---
	assert.ErrorIs(t, ErrInvTimezone, err)
	assert.Nil(t, have)
}
