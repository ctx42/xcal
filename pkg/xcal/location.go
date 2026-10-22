package xcal

import (
	"bytes"
	"errors"
	"strconv"
	"time"

	// Embed the timezone database, so zones load on hosts without zoneinfo.
	_ "time/tzdata"
)

// Useful time.Location instances.
var (
	WAW *time.Location // Warsaw.
	ZRH *time.Location // Zurich.
	MOW *time.Location // Moscow.
)

// ErrInvTimezone is an error returned when timezone name unmarshalling
// failed.
var ErrInvTimezone = errors.New("invalid timezone")

// ErrNoZoneID indicates there is no zone ID for given timezone.
var ErrNoZoneID = errors.New("no zone ID for timezone name")

// ErrLocalTimezone represents an error when timezone is set to Local.
var ErrLocalTimezone = errors.New("timezone Local means nothing")

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

// Location is a wrapper for time.Location instance providing
// MarshalJSON and UnmarshalJSON methods.
type Location struct {
	ID  int            `json:"ID"`  // Timezone ID.
	Loc *time.Location `json:"Loc"` // Timezone.
}

// LoadLocation creates a new instance of [Location] from a full time zone name
// or its string numerical representation (e.g. "Europe/Warsaw" or "416").
func LoadLocation(locS string) (*Location, error) {
	if locS == "" {
		return nil, ErrInvTimezone
	}

	// Handle location being Zone ID.
	if locS[0] >= '0' && locS[0] <= '9' {
		zid, err := strconv.Atoi(locS)
		if err != nil {
			return nil, ErrInvTimezone
		}
		zone, ok := ZoneIDs[zid]
		if !ok {
			return nil, ErrInvTimezone
		}
		tLoc, err := zone.Loc()
		if err != nil {
			// Zone is not in the Go location database.
			return nil, ErrInvTimezone
		}

		loc := &Location{
			ID:  zone.Id,
			Loc: tLoc,
		}

		return loc, nil
	}

	tLoc, err := time.LoadLocation(locS)
	if err != nil {
		return nil, ErrInvTimezone
	}

	zone, ok := Zones[tLoc.String()]
	if !ok {
		// Location name is not on our timezone list.
		return nil, ErrInvTimezone
	}

	loc := &Location{
		ID:  zone.Id,
		Loc: tLoc,
	}

	return loc, nil
}

// FromTimeLoc creates Location from time.Location. If the zone ID for the
// given [time.Location] cannot be found, [ErrNoZoneID] is returned.
func FromTimeLoc(loc *time.Location) (*Location, error) {
	if zone, ok := Zones[loc.String()]; ok {
		ret := &Location{
			ID:  zone.Id,
			Loc: loc,
		}
		return ret, nil
	}
	return nil, ErrNoZoneID
}

func (loc Location) MarshalJSON() ([]byte, error) {
	return []byte(`"` + loc.Loc.String() + `"`), nil
}

func (loc *Location) UnmarshalJSON(b []byte) error {
	tz, err := LoadLocation(string(bytes.Trim(b, `"`)))
	if err != nil {
		return ErrInvTimezone
	}
	*loc = *tz
	return nil
}
