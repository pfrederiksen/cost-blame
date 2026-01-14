package anomaly

import (
	"math"
	"testing"
)

func TestComputeStats(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		wantMean float64
		wantStdDev float64
	}{
		{
			name:     "simple values",
			values:   []float64{10, 20, 30, 40, 50},
			wantMean: 30.0,
			wantStdDev: 14.142135623730951, // sqrt(200)
		},
		{
			name:     "all same values",
			values:   []float64{100, 100, 100},
			wantMean: 100.0,
			wantStdDev: 0.0,
		},
		{
			name:     "single value",
			values:   []float64{42.5},
			wantMean: 42.5,
			wantStdDev: 0.0,
		},
		{
			name:     "empty values",
			values:   []float64{},
			wantMean: 0.0,
			wantStdDev: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMean, gotStdDev := computeStats(tt.values)

			if math.Abs(gotMean-tt.wantMean) > 0.0001 {
				t.Errorf("computeStats() mean = %v, want %v", gotMean, tt.wantMean)
			}

			if math.Abs(gotStdDev-tt.wantStdDev) > 0.0001 {
				t.Errorf("computeStats() stdDev = %v, want %v", gotStdDev, tt.wantStdDev)
			}
		})
	}
}

func TestGetSeverity(t *testing.T) {
	tests := []struct {
		name   string
		zScore float64
		want   string
	}{
		{"critical positive", 4.5, "CRITICAL"},
		{"critical negative", -4.2, "CRITICAL"},
		{"high positive", 3.5, "HIGH"},
		{"high negative", -3.1, "HIGH"},
		{"medium positive", 2.5, "MEDIUM"},
		{"medium negative", -2.2, "MEDIUM"},
		{"low positive", 1.5, "LOW"},
		{"low negative", -1.0, "LOW"},
		{"zero", 0.0, "LOW"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSeverity(tt.zScore)
			if got != tt.want {
				t.Errorf("getSeverity(%v) = %v, want %v", tt.zScore, got, tt.want)
			}
		})
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  float64
	}{
		{"integer", "42", 42.0},
		{"float", "123.45", 123.45},
		{"zero", "0", 0.0},
		{"negative", "-50.5", -50.5},
		{"invalid", "abc", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseFloat(tt.input)
			if got != tt.want {
				t.Errorf("parseFloat(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
