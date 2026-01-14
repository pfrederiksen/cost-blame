package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/pfrederiksen/cost-blame/internal/anomaly"
	"github.com/pfrederiksen/cost-blame/internal/awsx"
	"github.com/pfrederiksen/cost-blame/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var anomalyCmd = &cobra.Command{
	Use:   "anomaly",
	Short: "Detect cost anomalies using statistical analysis",
	Long: `Analyze historical cost data to identify anomalies using z-score analysis.
Compares current costs against historical baseline (mean and standard deviation).

Example:
  cost-blame anomaly --historical-days 30 --threshold 2.0 --group-by service`,
	RunE: runAnomaly,
}

func init() {
	rootCmd.AddCommand(anomalyCmd)

	anomalyCmd.Flags().String("group-by", "service", "Group by: service or linked_account")
	anomalyCmd.Flags().Int("historical-days", 30, "Number of days of historical data to analyze")
	anomalyCmd.Flags().Float64("threshold", 2.0, "Z-score threshold for anomaly detection")
	anomalyCmd.Flags().Int("min-data-points", 7, "Minimum data points required")
	anomalyCmd.Flags().Int("top", 20, "Number of results to show")
	anomalyCmd.Flags().Bool("anomalies-only", false, "Show only detected anomalies")
	anomalyCmd.Flags().Bool("json", false, "Output as JSON")
}

func runAnomaly(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	log := getLogger()

	// Parse flags
	groupBy, _ := cmd.Flags().GetString("group-by")
	historicalDays, _ := cmd.Flags().GetInt("historical-days")
	threshold, _ := cmd.Flags().GetFloat64("threshold")
	minDataPoints, _ := cmd.Flags().GetInt("min-data-points")
	topN, _ := cmd.Flags().GetInt("top")
	anomaliesOnly, _ := cmd.Flags().GetBool("anomalies-only")
	asJSON, _ := cmd.Flags().GetBool("json")

	// Create AWS clients
	clients, err := awsx.New(ctx, awsx.Options{
		Profile: viper.GetString("profile"),
		Region:  viper.GetString("region"),
	})
	if err != nil {
		return fmt.Errorf("failed to create AWS clients: %w", err)
	}

	// Detect anomalies
	log.Info("analyzing historical cost data for anomalies...",
		zap.Int("historical_days", historicalDays),
		zap.Float64("z_score_threshold", threshold))

	results, err := anomaly.Detect(ctx, clients.CostExplorer, groupBy, anomaly.DetectorConfig{
		HistoricalDays:  historicalDays,
		ZScoreThreshold: threshold,
		MinDataPoints:   minDataPoints,
	})
	if err != nil {
		return fmt.Errorf("anomaly detection failed: %w", err)
	}

	// Filter to anomalies only if requested
	if anomaliesOnly {
		filtered := make([]anomaly.Anomaly, 0)
		for _, a := range results {
			if a.IsAnomaly {
				filtered = append(filtered, a)
			}
		}
		results = filtered
	}

	// Limit results
	if topN > 0 && len(results) > topN {
		results = results[:topN]
	}

	log.Debug("anomaly detection complete",
		zap.Int("total_results", len(results)),
		zap.Int("anomalies", countAnomalies(results)))

	// Output results
	if asJSON {
		return printAnomaliesJSON(results)
	}
	return printAnomaliesTable(results)
}

func printAnomaliesTable(anomalies []anomaly.Anomaly) error {
	if len(anomalies) == 0 {
		fmt.Println("No anomalies detected")
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Key", "Current", "Mean", "Std Dev", "Z-Score", "Deviation %", "Severity"})
	table.SetBorder(true)
	table.SetAutoWrapText(false)

	for _, a := range anomalies {
		severity := a.Severity
		if !a.IsAnomaly {
			severity = "-"
		}

		table.Append([]string{
			a.Key,
			fmt.Sprintf("$%.2f", a.CurrentCost),
			fmt.Sprintf("$%.2f", a.HistoricalMean),
			fmt.Sprintf("$%.2f", a.HistoricalStdDev),
			fmt.Sprintf("%.2f", a.ZScore),
			fmt.Sprintf("%.1f%%", a.PercentDeviation),
			severity,
		})
	}

	table.Render()
	return nil
}

func printAnomaliesJSON(anomalies []anomaly.Anomaly) error {
	return output.PrintJSON(os.Stdout, map[string]interface{}{
		"anomalies": anomalies,
		"count":     len(anomalies),
	})
}

func countAnomalies(results []anomaly.Anomaly) int {
	count := 0
	for _, r := range results {
		if r.IsAnomaly {
			count++
		}
	}
	return count
}
