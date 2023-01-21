package common

import "time"

// truncateToDay function takes in a single argument,
// a time.Time value, and returns a new time.Time value
// that represents the same calendar day as the original value,
// but with the time truncated to the beginning of the day (midnight).
func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
