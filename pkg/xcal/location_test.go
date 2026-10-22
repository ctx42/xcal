package xcal

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
)

func Test_LoadLocation(t *testing.T) {
	tt := []struct {
		testN string

		locS   string
		expLoc string
		expID  int
	}{
		{"1", "Europe/Warsaw", "Europe/Warsaw", 416},
		{"2", "Europe/Zurich", "Europe/Zurich", 419},
		{"3", "416", "Europe/Warsaw", 416},
		{"4", "419", "Europe/Zurich", 419},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			loc, err := LoadLocation(tc.locS)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.expLoc, loc.Loc.String())
		})
	}
}

func Test_LoadLocation_invalidZoneID(t *testing.T) {
	// --- When ---
	_, err := LoadLocation("123a")

	// --- Then ---
	assert.ErrorIs(t, ErrInvTimezone, err)
}

func Test_LoadLocation_unknownZoneID(t *testing.T) {
	// --- When ---
	_, err := LoadLocation("1000")

	// --- Then ---
	assert.ErrorIs(t, ErrInvTimezone, err)
}

func Test_LoadLocation_unknownTimezoneName(t *testing.T) {
	// --- When ---
	_, err := LoadLocation("Some/Unknown/Timezone")

	// --- Then ---
	assert.ErrorIs(t, ErrInvTimezone, err)
}

func Test_LoadLocation_timezoneNotInGoDatabase(t *testing.T) {
	// --- Given ---
	z := &Zone{
		Id:          999,
		Name:        "Europe/Krakow",
		Status:      "C",
		Offset:      120,
		OffsetDST:   180,
		Link:        "",
		Notes:       "",
		CountryCode: "PL",
	}
	Zones["Europe/Krakow"] = z
	ZoneIDs[999] = z
	defer func() { delete(Zones, "Europe/Krakow") }()

	// --- When ---
	loc, err := LoadLocation("999")

	// --- Then ---
	assert.ErrorIs(t, ErrInvTimezone, err)
	assert.Nil(t, loc)
}

func Test_LoadLocation_timezoneInGoDatabaseButNotOurs(t *testing.T) {
	// --- Given ---
	save := Zones["Europe/Warsaw"]
	delete(Zones, save.Name)
	delete(ZoneIDs, save.Id)
	defer func() {
		Zones[save.Name] = save
		ZoneIDs[save.Id] = save
	}()

	// --- When ---
	loc, err := LoadLocation("Europe/Warsaw")

	// --- Then ---
	assert.ErrorIs(t, ErrInvTimezone, err)
	assert.Nil(t, loc)
}

func Test_LoadLocation_emptyLocationString(t *testing.T) {
	// --- When ---
	_, err := LoadLocation("")

	// --- Then ---
	assert.ErrorIs(t, ErrInvTimezone, err)
}

func Test_Location_MarshalJSON(t *testing.T) {
	// --- Given ---
	tz, err := LoadLocation("Europe/Warsaw")
	assert.NoError(t, err)

	// --- When ---
	got, err := json.Marshal(tz)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, `"Europe/Warsaw"`, string(got))
}

func Test_Location_MarshalJSON_inStructure(t *testing.T) {
	// --- Given ---
	loc, err := LoadLocation("Europe/Warsaw")
	assert.NoError(t, err)

	var ts = struct {
		Loc *Location `json:"my_loc"`
	}{
		Loc: loc,
	}

	// --- When ---
	data, err := json.Marshal(ts)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, `{"my_loc":"Europe/Warsaw"}`, string(data))
}

func Test_Location_MarshalJSON_valueInStructure(t *testing.T) {
	// --- Given ---
	loc := must.Value(LoadLocation("Europe/Warsaw"))
	ts := struct {
		Loc Location `json:"my_loc"`
	}{
		Loc: *loc,
	}

	// --- When ---
	have, err := json.Marshal(ts)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, `{"my_loc":"Europe/Warsaw"}`, string(have))
}

func Test_Location_UnmarshalJSON(t *testing.T) {
	// --- Given ---
	loc := &Location{}

	// --- When ---
	err := json.Unmarshal([]byte(`"Europe/Warsaw"`), loc)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, "Europe/Warsaw", loc.Loc.String())
	assert.Equal(t, 416, loc.ID)
}

func Test_Location_UnmarshalJSON_badTimezone(t *testing.T) {
	// --- Given ---
	loc := &Location{}

	// --- When ---
	err := json.Unmarshal([]byte(`"Bad/Timezone"`), loc)

	// --- Then ---
	assert.ErrorIs(t, ErrInvTimezone, err)
}

func Test_Location_UnmarshalJSON_escapedString(t *testing.T) {
	// --- Given ---
	loc := &Location{}

	// --- When ---
	err := json.Unmarshal([]byte(`"Europe\/Warsaw"`), loc)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, "Europe/Warsaw", loc.Loc.String())
	assert.Equal(t, 416, loc.ID)
}

func Test_Location_UnmarshalJSON_number(t *testing.T) {
	// --- Given ---
	loc := &Location{}

	// --- When ---
	err := json.Unmarshal([]byte(`416`), loc)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, "Europe/Warsaw", loc.Loc.String())
	assert.Equal(t, 416, loc.ID)
}

func Test_Location_UnmarshalJSON_null(t *testing.T) {
	// --- Given ---
	have := struct {
		Loc Location `json:"my_loc"`
	}{
		Loc: Location{ID: 416, Loc: WAW},
	}

	// --- When ---
	err := json.Unmarshal([]byte(`{"my_loc":null}`), &have)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, 416, have.Loc.ID)
	assert.Same(t, WAW, have.Loc.Loc)
}

func Test_Location_UnmarshalJSON_inStructure(t *testing.T) {
	// --- Given ---
	type ts struct {
		Loc *Location `json:"my_loc"`
	}

	g := &ts{}

	// --- When ---
	err := json.Unmarshal([]byte(`{"my_loc":"Europe/Warsaw"}`), g)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, "Europe/Warsaw", g.Loc.Loc.String())
	assert.Equal(t, 416, g.Loc.ID)
}

func Test_Location_UnmarshalJSON_inStructure_ZoneID(t *testing.T) {
	// --- Given ---
	type ts struct {
		Loc *Location `json:"my_loc"`
	}

	g := &ts{}

	// --- When ---
	err := json.Unmarshal([]byte(`{"my_loc":"416"}`), g)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, "Europe/Warsaw", g.Loc.Loc.String())
	assert.Equal(t, 416, g.Loc.ID)
}

func Test_DefaultLocations(t *testing.T) {
	tt := []struct {
		testN string

		exp string
		loc *time.Location
	}{
		{"1", "Europe/Warsaw", WAW},
		{"2", "Europe/Zurich", ZRH},
		{"3", "Europe/Moscow", MOW},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			assert.Equal(t, tc.exp, tc.loc.String())
		})
	}
}

func Test_FromTimeLoc(t *testing.T) {
	// --- When ---
	loc, err := FromTimeLoc(WAW)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, "Europe/Warsaw", loc.Loc.String())
	assert.Equal(t, 416, loc.ID)
}

func Test_FromTimeLoc_UnknownTimezone(t *testing.T) {
	// --- Given ---
	tz := time.FixedZone("Europe/Krakow", int((2 * time.Hour).Seconds()))

	// --- When ---
	loc, err := FromTimeLoc(tz)

	// --- Then ---
	assert.ErrorIs(t, ErrNoZoneID, err)
	assert.Nil(t, loc)
}
