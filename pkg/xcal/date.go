package xcal

import (
	"time"
)

// Midnight returns time of midnight for given tim in its timezone.
func Midnight(tim time.Time) time.Time {
	year, month, day := tim.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, tim.Location())
}

// MonthStartEnd for any date returns the beginning of the month and the
// beginning of the next month for that date [m,m+1).
func MonthStartEnd(tim time.Time) (time.Time, time.Time) {
	year, month, _ := tim.Date()
	start := time.Date(year, month, 1, 0, 0, 0, 0, tim.Location())

	end := start.Add(35 * 24 * time.Hour)
	year, month, _ = end.Date()
	end = time.Date(year, month, 1, 0, 0, 0, 0, end.Location())
	return start, end
}

// SinceMidnight returns time elapsed since tim midnight.
func SinceMidnight(tim time.Time) time.Duration {
	return tim.Sub(Midnight(tim))
}

// InWindow returns true if tim is in time window defined by start and length.
// Start is defined as time elapsed since midnight of tim.
func InWindow(start, length time.Duration, tim time.Time) bool {
	midnight := Midnight(tim)
	begin := midnight.Add(start)
	end := begin.Add(length)
	return InBetweenInclusive(tim, begin, end)
}

// InBetween returns true if tim in between start and end not inclusive.
func InBetween(tim, start, end time.Time) bool {
	return tim.After(start) && tim.Before(end)
}

// InBetweenInclusive returns true if date in between start and end inclusive.
func InBetweenInclusive(tim, start, end time.Time) bool {
	return (tim.Equal(start) || tim.After(start)) &&
		(tim.Equal(end) || tim.Before(end))
}

// IsBeforeOrEqual returns true if 'a' is before or equal to 'b'.
func IsBeforeOrEqual(a, b time.Time) bool { return a.Before(b) || a.Equal(b) }

// IsAfterOrEqual returns true if 'a' is after or equal to 'b'.
func IsAfterOrEqual(a, b time.Time) bool { return a.After(b) || a.Equal(b) }

// TimeSortASC sorts dates in ascending order.
//
//	in := []time.Time{tim0, tim1, tim2}
//	sort.Sort(TimeSortASC(in))
type TimeSortASC []time.Time

func (tsa TimeSortASC) Len() int           { return len(tsa) }
func (tsa TimeSortASC) Less(i, j int) bool { return tsa[i].Before(tsa[j]) }
func (tsa TimeSortASC) Swap(i, j int)      { tsa[i], tsa[j] = tsa[j], tsa[i] }

// ChangeLocation changes the timezone for the date.
// It does not change the date string representation.
func ChangeLocation(date time.Time, tz *time.Location) time.Time {
	// Date has not been set.
	if date.IsZero() {
		return date.In(tz)
	}
	return time.Date(
		date.Year(),
		date.Month(),
		date.Day(),
		date.Hour(),
		date.Minute(),
		date.Second(),
		date.Nanosecond(),
		tz,
	)
}
