package xcal

import (
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
)

func Test_CheckTimeBounds_tabular(t *testing.T) {
	tt := []struct {
		testN string

		tim time.Time
		fip time.Time
		fif string
	}{
		{
			"within bounds",
			time.Now().UTC(),
			time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			"1s",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			fif := must.Value(time.ParseDuration(tc.fif))

			// --- When ---
			err := CheckTimeBounds("", tc.tim, tc.fip, fif)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_CheckTimeBounds_errors_tabular(t *testing.T) {
	tt := []struct {
		testN string

		tim   time.Time
		fip   time.Time
		fif   string
		field string
		want  string
	}{
		{
			"error - far future",
			time.Date(2200, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			"1s",
			"field",
			"date out of bounds: field date in far future",
		},
		{
			"error - far past",
			time.Date(1999, 12, 31, 23, 59, 59, 0, time.UTC),
			time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			"1s",
			"field",
			"date out of bounds: field date in far past",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			fif := must.Value(time.ParseDuration(tc.fif))

			// --- When ---
			err := CheckTimeBounds(tc.field, tc.tim, tc.fip, fif)

			// --- Then ---
			assert.ErrorIs(t, ErrTimeBounds, err)
			assert.ErrorEqual(t, tc.want, err)
		})
	}
}

func Test_ToZone(t *testing.T) {
	// --- When ---
	have, err := ToZone(WAW)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, 416, have.Id)
}

func Test_ToZone_Local(t *testing.T) {
	// --- Given ---
	orig := time.Local
	time.Local = time.FixedZone("Local", 120)
	t.Cleanup(func() { time.Local = orig })

	// --- When ---
	have, err := ToZone(time.Local)

	// --- Then ---
	assert.ErrorEqual(t, "timezone Local means nothing", err)
	assert.Nil(t, have)
}

func Test_ToZone_UnknownTimezone(t *testing.T) {
	// --- Given ---
	loc := time.FixedZone("MyTZ", 0)

	// --- When ---
	have, err := ToZone(loc)

	// --- Then ---
	assert.ErrorIs(t, ErrInvTimezone, err)
	assert.Nil(t, have)
}

func Test_ByZoneID(t *testing.T) {
	// --- When ---
	loc, err := ByZoneID(416)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, "Europe/Warsaw", loc.String())
}

func Test_ByZoneID_unknownID(t *testing.T) {
	// --- When ---
	loc, err := ByZoneID(1000)

	// --- Then ---
	assert.ErrorIs(t, ErrInvTimezone, err)
	assert.Nil(t, loc)
}

func Test_ByZoneID_unknownLocation(t *testing.T) {
	// --- Given ---
	ZoneIDs[1000] = &Zone{
		CountryCode: "",
		Id:          1000,
		Link:        "",
		Name:        "MyZone",
		Notes:       "",
		Offset:      0,
		OffsetDST:   0,
		Status:      "A",
	}
	defer func() { delete(ZoneIDs, 1000) }()

	// --- When ---
	loc, err := ByZoneID(1000)

	// --- Then ---
	assert.ErrorIs(t, ErrInvTimezone, err)
	assert.Nil(t, loc)
}
