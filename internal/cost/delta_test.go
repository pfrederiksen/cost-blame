package cost

import (
	"testing"
)

func TestDeltaPercentChange(t *testing.T) {
	tests := []struct {
		name          string
		currentCost   float64
		priorCost     float64
		wantPercent   float64
		wantNewSpender bool
	}{
		{
			name:          "50% increase",
			currentCost:   150.0,
			priorCost:     100.0,
			wantPercent:   50.0,
			wantNewSpender: false,
		},
		{
			name:          "100% increase (doubling)",
			currentCost:   200.0,
			priorCost:     100.0,
			wantPercent:   100.0,
			wantNewSpender: false,
		},
		{
			name:          "50% decrease",
			currentCost:   50.0,
			priorCost:     100.0,
			wantPercent:   -50.0,
			wantNewSpender: false,
		},
		{
			name:          "new spender (zero prior)",
			currentCost:   100.0,
			priorCost:     0.0,
			wantPercent:   9999.0,
			wantNewSpender: true,
		},
		{
			name:          "near-zero prior",
			currentCost:   100.0,
			priorCost:     0.001,
			wantPercent:   9999900.0, // (100 - 0.001) / 0.001 * 100
			wantNewSpender: true, // New spender since prior < 0.01
		},
		{
			name:          "no change",
			currentCost:   100.0,
			priorCost:     100.0,
			wantPercent:   0.0,
			wantNewSpender: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := map[string]float64{"test": tt.currentCost}
			prior := map[string]float64{"test": tt.priorCost}

			deltas := computeDeltas(current, prior)

			if len(deltas) != 1 {
				t.Fatalf("Expected 1 delta, got %d", len(deltas))
			}

			d := deltas[0]

			if d.PercentChange != tt.wantPercent {
				t.Errorf("PercentChange = %.1f, want %.1f", d.PercentChange, tt.wantPercent)
			}

			if d.IsNewSpender != tt.wantNewSpender {
				t.Errorf("IsNewSpender = %v, want %v", d.IsNewSpender, tt.wantNewSpender)
			}
		})
	}
}

func TestBuildGroupKeyMultipleDimensions(t *testing.T) {
	tests := []struct {
		name string
		keys []string
		want string
	}{
		{
			name: "service and account",
			keys: []string{"AmazonEC2", "123456789012"},
			want: "AmazonEC2 | 123456789012",
		},
		{
			name: "service, region, and tag",
			keys: []string{"AmazonEC2", "us-east-1", "team-platform"},
			want: "AmazonEC2 | us-east-1 | team-platform",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildGroupKey(tt.keys)
			if got != tt.want {
				t.Errorf("buildGroupKey() = %q, want %q", got, tt.want)
			}
		})
	}
}
