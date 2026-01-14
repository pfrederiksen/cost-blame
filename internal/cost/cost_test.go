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
