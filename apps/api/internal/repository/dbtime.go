package repository

import "time"

// SQLite stores DATETIME columns as text, and the driver may hand them back
// either verbatim ("2006-01-02 15:04:05") or normalized to RFC 3339.
var dbTimeLayouts = []string{
	time.RFC3339Nano,
	"2006-01-02 15:04:05.999999999-07:00",
	"2006-01-02 15:04:05.999999999",
	"2006-01-02 15:04:05",
}

// parseDBTime returns the zero time when the value cannot be parsed.
func parseDBTime(s string) time.Time {
	for _, layout := range dbTimeLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}
