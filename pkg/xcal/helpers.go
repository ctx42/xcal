package xcal

import (
	"errors"
	"fmt"
	"time"
)

// ErrTimeBounds is wrapped by errors returned when time is either far in the
// past or far in the future.
var ErrTimeBounds = errors.New("date out of bounds")

// CheckTimeBounds returns error if tim is before fip or after fif.
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
	z, ok := Zones[tzs]
	if !ok {
		return nil, ErrInvTimezone
	}
	return z, nil
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
