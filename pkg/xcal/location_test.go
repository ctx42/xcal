package xcal

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
)

func Test_LoadLocation_tabular(t *testing.T) {
	tt := []struct {
		testN string

		locS string
		wLoc string
		wID  int
	}{
		{"Warsaw name", "Europe/Warsaw", "Europe/Warsaw", 416},
		{"Zurich name", "Europe/Zurich", "Europe/Zurich", 419},
		{"Warsaw ID", "416", "Europe/Warsaw", 416},
		{"Zurich ID", "419", "Europe/Zurich", 419},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, err := LoadLocation(tc.locS)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.wLoc, have.Loc.String())
			assert.Equal(t, tc.wID, have.ID)
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
	zon := &Zone{
		Id:          999,
		Name:        "Europe/Krakow",
		Status:      "C",
		Offset:      120,
		OffsetDST:   180,
		Link:        "",
		Notes:       "",
		CountryCode: "PL",
	}
	Zones["Europe/Krakow"] = zon
	ZoneIDs[999] = zon
	t.Cleanup(func() {
		delete(Zones, "Europe/Krakow")
		delete(ZoneIDs, 999)
	})

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

func Test_Location_MarshalJSON(t *testing.T) {
	// --- Given ---
	tz := must.Value(LoadLocation("Europe/Warsaw"))

	// --- When ---
	have, err := json.Marshal(tz)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, `"Europe/Warsaw"`, string(have))
}

func Test_Location_MarshalJSON_inStructure(t *testing.T) {
	// --- Given ---
	loc := must.Value(LoadLocation("Europe/Warsaw"))

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

func Test_Location_MarshalJSON_zero(t *testing.T) {
	// --- Given ---
	ts := struct {
		Loc Location `json:"my_loc"`
	}{}

	// --- When ---
	have, err := json.Marshal(ts)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, `{"my_loc":null}`, string(have))
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
