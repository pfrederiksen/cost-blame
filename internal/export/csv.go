package export

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/pfrederiksen/cost-blame/internal/cost"
)

// WriteCSV exports cost deltas to CSV format
func WriteCSV(w io.Writer, deltas []cost.Delta) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{"Key", "Current Cost", "Prior Cost", "Absolute Delta", "Percent Change", "New Spender", "Currency"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, d := range deltas {
		newSpender := "No"
		if d.IsNewSpender {
			newSpender = "Yes"
		}

		row := []string{
			d.Key,
			fmt.Sprintf("%.2f", d.CurrentCost),
			fmt.Sprintf("%.2f", d.PriorCost),
			fmt.Sprintf("%.2f", d.AbsoluteDelta),
			fmt.Sprintf("%.2f", d.PercentChange),
			newSpender,
			d.Currency,
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}
