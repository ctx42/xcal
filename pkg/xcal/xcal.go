// Package xcal provides calendar and timezone helpers built on [time].
//
// The main types are:
//   - [Day], a calendar date with no time of day, encoded as "YYYY-MM-DD".
//   - [Location], a [time.Location] paired with a stable numeric zone ID from
//     the generated [Zones] and [ZoneIDs] tables, encoded in JSON as the zone
//     name.
//
// Functions such as [Midnight], [MonthStartEnd], [InWindow] and
// [CheckTimeBounds] cover common date arithmetic and checks.
//
// Import the package as:
//
//	import "github.com/ctx42/xcal/pkg/xcal"
//
// The timezone database is embedded, so zones the build's Go release knows
// load on hosts without zoneinfo. See the package examples for usage.
package xcal

import (
	"errors"
	"time"

	// Embed the timezone database, so zones load on hosts without zoneinfo.
	_ "time/tzdata"
)

// RFC339MICRO is the RFC 3339 layout with up to six fractional second digits
// (microseconds).
const RFC339MICRO = "2006-01-02T15:04:05.999999Z07:00"

// Useful time.Location instances.
var (
	WAW *time.Location // Warsaw.
	ZRH *time.Location // Zurich.
	MOW *time.Location // Moscow.
)

// Sentinel errors.
var (
	// ErrInvTimezone is an error returned when a timezone name or ID is not
	// valid.
	ErrInvTimezone = errors.New("invalid timezone")

	// ErrNoZoneID indicates there is no zone ID for given timezone.
	ErrNoZoneID = errors.New("no zone ID for timezone name")

	// ErrLocalTimezone represents an error when timezone is set to Local.
	ErrLocalTimezone = errors.New("timezone Local means nothing")

	// ErrTimeBounds is wrapped by errors returned when time is either far in
	// the past or far in the future.
	ErrTimeBounds = errors.New("date out of bounds")
)

func init() {
	var err error
	WAW, err = time.LoadLocation("Europe/Warsaw")
	if err != nil {
		panic(err)
	}

	ZRH, err = time.LoadLocation("Europe/Zurich")
	if err != nil {
		panic(err)
	}

	MOW, err = time.LoadLocation("Europe/Moscow")
	if err != nil {
		panic(err)
	}
}
