package timewin

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		checkPeriod bool
	}{
		{"valid 7 days", "7d", false, true},
		{"valid 48 hours", "48h", false, true},
		{"valid 30 days", "30d", false, true},
		{"invalid format", "7days", true, false},
		{"invalid no unit", "7", true, false},
		{"invalid empty", "", true, false},
		{"invalid unit", "7x", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkPeriod && w != nil {
				// Verify current and prior periods are equal duration
				currentDur := w.CurrentEnd.Sub(w.CurrentStart)
				priorDur := w.PriorEnd.Sub(w.PriorStart)
				if currentDur != priorDur {
					t.Errorf("Period mismatch: current=%v, prior=%v", currentDur, priorDur)
				}
				// Verify periods are adjacent
				if !w.PriorEnd.Equal(w.CurrentStart) {
					t.Errorf("Periods not adjacent: priorEnd=%v, currentStart=%v", w.PriorEnd, w.CurrentStart)
				}
			}
		})
	}
}

func TestFormatCE(t *testing.T) {
	dt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	got := FormatCE(dt)
	want := "2024-01-15"
	if got != want {
		t.Errorf("FormatCE() = %v, want %v", got, want)
	}
}

func TestIncludesToday(t *testing.T) {
	now := time.Now().UTC()
	today := now.Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	tests := []struct {
		name       string
		currentEnd time.Time
		want       bool
	}{
		{"window ends tomorrow (includes today)", tomorrow, true},
		{"window ends at start of today (excludes today)", today, false},
		{"window ends yesterday", today.Add(-24 * time.Hour), false},
		{"window ends in future", tomorrow.Add(24 * time.Hour), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Window{
				CurrentStart: tt.currentEnd.Add(-7 * 24 * time.Hour),
				CurrentEnd:   tt.currentEnd,
			}
			got := w.IncludesToday()
			if got != tt.want {
				t.Errorf("IncludesToday() = %v, want %v (currentEnd=%v, today=%v)",
					got, tt.want, tt.currentEnd, today)
			}
		})
	}
}
