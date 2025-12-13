package util

import (
	"time"

	"todo-api/internal/constants"
)

// FormatDate formats a time.Time pointer to a date string (YYYY-MM-DD).
// Returns nil if t is nil.
func FormatDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(constants.DateFormat)
	return &s
}

// FormatDateTime formats a time.Time to a datetime string (ISO 8601).
func FormatDateTime(t time.Time) string {
	return t.Format(constants.DateTimeFormat)
}

// FormatRFC3339 formats a time.Time to RFC3339 format.
func FormatRFC3339(t time.Time) string {
	return t.Format(time.RFC3339)
}

// ParseDate parses a date string (YYYY-MM-DD) to time.Time.
// Returns nil and an error if the string is empty or invalid.
func ParseDate(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse(constants.DateFormat, s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Today returns today's date at midnight (00:00:00).
func Today() time.Time {
	return time.Now().Truncate(24 * time.Hour)
}

// IsBeforeToday checks if the given time is before today.
func IsBeforeToday(t time.Time) bool {
	return t.Before(Today())
}
