package xcal

import (
	"fmt"
	"time"
)

// CheckTimeBounds returns an error wrapping [ErrTimeBounds] when tim is before
// fip or later than now plus fif. The name prefixes the error message.
func CheckTimeBounds(name string, tim, fip time.Time, fif time.Duration) error {
	if tim.After(time.Now().UTC().Add(fif)) {
		return fmt.Errorf("%w: %s date in far future", ErrTimeBounds, name)
	}
	if tim.Before(fip) {
		return fmt.Errorf("%w: %s date in far past", ErrTimeBounds, name)
	}
	return nil
}

// ToZone takes time.Location instance and returns corresponding Zone instance.
func ToZone(loc *time.Location) (*Zone, error) {
	tzs := loc.String()
	if tzs == "Local" {
		return nil, ErrLocalTimezone
	}
	zon, ok := Zones[tzs]
	if !ok {
		return nil, ErrInvTimezone
	}
	return zon, nil
}

// ByZoneID takes timezone ID and returns corresponding time.Location instance.
func ByZoneID(tzi int) (*time.Location, error) {
	tz, ok := ZoneIDs[tzi]
	if !ok {
		return nil, ErrInvTimezone
	}
	loc, err := tz.Loc()
	if err != nil {
		return nil, ErrInvTimezone
	}
	return loc, nil
}
