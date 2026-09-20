package tools

import (
	"sync"
	"time"
)

// cetLocation is resolved once: time.LoadLocation reads the tz database on every call,
// and GetCETLocation is used inside hot loops.
//
//nolint:gochecknoglobals // memoization of an immutable value
var cetLocation = sync.OnceValue(func() *time.Location {
	loc, err := time.LoadLocation("CET")
	if err != nil {
		// Fallback to UTC if CET is not found
		return time.UTC
	}
	return loc
})

func GetCETLocation() *time.Location {
	return cetLocation()
}
