package tzgen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"

	"github.com/ctx42/xcal/pkg/xcal"
)

func Test_ReadZones(t *testing.T) {
	// --- When ---
	have, err := ReadZones("testdata/timezones.json")

	// --- Then ---
	assert.NoError(t, err)
	assert.Len(t, 2, have)

	zon := have[0]
	assert.Equal(t, 0, zon.Id)
	assert.Equal(t, "UTC", zon.Name)
	assert.Equal(t, "A", zon.Status)
	assert.Equal(t, 0, zon.Offset)
	assert.Equal(t, 0, zon.OffsetDST)
	assert.Equal(t, "Etc/UTC", zon.Link)
	assert.Equal(t, "test note", zon.Notes)
	assert.Equal(t, "", zon.CountryCode)

	zon = have[1]
	assert.Equal(t, 416, zon.Id)
	assert.Equal(t, "Europe/Warsaw", zon.Name)
	assert.Equal(t, "C", zon.Status)
	assert.Equal(t, 120, zon.Offset)
	assert.Equal(t, 180, zon.OffsetDST)
	assert.Equal(t, "", zon.Link)
	assert.Equal(t, "", zon.Notes)
	assert.Equal(t, "PL", zon.CountryCode)
}

func Test_ReadZones_duplicateID(t *testing.T) {
	// --- When ---
	have, err := ReadZones("testdata/timezones_duplicate_id.json")

	// --- Then ---
	assert.ErrorEqual(t, "at index 2: duplicate ZoneID 416", err)
	assert.Nil(t, have)
}

func Test_ReadZones_duplicateName(t *testing.T) {
	// --- When ---
	have, err := ReadZones("testdata/timezones_duplicate_name.json")

	// --- Then ---
	assert.ErrorEqual(t, "at index 2: duplicate timezone UTC", err)
	assert.Nil(t, have)
}

func Test_ReadZones_invalidZone(t *testing.T) {
	// --- When ---
	have, err := ReadZones("testdata/timezones_invalid_zone.json")

	// --- Then ---
	assert.ErrorEqual(t, "at index 1: invalid Zone offset 1000", err)
	assert.Nil(t, have)
}

func Test_ReadZones_invalidJSON(t *testing.T) {
	// --- When ---
	have, err := ReadZones("testdata/timezones_invalid_json.json")

	// --- Then ---
	assert.ErrorEqual(t, "unexpected end of JSON input", err)
	assert.Nil(t, have)
}

func Test_ReadZones_notExistingFile(t *testing.T) {
	// --- When ---
	have, err := ReadZones("testdata/not_existing.json")

	// --- Then ---
	assert.ErrorContain(t, "testdata/not_existing.json", err)
	assert.Nil(t, have)
}

func Test_GenerateTZ(t *testing.T) {
	// --- Given ---
	dst := filepath.Join(t.TempDir(), "timezones.go")

	// --- When ---
	err := GenerateTZ("testdata/timezones.json", dst, "xcal", testPrev(t))

	// --- Then ---
	assert.NoError(t, err)

	want := must.Value(os.ReadFile("testdata/timezones_expected_go"))
	have := must.Value(os.ReadFile(dst))
	assert.Equal(t, string(want), string(have))
}

func Test_GenerateTZ_noSourceFile(t *testing.T) {
	// --- When ---
	err := GenerateTZ("testdata/not_existing.json", "", "xcal", nil)

	// --- Then ---
	assert.ErrorContain(t, "testdata/not_existing.json", err)
}

func Test_GenerateTZ_noDestinationDir(t *testing.T) {
	// --- Given ---
	dst := "testdata/not_existing/timezones.go"

	// --- When ---
	err := GenerateTZ("testdata/timezones.json", dst, "xcal", testPrev(t))

	// --- Then ---
	assert.ErrorContain(t, "testdata/not_existing/timezones.go", err)
}

func Test_GenerateTZ_unstableIDs(t *testing.T) {
	// --- Given ---
	dst := filepath.Join(t.TempDir(), "timezones.go")

	// --- When ---
	err := GenerateTZ("testdata/timezones.json", dst, "xcal", nil)

	// --- Then ---
	assert.ErrorEqual(t, "at index 1: expected ZoneID 1 got 416", err)
	assert.NoFileExist(t, dst)
}

func Test_DecodeZones(t *testing.T) {
	// --- When ---
	have, err := DecodeZones("testdata/timezones_invalid_zone.json")

	// --- Then ---
	assert.NoError(t, err)
	assert.Len(t, 2, have)
	assert.Equal(t, 1000, have[1].Offset)
}

func Test_DecodeZones_errors_tabular(t *testing.T) {
	tt := []struct {
		testN string

		pth  string
		want string
	}{
		{
			"error - missing file",
			"testdata/not_existing.json",
			"testdata/not_existing.json",
		},
		{
			"error - invalid JSON",
			"testdata/timezones_invalid_json.json",
			"unexpected end of JSON input",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, err := DecodeZones(tc.pth)

			// --- Then ---
			assert.ErrorContain(t, tc.want, err)
			assert.Nil(t, have)
		})
	}
}

func Test_CheckStable_tabular(t *testing.T) {
	tt := []struct {
		testN string

		prev map[string]*xcal.Zone
		zs   []*xcal.Zone
	}{
		{
			"no previous zones",
			nil,
			[]*xcal.Zone{{Id: 0, Name: "UTC"}, {Id: 1, Name: "Etc/UTC"}},
		},
		{
			"same IDs",
			map[string]*xcal.Zone{"UTC": {Id: 0, Name: "UTC"}},
			[]*xcal.Zone{{Id: 0, Name: "UTC"}},
		},
		{
			"new zone above previous IDs",
			map[string]*xcal.Zone{"UTC": {Id: 4, Name: "UTC"}},
			[]*xcal.Zone{{Id: 4, Name: "UTC"}, {Id: 5, Name: "Etc/UTC"}},
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := CheckStable(tc.prev, tc.zs)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_CheckStable_errors_tabular(t *testing.T) {
	tt := []struct {
		testN string

		prev map[string]*xcal.Zone
		zs   []*xcal.Zone
		want string
	}{
		{
			"error - not counting from zero",
			nil,
			[]*xcal.Zone{{Id: 0, Name: "UTC"}, {Id: 2, Name: "Etc/UTC"}},
			"at index 1: expected ZoneID 1 got 2",
		},
		{
			"error - changed ID",
			map[string]*xcal.Zone{"UTC": {Id: 0, Name: "UTC"}},
			[]*xcal.Zone{{Id: 1, Name: "UTC"}},
			"at index 0: expected ZoneID 0 got 1",
		},
		{
			"error - new zone reusing ID range",
			map[string]*xcal.Zone{
				"UTC":     {Id: 0, Name: "UTC"},
				"Etc/UTC": {Id: 2, Name: "Etc/UTC"},
			},
			[]*xcal.Zone{
				{Id: 0, Name: "UTC"},
				{Id: 1, Name: "Europe/Warsaw"},
				{Id: 2, Name: "Etc/UTC"},
			},
			"at index 1: new zone Europe/Warsaw has ID 1, expected above 2",
		},
		{
			"error - missing zone",
			map[string]*xcal.Zone{
				"UTC":     {Id: 0, Name: "UTC"},
				"Etc/UTC": {Id: 1, Name: "Etc/UTC"},
			},
			[]*xcal.Zone{{Id: 0, Name: "UTC"}},
			"zone Etc/UTC with ID 1 is missing",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := CheckStable(tc.prev, tc.zs)

			// --- Then ---
			assert.ErrorEqual(t, tc.want, err)
		})
	}
}

func Test_GenerateTZGo(t *testing.T) {
	// --- Given ---
	dst := filepath.Join(t.TempDir(), "timezones.go")
	zs := must.Value(ReadZones("testdata/timezones.json"))

	// --- When ---
	err := GenerateTZGo(zs, dst, "xcal")

	// --- Then ---
	assert.NoError(t, err)

	want := must.Value(os.ReadFile("testdata/timezones_expected_go"))
	have := must.Value(os.ReadFile(dst))
	assert.Equal(t, string(want), string(have))
}

func Test_zoneLit(t *testing.T) {
	// --- Given ---
	zon := &xcal.Zone{
		Id:          416,
		Name:        "Europe/Warsaw",
		Status:      "C",
		Offset:      120,
		OffsetDST:   180,
		Link:        "",
		Notes:       "say \"hi\"",
		CountryCode: "PL",
	}

	// --- When ---
	have := zoneLit(zon)

	// --- Then ---
	want := "" +
		"{\n" +
		"Id: 416,\n" +
		"Name: \"Europe/Warsaw\",\n" +
		"Status: \"C\",\n" +
		"Offset: 120,\n" +
		"OffsetDST: 180,\n" +
		"Link: \"\",\n" +
		"Notes: \"say \\\"hi\\\"\",\n" +
		"CountryCode: \"PL\",\n" +
		"}"
	assert.Equal(t, want, have)
}
