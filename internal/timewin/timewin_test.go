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
	w := &Window{
		CurrentStart: now.Add(-24 * time.Hour),
		CurrentEnd:   now,
	}
	if !w.IncludesToday() {
		t.Error("Expected IncludesToday() to be true for window ending now")
	}
}
