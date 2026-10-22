package xcal

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_Zone_Check(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		zon := validZone()

		// --- When ---
		err := zon.Check()

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("ok - alias with link", func(t *testing.T) {
		// --- Given ---
		zon := validZone()
		zon.Status = Alias
		zon.Link = "Europe/Berlin"

		// --- When ---
		err := zon.Check()

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("ok - deprecated", func(t *testing.T) {
		// --- Given ---
		zon := validZone()
		zon.Status = Deprecated

		// --- When ---
		err := zon.Check()

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("ok - removed with link", func(t *testing.T) {
		// --- Given ---
		zon := validZone()
		zon.Name = "Europe/Gone"
		zon.Status = Removed
		zon.Link = "Europe/Berlin"

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

	t.Run("error - empty name", func(t *testing.T) {
		// --- Given ---
		zon := validZone()
		zon.Name = ""

		// --- When ---
		err := zon.Check()

		// --- Then ---
		assert.ErrorEqual(t, "invalid timezone: ", err)
	})

	t.Run("error - Local name", func(t *testing.T) {
		// --- Given ---
		zon := validZone()
		zon.Name = "Local"

		// --- When ---
		err := zon.Check()

		// --- Then ---
		assert.ErrorEqual(t, "invalid timezone: Local", err)
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

	t.Run("error - invalid status", func(t *testing.T) {
		// --- Given ---
		zon := validZone()
		zon.Status = "X"

		// --- When ---
		err := zon.Check()

		// --- Then ---
		assert.ErrorEqual(t, `invalid Zone status "X"`, err)
	})

	t.Run("error - alias without link", func(t *testing.T) {
		// --- Given ---
		zon := validZone()
		zon.Status = Alias

		// --- When ---
		err := zon.Check()

		// --- Then ---
		want := `zone Europe/Warsaw with status "A" has no link`
		assert.ErrorEqual(t, want, err)
	})

	t.Run("error - removed with empty name", func(t *testing.T) {
		// --- Given ---
		zon := validZone()
		zon.Name = ""
		zon.Status = Removed
		zon.Link = "Etc/UTC"

		// --- When ---
		err := zon.Check()

		// --- Then ---
		assert.ErrorIs(t, ErrInvTimezone, err)
		assert.ErrorEqual(t, "invalid timezone: ", err)
	})

	t.Run("error - removed without link", func(t *testing.T) {
		// --- Given ---
		zon := validZone()
		zon.Status = Removed

		// --- When ---
		err := zon.Check()

		// --- Then ---
		want := `zone Europe/Warsaw with status "R" has no link`
		assert.ErrorEqual(t, want, err)
	})

	t.Run("error - alias with Local link", func(t *testing.T) {
		// --- Given ---
		zon := validZone()
		zon.Status = Alias
		zon.Link = "Local"

		// --- When ---
		err := zon.Check()

		// --- Then ---
		assert.ErrorEqual(t, "invalid link Local", err)
	})

	t.Run("error - alias with invalid link", func(t *testing.T) {
		// --- Given ---
		zon := validZone()
		zon.Status = Alias
		zon.Link = "Unknown/Unknown"

		// --- When ---
		err := zon.Check()

		// --- Then ---
		assert.ErrorEqual(t, "invalid link Unknown/Unknown", err)
	})
}

func Test_Zone_Validate(t *testing.T) {
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

func Test_Zone_Loc_removed(t *testing.T) {
	// --- Given ---
	zon := &Zone{Name: "Europe/Gone", Status: Removed, Link: "Europe/Berlin"}

	// --- When ---
	have, err := zon.Loc()

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, "Europe/Berlin", have.String())
}

func Test_Zone_Loc_rejectedNames_tabular(t *testing.T) {
	tt := []struct {
		testN string

		zon *Zone
	}{
		{"error - zero zone", &Zone{}},
		{"error - Local", &Zone{Name: "Local"}},
		{"error - removed without link", &Zone{Name: "X", Status: Removed}},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, err := tc.zon.Loc()

			// --- Then ---
			assert.ErrorIs(t, ErrInvTimezone, err)
			assert.Nil(t, have)
		})
	}
}

func Test_loadTZ(t *testing.T) {
	// --- When ---
	have, err := loadTZ("Europe/Warsaw")

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, "Europe/Warsaw", have.String())
}

func Test_loadTZ_rejected_tabular(t *testing.T) {
	tt := []struct {
		testN string

		name string
	}{
		{"error - empty", ""},
		{"error - Local", "Local"},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, err := loadTZ(tc.name)

			// --- Then ---
			assert.ErrorIs(t, ErrInvTimezone, err)
			assert.Nil(t, have)
		})
	}
}

func Test_loadTZ_unknown(t *testing.T) {
	// --- When ---
	have, err := loadTZ("Unknown/Unknown")

	// --- Then ---
	assert.ErrorContain(t, "unknown time zone Unknown/Unknown", err)
	assert.Nil(t, have)
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
