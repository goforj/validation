package main

import (
	"fmt"
	"time"

	"github.com/goforj/validation"
)

// main validates a booking window and its submitted calendar date.
func main() {
	// snippet:time
	starts := time.Date(2026, 11, 10, 10, 0, 0, 0, time.UTC)
	ends := starts.Add(2 * time.Hour)

	err := validation.Validate(
		validation.Time("starts_at", starts).
			Required().
			Before(ends).
			Min(starts).
			Max(ends).
			Equal(starts),

		validation.Time("ends_at", ends).After(starts),

		validation.String("date", "2026-11-10").
			IsDate(time.DateOnly),
		validation.String("local_time", "2026-11-10 10:00").
			IsDateIn("2006-01-02 15:04", time.UTC),
	)

	fmt.Println(err)
	// <nil>
	// endsnippet:time
}

// Output:
// <nil>
