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

// NewDay returns the day of the given calendar date.
func NewDay(year int, month time.Month, day int) Day {
	return Day{year: year, month: month, day: day}
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
// from a query parameter.
func (d *Day) UnmarshalText(text []byte) error {
	day, err := parseDay(string(text))
	if err != nil {
		return err
	}
	*d = day
	return nil
}

// MarshalText implements encoding.TextMarshaler, so a Day encodes in the
// same "YYYY-MM-DD" format UnmarshalText accepts.
func (d Day) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// IsZero reports whether the day is unset.
func (d Day) IsZero() bool {
	return d == Day{}
}

// In returns the instant the day starts at in the location.
func (d Day) In(loc *time.Location) time.Time {
	return time.Date(d.year, d.month, d.day, 0, 0, 0, 0, loc)
}

// AddDays returns the day n calendar days later (or earlier for negative n).
func (d Day) AddDays(n int) Day {
	return DayOf(d.In(time.UTC).AddDate(0, 0, n))
}

// Before reports whether d is before other.
func (d Day) Before(other Day) bool {
	return d.In(time.UTC).Before(other.In(time.UTC))
}

// DaysUntil returns the number of calendar days from d to other - zero for
// the same day, negative when other is earlier.
func (d Day) DaysUntil(other Day) int {
	diff := other.In(time.UTC).Sub(d.In(time.UTC))
	return int(diff / (24 * time.Hour))
}

func (d Day) String() string {
	return d.In(time.UTC).Format(dayLayout)
}
