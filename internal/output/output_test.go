package output

import (
	"testing"

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
