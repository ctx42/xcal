package xcal

import (
	"encoding/json"
	"strconv"
	"time"
)

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

// MarshalJSON encodes the zone name, or null when Loc is nil.
func (loc Location) MarshalJSON() ([]byte, error) {
	if loc.Loc == nil {
		return []byte("null"), nil
	}
	return []byte(`"` + loc.Loc.String() + `"`), nil
}

// UnmarshalJSON accepts a zone name or ID, as a JSON string or a number. A
// JSON null leaves the Location unchanged.
func (loc *Location) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	locS := string(b)
	if len(b) > 0 && b[0] == '"' {
		if err := json.Unmarshal(b, &locS); err != nil {
			return ErrInvTimezone
		}
	}
	tz, err := LoadLocation(locS)
	if err != nil {
		return ErrInvTimezone
	}
	*loc = *tz
	return nil
}
