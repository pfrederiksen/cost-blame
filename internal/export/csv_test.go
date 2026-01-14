package export

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pfrederiksen/cost-blame/internal/cost"
)

func TestWriteCSV(t *testing.T) {
	deltas := []cost.Delta{
		{
			Key:           "AmazonEC2",
			CurrentCost:   500.0,
			PriorCost:     400.0,
			AbsoluteDelta: 100.0,
			PercentChange: 25.0,
			IsNewSpender:  false,
			Currency:      "USD",
		},
		{
			Key:           "NewService",
			CurrentCost:   100.0,
			PriorCost:     0.0,
			AbsoluteDelta: 100.0,
			PercentChange: 9999.0,
			IsNewSpender:  true,
			Currency:      "USD",
		},
	}

	var buf bytes.Buffer
	err := WriteCSV(&buf, deltas)
	if err != nil {
		t.Fatalf("WriteCSV() error = %v", err)
	}

	output := buf.String()

	// Check header
	if !strings.Contains(output, "Key,Current Cost,Prior Cost") {
		t.Error("CSV header missing or incorrect")
	}

	// Check data rows
	if !strings.Contains(output, "AmazonEC2,500.00,400.00,100.00,25.00,No,USD") {
		t.Error("AmazonEC2 row missing or incorrect")
	}

	if !strings.Contains(output, "NewService,100.00,0.00,100.00,9999.00,Yes,USD") {
		t.Error("NewService row missing or incorrect")
	}

	// Check row count (header + 2 data rows)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}
}

func TestWriteCSV_Empty(t *testing.T) {
	var buf bytes.Buffer
	err := WriteCSV(&buf, []cost.Delta{})
	if err != nil {
		t.Fatalf("WriteCSV() error = %v", err)
	}

	output := buf.String()

	// Should still have header
	if !strings.Contains(output, "Key,Current Cost") {
		t.Error("CSV header missing for empty data")
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 1 {
		t.Errorf("Expected 1 line (header only), got %d", len(lines))
	}
}

func TestWriteCSV_NegativeValues(t *testing.T) {
	deltas := []cost.Delta{
		{
			Key:           "AmazonS3",
			CurrentCost:   50.0,
			PriorCost:     100.0,
			AbsoluteDelta: -50.0,
			PercentChange: -50.0,
			IsNewSpender:  false,
			Currency:      "USD",
		},
	}

	var buf bytes.Buffer
	err := WriteCSV(&buf, deltas)
	if err != nil {
		t.Fatalf("WriteCSV() error = %v", err)
	}

	output := buf.String()

	// Check negative values are formatted correctly
	if !strings.Contains(output, "-50.00") {
		t.Error("Negative delta not formatted correctly")
	}
}

func TestWriteCSV_SpecialCharactersInKey(t *testing.T) {
	deltas := []cost.Delta{
		{
			Key:           "Service with, comma",
			CurrentCost:   100.0,
			PriorCost:     80.0,
			AbsoluteDelta: 20.0,
			PercentChange: 25.0,
			IsNewSpender:  false,
			Currency:      "USD",
		},
	}

	var buf bytes.Buffer
	err := WriteCSV(&buf, deltas)
	if err != nil {
		t.Fatalf("WriteCSV() error = %v", err)
	}

	output := buf.String()

	// CSV should handle special characters (quotes)
	if !strings.Contains(output, "Service with, comma") && !strings.Contains(output, "\"Service with, comma\"") {
		t.Error("Special characters in key not handled properly")
	}
}

func TestWriteCSV_LargeDataset(t *testing.T) {
	// Test with many deltas
	deltas := make([]cost.Delta, 100)
	for i := range deltas {
		deltas[i] = cost.Delta{
			Key:           string(rune('A' + i%26)),
			CurrentCost:   float64(i * 10),
			PriorCost:     float64(i * 5),
			AbsoluteDelta: float64(i * 5),
			PercentChange: 100.0,
			IsNewSpender:  false,
			Currency:      "USD",
		}
	}

	var buf bytes.Buffer
	err := WriteCSV(&buf, deltas)
	if err != nil {
		t.Fatalf("WriteCSV() error = %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 101 { // header + 100 rows
		t.Errorf("Expected 101 lines, got %d", len(lines))
	}
}
