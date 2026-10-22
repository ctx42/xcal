package tzgen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"

	"github.com/ctx42/xcal/pkg/xcal"
)

func Test_Merge(t *testing.T) {
	// --- Given ---
	prev := []*xcal.Zone{
		{Id: 0, Name: "UTC", Status: xcal.Alias, Link: "Etc/UTC"},
		{
			Id:          416,
			Name:        "Europe/Warsaw",
			Status:      xcal.Canonical,
			Offset:      120,
			OffsetDST:   180,
			Notes:       "Poland",
			CountryCode: "PL",
		},
		{
			Id:     417,
			Name:   "Europe/Gone",
			Status: xcal.Removed,
			Link:   "Europe/Berlin",
		},
	}
	ups := must.Value(ReadUpstream(
		"testdata/tzdata",
		"testdata/zoneinfo.zip",
		2026,
	))

	// --- When ---
	have, err := Merge(prev, ups)

	// --- Then ---
	assert.NoError(t, err)
	want := []*xcal.Zone{
		{Id: 0, Name: "UTC", Status: xcal.Alias, Link: "Etc/UTC"},
		{
			Id:          416,
			Name:        "Europe/Warsaw",
			Status:      xcal.Canonical,
			Offset:      60,
			OffsetDST:   120,
			Notes:       "Poland",
			CountryCode: "PL",
		},
		{
			Id:     417,
			Name:   "Europe/Gone",
			Status: xcal.Removed,
			Link:   "Europe/Berlin",
		},
		{Id: 418, Name: "Etc/UTC", Status: xcal.Canonical},
		{
			Id:        419,
			Name:      "Europe/Kiev",
			Status:    xcal.Alias,
			Offset:    120,
			OffsetDST: 180,
			Link:      "Europe/Kyiv",
		},
		{
			Id:          420,
			Name:        "Europe/Kyiv",
			Status:      xcal.Canonical,
			Offset:      120,
			OffsetDST:   180,
			Notes:       "most of Ukraine",
			CountryCode: "UA",
		},
	}
	assert.Equal(t, want, have)
	assert.Equal(t, 0, ups[0].Id)
}

func Test_Merge_errors_tabular(t *testing.T) {
	tt := []struct {
		testN string

		prev []*xcal.Zone
		ups  []*xcal.Zone
		want string
	}{
		{
			"error - zone missing upstream",
			[]*xcal.Zone{
				{Id: 0, Name: "Etc/UTC", Status: xcal.Canonical},
				{Id: 1, Name: "Europe/Paris", Status: xcal.Canonical},
			},
			[]*xcal.Zone{{Name: "Etc/UTC", Status: xcal.Canonical}},
			"" +
				`zones missing upstream, mark them removed ("R") with a ` +
				"link to their replacement: Europe/Paris",
		},
		{
			"error - invalid upstream zone",
			nil,
			[]*xcal.Zone{
				{Name: "Etc/UTC", Status: xcal.Canonical, Offset: 1000},
			},
			"at index 0: invalid Zone offset 1000",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, err := Merge(tc.prev, tc.ups)

			// --- Then ---
			assert.ErrorEqual(t, tc.want, err)
			assert.Nil(t, have)
		})
	}
}

func Test_WriteZones(t *testing.T) {
	// --- Given ---
	pth := filepath.Join(t.TempDir(), "timezones.json")
	zs := []*xcal.Zone{
		{
			Id:          7,
			Name:        "Europe/Warsaw",
			Status:      xcal.Canonical,
			Offset:      60,
			OffsetDST:   120,
			Notes:       "A & B",
			CountryCode: "PL",
		},
	}

	// --- When ---
	err := WriteZones(pth, zs)

	// --- Then ---
	assert.NoError(t, err)

	want := "" +
		"[\n" +
		" {\n" +
		"  \"id\": 7,\n" +
		"  \"timezone\": \"Europe/Warsaw\",\n" +
		"  \"status\": \"C\",\n" +
		"  \"offset\": 60,\n" +
		"  \"offset_dst\": 120,\n" +
		"  \"link\": \"\",\n" +
		"  \"notes\": \"A & B\",\n" +
		"  \"country_code\": \"PL\"\n" +
		" }\n" +
		"]\n"
	have := must.Value(os.ReadFile(pth))
	assert.Equal(t, want, string(have))
}

func Test_WriteZones_noDirectory(t *testing.T) {
	// --- Given ---
	pth := filepath.Join(t.TempDir(), "not_existing", "timezones.json")

	// --- When ---
	err := WriteZones(pth, nil)

	// --- Then ---
	assert.ErrorContain(t, "not_existing/timezones.json", err)
}
