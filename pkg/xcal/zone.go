package xcal

import (
	"fmt"
	"time"
)

//go:generate go run 00_generate_main.go

// Timezone statuses.
const (
	// Canonical represents the primary, preferred zone name.
	Canonical = "C"

	// Alias represents an alternative name, which may fit better within a
	// particular country.
	Alias = "A"

	// Deprecated represents an older style name, left in the tz database for
	// backwards compatibility, which should generally not be used.
	Deprecated = "D"

	// Removed represents a name dropped from the tz database. It is kept so
	// its ID stays reserved and keeps loading: it loads as the zone its Link
	// names, so its location reports that zone's name.
	Removed = "R"
)

// Zone represents timezone information taken from
// ICANN TZ database. See https://en.wikipedia.org/wiki/Tz_database.
type Zone struct {
	// Timezone ID.
	Id int `json:"id"`

	// Timezone name.
	// The Area and Location names have a maximum length of 14 characters.
	Name string `json:"timezone"`

	// Timezone status.
	Status string `json:"status"`

	// Offset in minutes from UTC.
	Offset int `json:"offset"`

	// Offset in minutes from UTC during DST.
	OffsetDST int `json:"offset_dst"`

	// When timezone is an alias this points to some other timezone.
	Link string `json:"link"`

	// Free text notes about the timezone.
	Notes string `json:"notes"`

	// Two-letter ISO country code.
	CountryCode string `json:"country_code"`
}

// Check validates zone during import of new timezones.
func (zon *Zone) Check() error {
	if zon.Id < 0 {
		return fmt.Errorf("invalid ZoneID %d", zon.Id)
	}

	if zon.Status != Removed {
		if _, err := time.LoadLocation(zon.Name); err != nil {
			return fmt.Errorf("%w: %s", ErrInvTimezone, zon.Name)
		}
	}

	if zon.Offset < -12*60 || zon.Offset > 14*60 {
		return fmt.Errorf("invalid Zone offset %d", zon.Offset)
	}

	if zon.OffsetDST < -12*60 || zon.OffsetDST > 14*60 {
		return fmt.Errorf("invalid Zone DST offset %d", zon.OffsetDST)
	}

	switch zon.Status {
	case Canonical, Alias, Deprecated, Removed:
	default:
		return fmt.Errorf("invalid Zone status %q", zon.Status)
	}

	if zon.Status == Alias || zon.Status == Removed {
		if zon.Link == "" {
			format := "zone %s with status %q has no link"
			return fmt.Errorf(format, zon.Name, zon.Status)
		}

		if _, err := time.LoadLocation(zon.Link); err != nil {
			return fmt.Errorf("invalid link %s", zon.Link)
		}
	}

	return nil
}

// Validate runs [Zone.Check] and also requires the zone name to be in
// [Zones], returning [ErrInvTimezone] otherwise.
func (zon *Zone) Validate() error {
	if err := zon.Check(); err != nil {
		return err
	}
	if _, err := LoadLocation(zon.Name); err != nil {
		return err
	}
	return nil
}

// Loc returns the zone's location, or [ErrInvTimezone] when the Go location
// database does not know it. A [Removed] zone returns its Link's location.
func (zon *Zone) Loc() (*time.Location, error) {
	name := zon.Name
	if zon.Status == Removed {
		name = zon.Link
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, ErrInvTimezone
	}
	return loc, nil
}
