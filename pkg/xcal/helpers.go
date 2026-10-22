package xcal

import (
	"fmt"
	"strings"
	"time"
)

// CheckTimeBounds returns an error wrapping [ErrTimeBounds] when tim is before
// fip or later than now plus fif. The name prefixes the error message.
func CheckTimeBounds(name string, tim, fip time.Time, fif time.Duration) error {
	if tim.After(time.Now().UTC().Add(fif)) {
		msg := strings.TrimSpace(name + " date in far future")
		return fmt.Errorf("%w: %s", ErrTimeBounds, msg)
	}
	if tim.Before(fip) {
		msg := strings.TrimSpace(name + " date in far past")
		return fmt.Errorf("%w: %s", ErrTimeBounds, msg)
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
