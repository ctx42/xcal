package xcal

import (
	"sort"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
)

func Test_Midnight(t *testing.T) {
	tt := []struct {
		testN string

		date time.Time
		want time.Time
	}{
		{
			"1",
			time.Date(2019, 5, 28, 22, 0, 0, 0, time.UTC),
			time.Date(2019, 5, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			"2",
			time.Date(2019, 5, 28, 22, 0, 0, 0, WAW),
			time.Date(2019, 5, 28, 0, 0, 0, 0, WAW),
		},
		{
			"3",
			time.Date(2019, 5, 28, 22, 0, 0, 0, time.FixedZone("", -60*60)),
			time.Date(2019, 5, 28, 0, 0, 0, 0, time.FixedZone("", -60*60)),
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := Midnight(tc.date)

			// --- Then ---
			assert.Exact(t, tc.want, have)
		})
	}
}

func Test_MonthStartEnd(t *testing.T) {
	tt := []struct {
		testN string

		date   time.Time
		wStart time.Time
		wEnd   time.Time
	}{
		{
			"1",
			time.Date(2019, 5, 28, 22, 0, 0, 0, time.UTC),
			time.Date(2019, 5, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2019, 6, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			"2",
			time.Date(2004, 2, 29, 22, 0, 0, 0, time.UTC),
			time.Date(2004, 2, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2004, 3, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			"3",
			time.Date(2004, 12, 31, 22, 0, 0, 0, WAW),
			time.Date(2004, 12, 1, 0, 0, 0, 0, WAW),
			time.Date(2005, 1, 1, 0, 0, 0, 0, WAW),
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			hStart, hEnd := MonthStartEnd(tc.date)

			// --- Then ---
			assert.Equal(t, tc.wStart, hStart)
			assert.Equal(t, tc.wEnd, hEnd)
		})
	}
}

func Test_SinceMidnight(t *testing.T) {
	tt := []struct {
		testN string

		date time.Time
		want string
	}{
		{"1", time.Date(2019, 5, 28, 22, 0, 0, 0, time.UTC), "22h0m0s"},
		{"2", time.Date(2019, 5, 28, 22, 0, 0, 0, WAW), "22h0m0s"},
		{
			"3",
			time.Date(2019, 5, 28, 22, 0, 0, 0, time.FixedZone("", -60*60)),
			"22h0m0s",
		},
		{
			"4",
			time.Date(2019, 5, 28, 22, 0, 13, 0, time.FixedZone("", -60*60)),
			"22h0m13s",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := SinceMidnight(tc.date).String()

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_InWindow(t *testing.T) {
	tt := []struct {
		testN string

		want   bool
		start  string
		length string
		date   time.Time
	}{
		{
			"1",
			false,
			"2h0m0s",
			"2h0m0s",
			time.Date(2019, 5, 28, 1, 59, 59, 0, WAW),
		},
		{
			"2",
			true,
			"2h0m0s",
			"2h0m0s",
			time.Date(2019, 5, 28, 2, 0, 0, 0, WAW),
		},
		{
			"3",
			true,
			"2h0m0s",
			"2h0m0s",
			time.Date(2019, 5, 28, 4, 0, 0, 0, WAW),
		},
		{
			"4",
			false,
			"2h0m0s",
			"2h0m0s",
			time.Date(2019, 5, 28, 4, 0, 1, 0, WAW),
		},

		{
			"5",
			false,
			"2h0m0s",
			"2h0m0s",
			time.Date(2019, 2, 28, 1, 59, 59, 0, ZRH),
		},
		{
			"6",
			true,
			"2h0m0s",
			"2h0m0s",
			time.Date(2019, 2, 28, 2, 0, 0, 0, ZRH),
		},
		{
			"7",
			true,
			"2h0m0s",
			"2h0m0s",
			time.Date(2019, 2, 28, 4, 0, 0, 0, ZRH),
		},
		{
			"8",
			false,
			"2h0m0s",
			"2h0m0s",
			time.Date(2019, 2, 28, 4, 0, 1, 0, ZRH),
		},

		{
			"9",
			false,
			"2h0m0s",
			"1h30m0s",
			time.Date(2019, 10, 27, 2, 40, 45, 0, WAW),
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			mmStart := must.Value(time.ParseDuration(tc.start))
			mmLength := must.Value(time.ParseDuration(tc.length))

			// --- When ---
			have := InWindow(mmStart, mmLength, tc.date)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_InBetween(t *testing.T) {
	tt := []struct {
		testN string

		want  bool
		start time.Time
		end   time.Time
		date  time.Time
	}{
		{
			"1",
			false,
			time.Date(2019, 5, 28, 1, 59, 59, 0, time.UTC),
			time.Date(2019, 5, 29, 1, 59, 59, 0, time.UTC),
			time.Date(2019, 5, 28, 1, 59, 59, 0, time.UTC),
		},
		{
			"2",
			false,
			time.Date(2019, 5, 28, 1, 59, 59, 0, time.UTC),
			time.Date(2019, 5, 29, 1, 59, 59, 0, time.UTC),
			time.Date(2019, 5, 29, 1, 59, 59, 0, time.UTC),
		},
		{
			"3",
			true,
			time.Date(2019, 5, 28, 1, 59, 59, 0, time.UTC),
			time.Date(2019, 5, 29, 1, 59, 59, 0, time.UTC),
			time.Date(2019, 5, 28, 16, 59, 59, 0, time.UTC),
		},
		{
			"4",
			false,
			time.Date(2019, 5, 28, 1, 59, 59, 0, time.UTC),
			time.Date(2019, 5, 29, 1, 59, 59, 0, time.UTC),
			time.Date(2019, 5, 28, 1, 59, 58, 0, time.UTC),
		},
		{
			"5",
			false,
			time.Date(2019, 5, 28, 1, 59, 59, 0, time.UTC),
			time.Date(2019, 5, 29, 1, 59, 59, 0, time.UTC),
			time.Date(2019, 5, 29, 2, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := InBetween(tc.date, tc.start, tc.end)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_InBetweenInclusive(t *testing.T) {
	tt := []struct {
		testN string

		want  bool
		start time.Time
		end   time.Time
		date  time.Time
	}{
		{
			"1",
			true,
			time.Date(2019, 5, 28, 1, 59, 59, 0, time.UTC),
			time.Date(2019, 5, 29, 1, 59, 59, 0, time.UTC),
			time.Date(2019, 5, 28, 1, 59, 59, 0, time.UTC),
		},
		{
			"2",
			true,
			time.Date(2019, 5, 28, 1, 59, 59, 0, time.UTC),
			time.Date(2019, 5, 29, 1, 59, 59, 0, time.UTC),
			time.Date(2019, 5, 29, 1, 59, 59, 0, time.UTC),
		},
		{
			"3",
			true,
			time.Date(2019, 5, 28, 1, 59, 59, 0, time.UTC),
			time.Date(2019, 5, 29, 1, 59, 59, 0, time.UTC),
			time.Date(2019, 5, 28, 16, 59, 59, 0, time.UTC),
		},
		{
			"4",
			false,
			time.Date(2019, 5, 28, 1, 59, 59, 0, time.UTC),
			time.Date(2019, 5, 29, 1, 59, 59, 0, time.UTC),
			time.Date(2019, 5, 28, 1, 59, 58, 0, time.UTC),
		},
		{
			"5",
			false,
			time.Date(2019, 5, 28, 1, 59, 59, 0, time.UTC),
			time.Date(2019, 5, 29, 1, 59, 59, 0, time.UTC),
			time.Date(2019, 5, 29, 2, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := InBetweenInclusive(tc.date, tc.start, tc.end)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_IsBeforeOrEqual(t *testing.T) {
	tt := []struct {
		testN string

		want bool
		a, b time.Time
	}{
		{
			"1",
			true,
			time.Date(2019, 5, 28, 1, 59, 59, 0, time.UTC),
			time.Date(2019, 5, 28, 1, 59, 59, 0, time.UTC),
		},
		{
			"2",
			true,
			time.Date(2019, 5, 28, 1, 59, 58, 0, time.UTC),
			time.Date(2019, 5, 28, 1, 59, 59, 0, time.UTC),
		},
		{
			"3",
			false,
			time.Date(2019, 5, 28, 2, 0, 0, 0, time.UTC),
			time.Date(2019, 5, 28, 1, 59, 59, 0, time.UTC),
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := IsBeforeOrEqual(tc.a, tc.b)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_IsAfterOrEqual(t *testing.T) {
	tt := []struct {
		testN string

		want bool
		a, b time.Time
	}{
		{
			"1",
			true,
			time.Date(2019, 5, 28, 1, 59, 59, 0, time.UTC),
			time.Date(2019, 5, 28, 1, 59, 59, 0, time.UTC),
		},
		{
			"2",
			true,
			time.Date(2019, 5, 28, 1, 59, 59, 0, time.UTC),
			time.Date(2019, 5, 28, 1, 59, 58, 0, time.UTC),
		},
		{
			"3",
			false,
			time.Date(2019, 5, 28, 1, 59, 59, 0, time.UTC),
			time.Date(2019, 5, 28, 2, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := IsAfterOrEqual(tc.a, tc.b)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_TimeSortASC(t *testing.T) {
	// --- Given ---
	in := []time.Time{
		time.Date(2019, time.July, 18, 23, 17, 10, 0, time.UTC),
		time.Date(2019, time.July, 18, 23, 17, 10, 0, WAW),
	}

	want := []time.Time{
		time.Date(2019, time.July, 18, 23, 17, 10, 0, WAW),
		time.Date(2019, time.July, 18, 23, 17, 10, 0, time.UTC),
	}

	// --- When ---
	sort.Sort(TimeSortASC(in))

	// --- Then ---
	assert.Equal(t, want, in)
}

func Test_ChangeLocation(t *testing.T) {
	tt := []struct {
		testN string

		in   time.Time
		loc  string
		want string
	}{
		{
			"1",
			time.Date(2022, 1, 29, 8, 41, 12, 0, WAW),
			"America/Los_Angeles",
			"2022-01-29T08:41:12-08:00",
		},
		{
			"2",
			time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
			"America/Los_Angeles",
			"0000-12-31T16:07:02-07:52",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			t.Parallel()

			// --- Given ---
			tz := must.Value(time.LoadLocation(tc.loc))

			// --- When ---
			have := ChangeLocation(tc.in, tz)

			// --- Then ---
			assert.Equal(t, tc.want, have.Format(time.RFC3339))
			assert.Equal(t, tz.String(), have.Location().String())
		})
	}
}

func Benchmark_ChangeLocation(b *testing.B) {
	b.StopTimer()
	tim := time.Date(2022, 4, 23, 10, 18, 19, 0, WAW)
	var n time.Time

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		n = ChangeLocation(tim, time.UTC)
	}
	_ = n
}
