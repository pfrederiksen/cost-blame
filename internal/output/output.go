package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/pfrederiksen/cost-blame/internal/cost"
	"github.com/pfrederiksen/cost-blame/internal/inventory"
)

// DeltaOutput formats cost deltas for output
type DeltaOutput struct {
	Deltas    []cost.Delta `json:"deltas"`
	Threshold float64      `json:"threshold,omitempty"`
	TopN      int          `json:"top_n,omitempty"`
}

// ResourceOutput formats resources for output
type ResourceOutput struct {
	Resources []inventory.Resource `json:"resources"`
	Service   string               `json:"service"`
}

// PrintDeltas outputs cost deltas as table or JSON
func PrintDeltas(deltas []cost.Delta, threshold float64, topN int, asJSON bool, includeToday bool) error {
	// Filter by threshold
	filtered := make([]cost.Delta, 0)
	for _, d := range deltas {
		if d.AbsoluteDelta >= threshold {
			filtered = append(filtered, d)
		}
	}

	// Limit to top N
	if topN > 0 && len(filtered) > topN {
		filtered = filtered[:topN]
	}

	if asJSON {
		return printDeltasJSON(filtered, threshold, topN)
	}

	return printDeltasTable(filtered, includeToday)
}

func printDeltasTable(deltas []cost.Delta, includeToday bool) error {
	if includeToday {
		fmt.Fprintf(os.Stderr, "⚠️  Warning: Current period includes today; costs are not final\n\n")
	}

	if len(deltas) == 0 {
		fmt.Println("No cost changes found matching criteria")
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Key", "Current", "Prior", "Delta", "Change %", "New?"})
	table.SetBorder(true)
	table.SetAutoWrapText(false)
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_CENTER,
	})

	for _, d := range deltas {
		newSpender := ""
		if d.IsNewSpender {
			newSpender = "✓"
		}

		pctStr := fmt.Sprintf("%.1f%%", d.PercentChange)
		if d.PercentChange >= 9999 {
			pctStr = "NEW"
		}

		table.Append([]string{
			d.Key,
			fmt.Sprintf("$%.2f", d.CurrentCost),
			fmt.Sprintf("$%.2f", d.PriorCost),
			fmt.Sprintf("$%.2f", d.AbsoluteDelta),
			pctStr,
			newSpender,
		})
	}

	table.Render()
	return nil
}

func printDeltasJSON(deltas []cost.Delta, threshold float64, topN int) error {
	output := DeltaOutput{
		Deltas:    deltas,
		Threshold: threshold,
		TopN:      topN,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// PrintResources outputs resources as table or JSON
func PrintResources(resources []inventory.Resource, service string, asJSON bool) error {
	if asJSON {
		return printResourcesJSON(resources, service)
	}

	return printResourcesTable(resources)
}

func printResourcesTable(resources []inventory.Resource) error {
	if len(resources) == 0 {
		fmt.Println("No resources found matching criteria")
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Type", "Tags"})
	table.SetBorder(true)
	table.SetAutoWrapText(false)

	for _, r := range resources {
		tagsStr := formatTags(r.Tags)
		table.Append([]string{
			r.ID,
			r.Type,
			tagsStr,
		})
	}

	table.Render()
	return nil
}

func printResourcesJSON(resources []inventory.Resource, service string) error {
	output := ResourceOutput{
		Resources: resources,
		Service:   service,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func formatTags(tags map[string]string) string {
	if len(tags) == 0 {
		return "-"
	}

	result := ""
	count := 0
	for k, v := range tags {
		if count > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%s=%s", k, v)
		count++
		if count >= 3 {
			result += "..."
			break
		}
	}
	return result
}

// PrintJSON prints any value as JSON
func PrintJSON(w io.Writer, v interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}
