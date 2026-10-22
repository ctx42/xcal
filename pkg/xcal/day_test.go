package xcal

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
)

func Test_NewDay(t *testing.T) {
	// --- When ---
	have := NewDay(2026, time.February, 30)

	// --- Then ---
	assert.Equal(t, NewDay(2026, time.March, 2), have)
	assert.Equal(t, "2026-03-02", have.String())
}

func Test_DayOf_tabular(t *testing.T) {
	tim := time.Date(2026, time.March, 24, 23, 30, 0, 0, time.UTC)

	tt := []struct {
		testN string

		tim  time.Time
		want Day
	}{
		{"UTC", tim, NewDay(2026, time.March, 24)},
		{"Warsaw", tim.In(WAW), NewDay(2026, time.March, 25)},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := DayOf(tc.tim)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

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
		assert.ErrorContain(t, "extra text", err)
	})

	t.Run("error - invalid date", func(t *testing.T) {
		// --- When ---
		_, err := parseDay("2026-02-30")

		// --- Then ---
		assert.ErrorContain(t, "day out of range", err)
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

	t.Run("empty", func(t *testing.T) {
		// --- Given ---
		have := NewDay(2026, time.March, 25)

		// --- When ---
		err := have.UnmarshalText([]byte(""))

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, Day{}, have)
	})

	t.Run("error - leaves the value untouched", func(t *testing.T) {
		// --- Given ---
		have := NewDay(2026, time.March, 25)

		// --- When ---
		err := have.UnmarshalText([]byte("abc"))

		// --- Then ---
		assert.ErrorContain(t, `cannot parse "abc"`, err)
		assert.Equal(t, NewDay(2026, time.March, 25), have)
	})
}

func Test_Day_MarshalText(t *testing.T) {
	// --- Given ---
	ts := struct {
		Day Day `json:"day"`
	}{
		Day: NewDay(2026, time.March, 5),
	}

	// --- When ---
	have, err := json.Marshal(ts)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, `{"day":"2026-03-05"}`, string(have))
}

func Test_Day_MarshalText_zero(t *testing.T) {
	// --- Given ---
	ts := struct {
		Day Day `json:"day"`
	}{}

	// --- When ---
	have, err := json.Marshal(ts)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, `{"day":""}`, string(have))
}

func Test_Day_MarshalText_yearOutOfRange_tabular(t *testing.T) {
	tt := []struct {
		testN string

		day Day
	}{
		{"error - year after 9999", NewDay(10000, time.January, 1)},
		{"error - negative year", NewDay(-1, time.January, 1)},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, err := tc.day.MarshalText()

			// --- Then ---
			assert.ErrorContain(t, "year outside of range [0,9999]", err)
			assert.Nil(t, have)
		})
	}
}

func Test_Day_MarshalText_zeroRoundTrip(t *testing.T) {
	// --- Given ---
	data := must.Value(json.Marshal(struct{ Day Day }{}))
	have := struct{ Day Day }{Day: NewDay(2026, time.March, 25)}

	// --- When ---
	err := json.Unmarshal(data, &have)

	// --- Then ---
	assert.NoError(t, err)
	assert.True(t, have.Day.IsZero())
}

func Test_Day_IsZero_tabular(t *testing.T) {
	tt := []struct {
		testN string

		day  Day
		want bool
	}{
		{"zero", Day{}, true},
		{"set", NewDay(2026, time.March, 25), false},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := tc.day.IsZero()

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
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

func Test_Day_AddDays_tabular(t *testing.T) {
	tt := []struct {
		testN string

		n    int
		want Day
	}{
		{"across month end", 2, NewDay(2026, time.March, 1)},
		{"negative", -1, NewDay(2026, time.February, 26)},
		{"zero", 0, NewDay(2026, time.February, 27)},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			day := NewDay(2026, time.February, 27)

			// --- When ---
			have := day.AddDays(tc.n)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_Day_Before_tabular(t *testing.T) {
	early := NewDay(2026, time.March, 24)
	late := NewDay(2026, time.March, 25)

	tt := []struct {
		testN string

		day   Day
		other Day
		want  bool
	}{
		{"earlier", early, late, true},
		{"later", late, early, false},
		{"same", early, early, false},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := tc.day.Before(tc.other)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_Day_DaysUntil_tabular(t *testing.T) {
	tt := []struct {
		testN string

		other Day
		want  int
	}{
		{"same day", NewDay(2026, time.March, 25), 0},
		{"next month", NewDay(2026, time.April, 25), 31},
		{"earlier", NewDay(2026, time.March, 24), -1},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			day := NewDay(2026, time.March, 25)

			// --- When ---
			have := day.DaysUntil(tc.other)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_Day_String(t *testing.T) {
	// --- When ---
	have := NewDay(2026, time.March, 5).String()

	// --- Then ---
	assert.Equal(t, "2026-03-05", have)
}

func Test_Day_String_zero(t *testing.T) {
	// --- When ---
	have := Day{}.String()

	// --- Then ---
	assert.Equal(t, "", have)
}
