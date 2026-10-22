package tzgen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ctx42/testing/pkg/assert"

	"github.com/ctx42/xcal/pkg/xcal"
)

func Test_ReadZones(t *testing.T) {
	// --- When ---
	zs, err := ReadZones("testdata/timezones.json")

	// --- Then ---
	assert.NoError(t, err)
	assert.Len(t, 2, zs)

	z := zs[0]
	assert.Equal(t, 0, z.Id)
	assert.Equal(t, "UTC", z.Name)
	assert.Equal(t, "A", z.Status)
	assert.Equal(t, 0, z.Offset)
	assert.Equal(t, 0, z.OffsetDST)
	assert.Equal(t, "Etc/UTC", z.Link)
	assert.Equal(t, "test note", z.Notes)
	assert.Equal(t, "", z.CountryCode)

	z = zs[1]
	assert.Equal(t, 416, z.Id)
	assert.Equal(t, "Europe/Warsaw", z.Name)
	assert.Equal(t, "C", z.Status)
	assert.Equal(t, 120, z.Offset)
	assert.Equal(t, 180, z.OffsetDST)
	assert.Equal(t, "", z.Link)
	assert.Equal(t, "", z.Notes)
	assert.Equal(t, "PL", z.CountryCode)
}

func Test_ReadZones_mismatchedIDs(t *testing.T) {
	// --- When ---
	zs, err := ReadZones("testdata/timezones_mismatched_ids.json")

	// --- Then ---
	assert.ErrorEqual(t, "at index 1: expected ZoneID 416 got 2", err)
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
	assert.Error(t, err)
	assert.Nil(t, zs)
}

func Test_GenerateTZ(t *testing.T) {
	// --- Given ---
	dst := filepath.Join(t.TempDir(), "timezones.go")

	// --- When ---
	err := GenerateTZ("testdata/timezones.json", dst, "xcal")

	// --- Then ---
	assert.NoError(t, err)

	exp, err := os.ReadFile("testdata/timezones_expected_go")
	assert.NoError(t, err)

	got, err := os.ReadFile(dst)
	assert.NoError(t, err)
	assert.Equal(t, string(exp), string(got))
}

func Test_GenerateTZ_noSourceFile(t *testing.T) {
	// --- When ---
	err := GenerateTZ("testdata/not_existing.json", "", "xcal")

	// --- Then ---
	assert.Error(t, err)
}

func Test_GenerateTZ_noDestinationDir(t *testing.T) {
	// --- When ---
	dst := "testdata/not_existing/timezones.go"
	err := GenerateTZ("testdata/timezones.json", dst, "xcal")

	// --- Then ---
	assert.Error(t, err)
}

func Test_GenerateTZGo(t *testing.T) {
	// --- Given ---
	dst := filepath.Join(t.TempDir(), "timezones.go")
	zs, err := ReadZones("testdata/timezones.json")
	assert.NoError(t, err)

	// --- When ---
	err = GenerateTZGo(zs, dst, "xcal")

	// --- Then ---
	assert.NoError(t, err)

	exp, err := os.ReadFile("testdata/timezones_expected_go")
	assert.NoError(t, err)

	got, err := os.ReadFile(dst)
	assert.NoError(t, err)
	assert.Equal(t, string(exp), string(got))
}

func Test_zoneLit(t *testing.T) {
	// --- Given ---
	z := &xcal.Zone{
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
	have := zoneLit(z)

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
