package calc

import (
	"math"
	"time"
)

// FermentationDays returns elapsed days since startedAt ("2006-01-02" format).
func FermentationDays(startedAt string) (int, error) {
	t, err := time.Parse("2006-01-02", startedAt)
	if err != nil {
		return 0, err
	}
	return daysSince(t), nil
}

// DaysSince returns elapsed days from an ISO-8601 timestamp to now.
func DaysSince(ts string) (int, error) {
	t, err := parseTimestamp(ts)
	if err != nil {
		return 0, err
	}
	return daysSince(t), nil
}

// DaysBetween returns elapsed days between two ISO-8601 timestamps.
func DaysBetween(from, to string) (int, error) {
	tf, err := parseTimestamp(from)
	if err != nil {
		return 0, err
	}
	tt, err := parseTimestamp(to)
	if err != nil {
		return 0, err
	}
	return int(math.Round(tt.Sub(tf).Hours() / 24)), nil
}

func daysSince(t time.Time) int {
	return int(math.Round(time.Since(t).Hours() / 24))
}

func parseTimestamp(s string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02T15:04:05Z", s); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", s)
}
