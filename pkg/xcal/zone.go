package xcal

import (
	"fmt"
	"time"
)

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
func (z *Zone) Check() error {
	if z.Id < 0 {
		return fmt.Errorf("invalid ZoneID %d", z.Id)
	}

	if _, err := time.LoadLocation(z.Name); err != nil {
		return fmt.Errorf("%w: %s", ErrInvTimezone, z.Name)
	}

	if z.Offset < -12*60 || z.Offset > 14*60 {
		return fmt.Errorf("invalid Zone offset %d", z.Offset)
	}

	if z.OffsetDST < -12*60 || z.OffsetDST > 14*60 {
		return fmt.Errorf("invalid Zone DST offset %d", z.OffsetDST)
	}

	if z.Status == Alias {
		if z.Link == "" {
			return fmt.Errorf("alias without link for %s", z.Name)
		}

		if _, err := time.LoadLocation(z.Link); err != nil {
			return fmt.Errorf("invalid alias link %s", z.Link)
		}
	}

	return nil
}

// Validate validates zone.
func (z *Zone) Validate() error {
	if err := z.Check(); err != nil {
		return err
	}
	if _, err := LoadLocation(z.Name); err != nil {
		return err
	}
	return nil
}

// Loc returns timezone location.
func (z *Zone) Loc() (*time.Location, error) {
	loc, err := LoadLocation(z.Name)
	if err != nil {
		return nil, err
	}
	return loc.Loc, nil
}
