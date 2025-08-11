package tools

import "time"

func GetCETLocation() *time.Location {
	loc, err := time.LoadLocation("CET")
	if err != nil {
		// Fallback to UTC if CET is not found
		return time.UTC
	}
	return loc
}
