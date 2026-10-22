[![Go Reference](https://pkg.go.dev/badge/github.com/ctx42/xcal.svg)](https://pkg.go.dev/github.com/ctx42/xcal)
[![Go Version](https://img.shields.io/github/go-mod/go-version/ctx42/xcal)](go.mod)
[![License](https://img.shields.io/github/license/ctx42/xcal)](LICENSE.md)
[![Tests](https://github.com/ctx42/xcal/actions/workflows/go.yml/badge.svg)](https://github.com/ctx42/xcal/actions/workflows/go.yml)

# xcal

Calendar days and timezone helpers for Go, with a stable numeric ID for every
IANA zone so you can store a zone as an `int`.

## Overview

Time zone names such as `Europe/Warsaw` are awkward to store and index, and the
tz database renames and aliases zones over time. `xcal` keeps a generated table
that gives every zone a numeric ID which never changes and is never reused, so
API and database code can persist a zone as an integer and still load it back
as a `*time.Location`.

Around that table it adds the date pieces such code keeps rewriting: a
calendar `Day` with no time of day, a `Location` that travels through JSON as
its zone name, and helpers for midnights, month ranges, and time windows.

## Features

- `Day` — a calendar date with no time or zone that round-trips through JSON
  and text as `YYYY-MM-DD`; the zero `Day` encodes as empty text.
- `Location` — a `*time.Location` paired with its zone ID; it encodes to JSON
  as the zone name and decodes from a zone name or ID.
- Zone tables — `Zones` and `ZoneIDs` map names and IDs to `Zone` records
  with offsets, status, alias links, and country codes, kept in step with the
  IANA tz database the Go toolchain embeds.
- Embedded tz database — zones load on hosts without zoneinfo, such as scratch
  or distroless containers, when the Go release the program is built with
  knows them.
- Date helpers — `Midnight`, `MonthStartEnd`, `InWindow`, `ChangeLocation`,
  and more.
- Bounds checks — `CheckTimeBounds` rejects dates too far in the past or
  future with an error wrapping `ErrTimeBounds`.

## Prerequisites

- Go 1.26 or later.

## Installation

```shell
go get github.com/ctx42/xcal
```

```go
import "github.com/ctx42/xcal/pkg/xcal"
```

## Usage

### Load a zone by name or ID

`LoadLocation` accepts a zone name or its numeric ID as a string, and returns
the `*time.Location` with its ID:

<!-- gmmce:pkg/xcal/ExampleLoadLocation -->
```go
loc, err := xcal.LoadLocation("416")
if err != nil {
	fmt.Println(err)
	return
}
fmt.Println(loc.Loc, loc.ID)
// Output: Europe/Warsaw 416
```

### Zones in JSON

A `Location` field decodes from a zone name, an ID, or `null`, and always
encodes as the zone name:

<!-- gmmce:pkg/xcal/ExampleLocation_UnmarshalJSON -->
```go
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
```

### Calendar days

A `Day` names a date without committing to a time zone; `In` turns it into the
instant the day starts at in a given location:

<!-- gmmce:pkg/xcal/ExampleDay -->
```go
day := xcal.NewDay(2026, time.March, 25)
fmt.Println(day.AddDays(7))
fmt.Println(day.In(xcal.WAW).UTC())
// Output:
// 2026-04-01
// 2026-03-24 23:00:00 +0000 UTC
```

### Date helpers

| Function             | Returns                                               |
|----------------------|-------------------------------------------------------|
| `Midnight`           | The start of the day of a time, in its location.      |
| `SinceMidnight`      | The time elapsed since that midnight.                 |
| `MonthStartEnd`      | The start of the month and of the next month.         |
| `InWindow`           | Whether a time falls in a daily window from midnight. |
| `InBetween`          | Whether a time is strictly between two others.        |
| `InBetweenInclusive` | Whether a time is between two others, ends included.  |
| `ChangeLocation`     | The same wall clock time in another location.         |
| `CheckTimeBounds`    | An error for a time too far in the past or future.    |
| `ToZone`, `ByZoneID` | The `Zone` for a location, the location for an ID.    |

## Updating the zone table

The table follows the IANA tz database release the Go toolchain embeds. After
upgrading Go, refresh the registry and regenerate the tables from the module
root:

```shell
go run ./internal/tzgen/cmd/tzupdate
go generate ./pkg/xcal
```

`tzupdate` downloads the release's `zone.tab` and links into
`pkg/xcal/data/tzdata` unless they are already there, and merges them into
`pkg/xcal/data/timezones.json`, the registry of zone IDs:

- Existing zones keep their IDs; new zones get IDs above every existing one.
- A renamed zone stays as an alias of its new name, under its old ID.
- A zone dropped from the database stops the update until its registry entry
  is marked removed (`"status": "R"`) with a `link` to its replacement. It
  keeps its ID and loads as the replacement.

`go generate` refuses a registry that changes, drops, or reuses an ID.

## License

MIT License — see [LICENSE.md](LICENSE.md).
