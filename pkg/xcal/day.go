package xcal

import (
	"time"
)

// dayLayout is the wire format of a Day.
const dayLayout = "2006-01-02"

// Day is a calendar date with no time of day and no zone. It binds from the
// "YYYY-MM-DD" wire format and resolves to an instant only when a location is
// supplied, so the same request day can start at different instants in
// different time zones.
type Day struct {
	year  int
	month time.Month
	day   int
}

// NewDay returns the day of the given calendar date. Out-of-range values are
// normalized as [time.Date] does, so February 30 becomes March 2 or 1.
func NewDay(year int, month time.Month, day int) Day {
	return DayOf(time.Date(year, month, day, 0, 0, 0, 0, time.UTC))
}

// DayOf returns the calendar day the instant falls on in its own location.
func DayOf(tim time.Time) Day {
	year, month, day := tim.Date()
	return Day{year: year, month: month, day: day}
}

func parseDay(s string) (Day, error) {
	tim, err := time.Parse(dayLayout, s)
	if err != nil {
		return Day{}, err
	}
	return DayOf(tim), nil
}

// UnmarshalText implements encoding.TextUnmarshaler so a Day can be bound
// from a query parameter. Empty text sets the zero Day.
func (day *Day) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*day = Day{}
		return nil
	}
	parsed, err := parseDay(string(text))
	if err != nil {
		return err
	}
	*day = parsed
	return nil
}

// MarshalText implements encoding.TextMarshaler, so a Day encodes in the
// same "YYYY-MM-DD" format UnmarshalText accepts. The zero Day encodes as
// empty text.
func (day Day) MarshalText() ([]byte, error) {
	return []byte(day.String()), nil
}

// IsZero reports whether the day is unset.
func (day Day) IsZero() bool {
	return day == Day{}
}

// In returns the instant the day starts at in the location.
func (day Day) In(loc *time.Location) time.Time {
	return time.Date(day.year, day.month, day.day, 0, 0, 0, 0, loc)
}

// AddDays returns the day n calendar days later (or earlier for negative n).
func (day Day) AddDays(n int) Day {
	return DayOf(day.In(time.UTC).AddDate(0, 0, n))
}

// Before reports whether the day is before other.
func (day Day) Before(other Day) bool {
	return day.In(time.UTC).Before(other.In(time.UTC))
}

// DaysUntil returns the number of calendar days from the day to other - zero
// for the same day, negative when other is earlier.
func (day Day) DaysUntil(other Day) int {
	diff := other.In(time.UTC).Sub(day.In(time.UTC))
	return int(diff / (24 * time.Hour))
}

// String returns the day in "YYYY-MM-DD" format, or an empty string for the
// zero Day.
func (day Day) String() string {
	if day.IsZero() {
		return ""
	}
	return day.In(time.UTC).Format(dayLayout)
}
