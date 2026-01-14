package timewin

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// Window represents a parsed time window with current and prior periods
type Window struct {
	CurrentStart time.Time
	CurrentEnd   time.Time
	PriorStart   time.Time
	PriorEnd     time.Time
	Duration     time.Duration
}

// Parse parses a duration string like "7d", "48h", "30d" and returns current/prior windows
// The current period is the most recent duration, prior is the same duration before that
func Parse(s string) (*Window, error) {
	re := regexp.MustCompile(`^(\d+)([hd])$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return nil, fmt.Errorf("invalid time window format: %s (expected format: 48h, 7d, 30d)", s)
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid numeric value: %w", err)
	}

	var duration time.Duration
	unit := matches[2]
	switch unit {
	case "h":
		duration = time.Duration(value) * time.Hour
	case "d":
		duration = time.Duration(value) * 24 * time.Hour
	default:
		return nil, fmt.Errorf("unsupported unit: %s", unit)
	}

	// Current period ends now (truncated to start of hour or day for consistency)
	// Prior period is the same duration before current period
	now := time.Now().UTC()

	// Truncate based on unit for cleaner boundaries
	var currentEnd time.Time
	if unit == "h" {
		currentEnd = now.Truncate(time.Hour)
	} else {
		currentEnd = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	}

	currentStart := currentEnd.Add(-duration)
	priorEnd := currentStart
	priorStart := priorEnd.Add(-duration)

	return &Window{
		CurrentStart: currentStart,
		CurrentEnd:   currentEnd,
		PriorStart:   priorStart,
		PriorEnd:     priorEnd,
		Duration:     duration,
	}, nil
}

// FormatCE formats a time for Cost Explorer API (YYYY-MM-DD)
func FormatCE(t time.Time) string {
	return t.Format("2006-01-02")
}

// IncludesToday returns true if the current window includes today (incomplete data warning)
// Note: Cost Explorer end dates are exclusive, so currentEnd = today means query excludes today
func (w *Window) IncludesToday() bool {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	return w.CurrentEnd.After(today)
}
