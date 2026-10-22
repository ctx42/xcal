package xcal_test

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ctx42/xcal/pkg/xcal"
)

func ExampleLoadLocation() {
	loc, err := xcal.LoadLocation("416")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(loc.Loc, loc.ID)
	// Output: Europe/Warsaw 416
}

func ExampleLocation_UnmarshalJSON() {
	var have struct {
		Loc xcal.Location `json:"loc"`
	}
	data := []byte(`{"loc":"Europe/Zurich"}`)
	if err := json.Unmarshal(data, &have); err != nil {
		fmt.Println(err)
		return
	}
	data, err := json.Marshal(have)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(have.Loc.ID, string(data))
	// Output: 419 {"loc":"Europe/Zurich"}
}

func ExampleDay() {
	day := xcal.NewDay(2026, time.March, 25)
	fmt.Println(day.AddDays(7))
	fmt.Println(day.In(xcal.WAW).UTC())
	// Output:
	// 2026-04-01
	// 2026-03-24 23:00:00 +0000 UTC
}
