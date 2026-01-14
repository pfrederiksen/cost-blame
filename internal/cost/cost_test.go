package cost

import (
	"testing"
)

func TestComputeDeltas(t *testing.T) {
	current := map[string]float64{
		"AmazonEC2":  500.0,
		"AmazonRDS":  200.0,
		"AmazonS3":   50.0,
		"NewService": 100.0,
	}

	prior := map[string]float64{
		"AmazonEC2": 400.0,
		"AmazonRDS": 200.0,
		"AmazonS3":  100.0,
	}

	deltas := computeDeltas(current, prior)

	if len(deltas) != 4 {
		t.Errorf("Expected 4 deltas, got %d", len(deltas))
	}

	// Find specific deltas
	var ec2, rds, s3, newSvc *Delta
	for i := range deltas {
		switch deltas[i].Key {
		case "AmazonEC2":
			ec2 = &deltas[i]
		case "AmazonRDS":
			rds = &deltas[i]
		case "AmazonS3":
			s3 = &deltas[i]
		case "NewService":
			newSvc = &deltas[i]
		}
	}

	// Test EC2 increase
	if ec2 == nil {
		t.Fatal("EC2 delta not found")
	}
	if ec2.AbsoluteDelta != 100.0 {
		t.Errorf("EC2 delta = %v, want 100.0", ec2.AbsoluteDelta)
	}
	if ec2.PercentChange != 25.0 {
		t.Errorf("EC2 percent = %v, want 25.0", ec2.PercentChange)
	}

	// Test RDS no change
	if rds == nil {
		t.Fatal("RDS delta not found")
	}
	if rds.AbsoluteDelta != 0.0 {
		t.Errorf("RDS delta = %v, want 0.0", rds.AbsoluteDelta)
	}

	// Test S3 decrease
	if s3 == nil {
		t.Fatal("S3 delta not found")
	}
	if s3.AbsoluteDelta != -50.0 {
		t.Errorf("S3 delta = %v, want -50.0", s3.AbsoluteDelta)
	}

	// Test new spender
	if newSvc == nil {
		t.Fatal("NewService delta not found")
	}
	if !newSvc.IsNewSpender {
		t.Error("NewService should be marked as new spender")
	}
	if newSvc.AbsoluteDelta != 100.0 {
		t.Errorf("NewService delta = %v, want 100.0", newSvc.AbsoluteDelta)
	}
}

func TestBuildGroupKey(t *testing.T) {
	tests := []struct {
		name string
		keys []string
		want string
	}{
		{"single key", []string{"AmazonEC2"}, "AmazonEC2"},
		{"multiple keys", []string{"AmazonEC2", "us-east-1"}, "AmazonEC2 | us-east-1"},
		{"empty keys", []string{}, "Unknown"},
		{"three keys", []string{"AmazonEC2", "us-east-1", "production"}, "AmazonEC2 | us-east-1 | production"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildGroupKey(tt.keys)
			if got != tt.want {
				t.Errorf("buildGroupKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComputeDeltas_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		current map[string]float64
		prior   map[string]float64
		wantLen int
		checks  func(t *testing.T, deltas []Delta)
	}{
		{
			name:    "empty maps",
			current: map[string]float64{},
			prior:   map[string]float64{},
			wantLen: 0,
		},
		{
			name:    "only current",
			current: map[string]float64{"AmazonEC2": 100.0},
			prior:   map[string]float64{},
			wantLen: 1,
			checks: func(t *testing.T, deltas []Delta) {
				if !deltas[0].IsNewSpender {
					t.Error("Expected service to be marked as new spender")
				}
				if deltas[0].PercentChange != 9999 {
					t.Errorf("PercentChange = %v, want 9999", deltas[0].PercentChange)
				}
			},
		},
		{
			name:    "only prior",
			current: map[string]float64{},
			prior:   map[string]float64{"AmazonEC2": 100.0},
			wantLen: 1,
			checks: func(t *testing.T, deltas []Delta) {
				if deltas[0].AbsoluteDelta != -100.0 {
					t.Errorf("AbsoluteDelta = %v, want -100.0", deltas[0].AbsoluteDelta)
				}
				if deltas[0].IsNewSpender {
					t.Error("Service that disappeared should not be marked as new spender")
				}
			},
		},
		{
			name:    "very small costs below threshold",
			current: map[string]float64{"AmazonEC2": 0.005},
			prior:   map[string]float64{"AmazonEC2": 0.003},
			wantLen: 1,
			checks: func(t *testing.T, deltas []Delta) {
				if deltas[0].IsNewSpender {
					t.Error("Very small costs should not trigger new spender")
				}
			},
		},
		{
			name:    "crossing new spender threshold",
			current: map[string]float64{"AmazonEC2": 0.02},
			prior:   map[string]float64{"AmazonEC2": 0.005},
			wantLen: 1,
			checks: func(t *testing.T, deltas []Delta) {
				if !deltas[0].IsNewSpender {
					t.Error("Crossing $0.01 threshold should be marked as new spender")
				}
			},
		},
		{
			name:    "negative percentage change",
			current: map[string]float64{"AmazonS3": 50.0},
			prior:   map[string]float64{"AmazonS3": 100.0},
			wantLen: 1,
			checks: func(t *testing.T, deltas []Delta) {
				if deltas[0].PercentChange != -50.0 {
					t.Errorf("PercentChange = %v, want -50.0", deltas[0].PercentChange)
				}
			},
		},
		{
			name:    "zero current and prior",
			current: map[string]float64{"AmazonEC2": 0.0},
			prior:   map[string]float64{"AmazonEC2": 0.0},
			wantLen: 1,
			checks: func(t *testing.T, deltas []Delta) {
				if deltas[0].AbsoluteDelta != 0.0 {
					t.Errorf("AbsoluteDelta = %v, want 0.0", deltas[0].AbsoluteDelta)
				}
				if deltas[0].PercentChange != 0.0 {
					t.Errorf("PercentChange = %v, want 0.0", deltas[0].PercentChange)
				}
				if deltas[0].IsNewSpender {
					t.Error("Zero costs should not be marked as new spender")
				}
			},
		},
		{
			name:    "large percentage increase",
			current: map[string]float64{"AmazonEC2": 1000.0},
			prior:   map[string]float64{"AmazonEC2": 10.0},
			wantLen: 1,
			checks: func(t *testing.T, deltas []Delta) {
				expected := 9900.0 // (990/10)*100
				if deltas[0].PercentChange != expected {
					t.Errorf("PercentChange = %v, want %v", deltas[0].PercentChange, expected)
				}
			},
		},
		{
			name: "mixed scenarios",
			current: map[string]float64{
				"Service1": 100.0, // increase
				"Service2": 50.0,  // decrease
				"Service3": 0.0,   // disappeared
				"Service4": 75.0,  // new
			},
			prior: map[string]float64{
				"Service1": 80.0,
				"Service2": 100.0,
				"Service3": 50.0,
			},
			wantLen: 4,
			checks: func(t *testing.T, deltas []Delta) {
				var service4 *Delta
				for i := range deltas {
					if deltas[i].Key == "Service4" {
						service4 = &deltas[i]
						break
					}
				}
				if service4 == nil {
					t.Fatal("Service4 not found in deltas")
				}
				if !service4.IsNewSpender {
					t.Error("Service4 should be marked as new spender")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deltas := computeDeltas(tt.current, tt.prior)

			if len(deltas) != tt.wantLen {
				t.Errorf("computeDeltas() returned %d deltas, want %d", len(deltas), tt.wantLen)
			}

			if tt.checks != nil && len(deltas) > 0 {
				tt.checks(t, deltas)
			}
		})
	}
}

func TestDeltaStruct(t *testing.T) {
	// Test that Delta struct has all expected fields
	delta := Delta{
		Key:           "AmazonEC2",
		CurrentCost:   100.0,
		PriorCost:     80.0,
		AbsoluteDelta: 20.0,
		PercentChange: 25.0,
		IsNewSpender:  false,
		Currency:      "USD",
	}

	if delta.Key == "" {
		t.Error("Delta Key should not be empty")
	}
	if delta.Currency == "" {
		t.Error("Delta Currency should not be empty")
	}
}

func TestQueryParams_Validation(t *testing.T) {
	// Document QueryParams structure and validation expectations
	tests := []struct {
		name        string
		groupBy     string
		granularity string
		valid       bool
	}{
		{"valid service groupBy", "service", "DAILY", true},
		{"valid linked_account groupBy", "linked_account", "HOURLY", true},
		{"valid region groupBy", "region", "DAILY", true},
		{"valid usage_type groupBy", "usage_type", "DAILY", true},
		{"invalid groupBy", "invalid", "DAILY", false},
		{"empty granularity defaults to DAILY", "service", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Document expected behavior
			t.Logf("GroupBy: %s, Granularity: %s, Valid: %v", tt.groupBy, tt.granularity, tt.valid)

			// Query function should:
			// - Return error for unsupported group-by values
			// - Default to DAILY granularity if not specified or invalid
			// - Accept AccountIDs as optional filter
			// - Accept TagKey for additional grouping dimension
			// - Accept TagValues for filtering specific tag values

			t.Skip("Requires AWS Cost Explorer mocking for full test")
		})
	}
}
