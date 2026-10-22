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
	zs, err := ReadZones("testdata/timezones.json")

	// --- Then ---
	assert.NoError(t, err)
	assert.Len(t, 2, zs)

	zon := zs[0]
	assert.Equal(t, 0, zon.Id)
	assert.Equal(t, "UTC", zon.Name)
	assert.Equal(t, "A", zon.Status)
	assert.Equal(t, 0, zon.Offset)
	assert.Equal(t, 0, zon.OffsetDST)
	assert.Equal(t, "Etc/UTC", zon.Link)
	assert.Equal(t, "test note", zon.Notes)
	assert.Equal(t, "", zon.CountryCode)

	zon = zs[1]
	assert.Equal(t, 416, zon.Id)
	assert.Equal(t, "Europe/Warsaw", zon.Name)
	assert.Equal(t, "C", zon.Status)
	assert.Equal(t, 120, zon.Offset)
	assert.Equal(t, 180, zon.OffsetDST)
	assert.Equal(t, "", zon.Link)
	assert.Equal(t, "", zon.Notes)
	assert.Equal(t, "PL", zon.CountryCode)
}

func Test_ReadZones_mismatchedIDs(t *testing.T) {
	// --- When ---
	zs, err := ReadZones("testdata/timezones_mismatched_ids.json")

	// --- Then ---
	assert.ErrorEqual(t, "at index 1: expected ZoneID 416 got 2", err)
	assert.Nil(t, zs)
}

func Test_ReadZones_duplicateID(t *testing.T) {
	// --- When ---
	zs, err := ReadZones("testdata/timezones_duplicate_id.json")

	// --- Then ---
	assert.ErrorEqual(t, "at index 2: duplicate ZoneID 416", err)
	assert.Nil(t, zs)
}

func Test_ReadZones_duplicateName(t *testing.T) {
	// --- When ---
	zs, err := ReadZones("testdata/timezones_duplicate_name.json")

	// --- Then ---
	assert.ErrorEqual(t, "at index 2: duplicate timezone UTC", err)
	assert.Nil(t, zs)
}

func Test_ReadZones_invalidZone(t *testing.T) {
	// --- When ---
	zs, err := ReadZones("testdata/timezones_invalid_zone.json")

	// --- Then ---
	assert.ErrorEqual(t, "at index 1: invalid Zone offset 1000", err)
	assert.Nil(t, zs)
}

func Test_ReadZones_noGeneratedZones(t *testing.T) {
	// --- Given ---
	save := xcal.Zones
	xcal.Zones = map[string]*xcal.Zone{}
	t.Cleanup(func() { xcal.Zones = save })

	// --- When ---
	zs, err := ReadZones("testdata/timezones_mismatched_ids.json")

	// --- Then ---
	assert.ErrorEqual(t, "at index 1: expected ZoneID 1 got 2", err)
	assert.Nil(t, zs)
}

func Test_ReadZones_invalidJSON(t *testing.T) {
	// --- When ---
	zs, err := ReadZones("testdata/timezones_invalid_json.json")

	// --- Then ---
	assert.ErrorEqual(t, "unexpected end of JSON input", err)
	assert.Nil(t, zs)
}

func Test_ReadZones_notExistingFile(t *testing.T) {
	// --- When ---
	zs, err := ReadZones("testdata/not_existing.json")

	// --- Then ---
	assert.ErrorContain(t, "testdata/not_existing.json", err)
	assert.Nil(t, zs)
}

func Test_GenerateTZ(t *testing.T) {
	// --- Given ---
	dst := filepath.Join(t.TempDir(), "timezones.go")

	// --- When ---
	err := GenerateTZ("testdata/timezones.json", dst, "xcal")

	// --- Then ---
	assert.NoError(t, err)

	want := must.Value(os.ReadFile("testdata/timezones_expected_go"))
	have := must.Value(os.ReadFile(dst))
	assert.Equal(t, string(want), string(have))
}

func Test_GenerateTZ_noSourceFile(t *testing.T) {
	// --- When ---
	err := GenerateTZ("testdata/not_existing.json", "", "xcal")

	// --- Then ---
	assert.ErrorContain(t, "testdata/not_existing.json", err)
}

func Test_GenerateTZ_noDestinationDir(t *testing.T) {
	// --- Given ---
	dst := "testdata/not_existing/timezones.go"

	// --- When ---
	err := GenerateTZ("testdata/timezones.json", dst, "xcal")

	// --- Then ---
	assert.ErrorContain(t, "testdata/not_existing/timezones.go", err)
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
