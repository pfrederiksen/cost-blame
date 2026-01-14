package cmd

import (
	"context"
	"fmt"

	"github.com/pfrederiksen/cost-blame/internal/awsx"
	"github.com/pfrederiksen/cost-blame/internal/cost"
	"github.com/pfrederiksen/cost-blame/internal/output"
	"github.com/pfrederiksen/cost-blame/internal/timewin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var spikeCmd = &cobra.Command{
	Use:   "spike",
	Short: "Detect cost spikes between two periods",
	Long: `Compare cost data between current and prior periods of equal length.
Identify services, accounts, or resources with the largest cost increases.

Example:
  cost-blame spike --last 7d --threshold 100 --group-by service --top 10`,
	RunE: runSpike,
}

func init() {
	rootCmd.AddCommand(spikeCmd)

	spikeCmd.Flags().String("last", "7d", "Time window (48h, 7d, 30d)")
	spikeCmd.Flags().String("granularity", "DAILY", "Granularity: DAILY or HOURLY")
	spikeCmd.Flags().Float64("threshold", 0, "Minimum USD delta to report")
	spikeCmd.Flags().String("group-by", "service", "Group by: service, linked_account, region, usage_type")
	spikeCmd.Flags().String("tag-key", "", "Optional tag dimension to group by")
	spikeCmd.Flags().Int("top", 10, "Number of results to show")
	spikeCmd.Flags().Bool("json", false, "Output as JSON")
}

func runSpike(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	log := getLogger()

	// Parse flags
	lastWindow, _ := cmd.Flags().GetString("last")
	granularity, _ := cmd.Flags().GetString("granularity")
	threshold, _ := cmd.Flags().GetFloat64("threshold")
	groupBy, _ := cmd.Flags().GetString("group-by")
	tagKey, _ := cmd.Flags().GetString("tag-key")
	topN, _ := cmd.Flags().GetInt("top")
	asJSON, _ := cmd.Flags().GetBool("json")

	// Parse time window
	window, err := timewin.Parse(lastWindow)
	if err != nil {
		return fmt.Errorf("invalid time window: %w", err)
	}

	log.Debug("parsed time window",
		zap.Time("current_start", window.CurrentStart),
		zap.Time("current_end", window.CurrentEnd),
		zap.Time("prior_start", window.PriorStart),
		zap.Time("prior_end", window.PriorEnd))

	// Create AWS clients
	clients, err := awsx.New(ctx, awsx.Options{
		Profile: viper.GetString("profile"),
		Region:  viper.GetString("region"),
	})
	if err != nil {
		return fmt.Errorf("failed to create AWS clients: %w", err)
	}

	// Query cost data
	log.Info("querying cost data...")
	deltas, err := cost.Query(ctx, clients.CostExplorer, cost.QueryParams{
		Window:      window,
		Granularity: granularity,
		GroupBy:     groupBy,
		TagKey:      tagKey,
	})
	if err != nil {
		return fmt.Errorf("cost query failed: %w", err)
	}

	// Output results
	return output.PrintDeltas(deltas, threshold, topN, asJSON, window.IncludesToday())
}
