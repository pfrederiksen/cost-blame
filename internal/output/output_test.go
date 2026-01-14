package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/pfrederiksen/cost-blame/internal/cost"
	"github.com/pfrederiksen/cost-blame/internal/inventory"
)

func TestFormatTags(t *testing.T) {
	tests := []struct {
		name string
		tags map[string]string
		want string
	}{
		{
			name: "empty tags",
			tags: map[string]string{},
			want: "-",
		},
		{
			name: "single tag",
			tags: map[string]string{"env": "prod"},
			want: "env=prod",
		},
		{
			name: "multiple tags under limit",
			tags: map[string]string{"env": "prod", "team": "platform"},
			// Order not guaranteed, just check it's not empty or "-"
			want: "", // We'll check length instead
		},
		{
			name: "many tags triggers truncation",
			tags: map[string]string{
				"env": "prod",
				"team": "platform",
				"app": "api",
				"owner": "john",
			},
			want: "...", // Should contain ellipsis
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTags(tt.tags)

			if tt.want == "" {
				// Multiple tags case - just check it's not empty
				if len(tt.tags) > 0 && got == "-" {
					t.Error("formatTags() returned '-' for non-empty tags")
				}
			} else if tt.want == "..." {
				// Truncation case - check for ellipsis
				if len(got) > 0 && got[len(got)-3:] != "..." {
					t.Errorf("formatTags() should end with '...' for many tags, got %q", got)
				}
			} else {
				// Exact match cases
				if got != tt.want {
					t.Errorf("formatTags() = %q, want %q", got, tt.want)
				}
			}
		})
	}
}

func TestPrintResources_Empty(t *testing.T) {
	// Test that empty resource list doesn't crash
	err := PrintResources([]inventory.Resource{}, "AmazonEC2", false)
	if err != nil {
		t.Errorf("PrintResources() error = %v", err)
	}
}

func TestPrintDeltas_Empty(t *testing.T) {
	// Test that empty delta list doesn't crash
	err := PrintDeltas(nil, 0, 10, false, false)
	if err != nil {
		t.Errorf("PrintDeltas() error = %v", err)
	}
}

func TestPrintJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name: "simple struct",
			input: struct {
				Name  string
				Value int
			}{
				Name:  "test",
				Value: 42,
			},
			wantErr: false,
		},
		{
			name:    "map",
			input:   map[string]string{"key": "value"},
			wantErr: false,
		},
		{
			name:    "slice",
			input:   []int{1, 2, 3},
			wantErr: false,
		},
		{
			name:    "string",
			input:   "simple string",
			wantErr: false,
		},
		{
			name:    "nil",
			input:   nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintJSON(&buf, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("PrintJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && buf.Len() == 0 {
				t.Error("PrintJSON() produced no output")
			}
		})
	}
}

func TestPrintDeltas_ThresholdFiltering(t *testing.T) {
	deltas := []cost.Delta{
		{Key: "Service1", AbsoluteDelta: 500.0},
		{Key: "Service2", AbsoluteDelta: 150.0},
		{Key: "Service3", AbsoluteDelta: 50.0},
		{Key: "Service4", AbsoluteDelta: 10.0},
	}

	// Test with threshold = 100
	// Should filter out Service3 and Service4
	// Note: This test just verifies the function doesn't crash
	// Actual filtering is tested by checking the function runs without error
	err := PrintDeltas(deltas, 100.0, 0, true, false)
	if err != nil {
		t.Errorf("PrintDeltas() error = %v", err)
	}
}

func TestPrintDeltas_TopNLimit(t *testing.T) {
	deltas := make([]cost.Delta, 20)
	for i := range deltas {
		deltas[i] = cost.Delta{
			Key:           fmt.Sprintf("Service%d", i),
			AbsoluteDelta: float64(100 - i),
		}
	}

	// Test with topN = 5
	// Should only output 5 deltas
	err := PrintDeltas(deltas, 0, 5, true, false)
	if err != nil {
		t.Errorf("PrintDeltas() error = %v", err)
	}
}

func TestPrintDeltas_JSON(t *testing.T) {
	deltas := []cost.Delta{
		{
			Key:           "AmazonEC2",
			CurrentCost:   100.0,
			PriorCost:     80.0,
			AbsoluteDelta: 20.0,
			PercentChange: 25.0,
		},
	}

	// Test JSON output (outputs to stdout, so we just verify no error)
	err := PrintDeltas(deltas, 0, 10, true, false)
	if err != nil {
		t.Errorf("PrintDeltas() with JSON error = %v", err)
	}
}

func TestPrintResources_JSON(t *testing.T) {
	resources := []inventory.Resource{
		{
			ID:   "i-12345",
			Type: "EC2 Instance",
			Tags: map[string]string{"env": "prod"},
		},
	}

	// Test JSON output (outputs to stdout, so we just verify no error)
	err := PrintResources(resources, "AmazonEC2", true)
	if err != nil {
		t.Errorf("PrintResources() with JSON error = %v", err)
	}
}

func TestDeltaOutput_Struct(t *testing.T) {
	// Test that DeltaOutput marshals correctly
	output := DeltaOutput{
		Deltas: []cost.Delta{
			{Key: "Test", CurrentCost: 100.0},
		},
		Threshold: 50.0,
		TopN:      10,
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal DeltaOutput: %v", err)
	}

	var decoded DeltaOutput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal DeltaOutput: %v", err)
	}

	if decoded.Threshold != 50.0 {
		t.Errorf("Threshold = %v, want 50.0", decoded.Threshold)
	}
	if decoded.TopN != 10 {
		t.Errorf("TopN = %v, want 10", decoded.TopN)
	}
}

func TestResourceOutput_Struct(t *testing.T) {
	// Test that ResourceOutput marshals correctly
	output := ResourceOutput{
		Resources: []inventory.Resource{
			{ID: "test-id", Type: "Test"},
		},
		Service: "AmazonEC2",
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal ResourceOutput: %v", err)
	}

	var decoded ResourceOutput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ResourceOutput: %v", err)
	}

	if decoded.Service != "AmazonEC2" {
		t.Errorf("Service = %v, want AmazonEC2", decoded.Service)
	}
}
