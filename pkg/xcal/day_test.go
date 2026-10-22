package xcal

import (
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_parseDay(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- When ---
		have, err := parseDay("2026-03-25")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, NewDay(2026, time.March, 25), have)
	})

	t.Run("error - time component not allowed", func(t *testing.T) {
		// --- When ---
		_, err := parseDay("2026-03-25T00:00:00Z")

		// --- Then ---
		assert.Error(t, err)
	})

	t.Run("error - invalid date", func(t *testing.T) {
		// --- When ---
		_, err := parseDay("2026-02-30")

		// --- Then ---
		assert.Error(t, err)
	})
}

func Test_Day_UnmarshalText(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		var have Day

		// --- When ---
		err := have.UnmarshalText([]byte("2026-03-25"))

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, NewDay(2026, time.March, 25), have)
	})

	t.Run("error - leaves the value untouched", func(t *testing.T) {
		// --- Given ---
		have := NewDay(2026, time.March, 25)

		// --- When ---
		err := have.UnmarshalText([]byte("abc"))

		// --- Then ---
		assert.Error(t, err)
		assert.Equal(t, NewDay(2026, time.March, 25), have)
	})
}

func Test_Day_IsZero(t *testing.T) {
	assert.True(t, Day{}.IsZero())
	assert.False(t, NewDay(2026, time.March, 25).IsZero())
}

func Test_Day_In(t *testing.T) {
	t.Run("UTC", func(t *testing.T) {
		// --- When ---
		have := NewDay(2026, time.March, 25).In(time.UTC)

		// --- Then ---
		assert.Equal(t, "2026-03-25T00:00:00Z", have.Format(time.RFC3339))
	})

	t.Run("localized day starts at local midnight", func(t *testing.T) {
		// --- When ---
		have := NewDay(2026, time.March, 25).In(WAW)

		// --- Then ---
		assert.Equal(t, "2026-03-25T00:00:00+01:00", have.Format(time.RFC3339))
		assert.Equal(t, "2026-03-24T23:00:00Z", have.UTC().Format(time.RFC3339))
	})
}

func Test_DayOf(t *testing.T) {
	// --- Given ---
	tim := time.Date(2026, time.March, 24, 23, 30, 0, 0, time.UTC)

	// --- When ---
	utcDay := DayOf(tim)
	locDay := DayOf(tim.In(WAW))

	// --- Then ---
	assert.Equal(t, NewDay(2026, time.March, 24), utcDay)
	assert.Equal(t, NewDay(2026, time.March, 25), locDay)
}

func Test_Day_AddDays(t *testing.T) {
	// --- Given ---
	day := NewDay(2026, time.February, 27)

	// --- Then ---
	assert.Equal(t, NewDay(2026, time.March, 1), day.AddDays(2))
	assert.Equal(t, NewDay(2026, time.February, 26), day.AddDays(-1))
	assert.Equal(t, day, day.AddDays(0))
}

func Test_Day_Before(t *testing.T) {
	// --- Given ---
	early := NewDay(2026, time.March, 24)
	late := NewDay(2026, time.March, 25)

	// --- Then ---
	assert.True(t, early.Before(late))
	assert.False(t, late.Before(early))
	assert.False(t, early.Before(early))
}

func Test_Day_DaysUntil(t *testing.T) {
	// --- Given ---
	day := NewDay(2026, time.March, 25)

	// --- Then ---
	assert.Equal(t, 0, day.DaysUntil(day))
	assert.Equal(t, 31, day.DaysUntil(NewDay(2026, time.April, 25)))
	assert.Equal(t, -1, day.DaysUntil(NewDay(2026, time.March, 24)))
}

func Test_Day_String(t *testing.T) {
	assert.Equal(t, "2026-03-05", NewDay(2026, time.March, 5).String())
}
