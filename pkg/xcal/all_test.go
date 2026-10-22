package xcal

// validZone returns a new Zone that passes [Zone.Check] and [Zone.Validate].
func validZone() *Zone {
	return &Zone{
		ID:          416,
		CountryCode: "PL",
		Name:        "Europe/Warsaw",
		Offset:      120,
		OffsetDST:   180,
		Status:      "C",
		Link:        "",
		Notes:       "",
	}
}
