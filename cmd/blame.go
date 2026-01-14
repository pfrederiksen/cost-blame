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

var blameCmd = &cobra.Command{
	Use:   "blame",
	Short: "Attribute cost changes by tag values",
	Long: `Group cost deltas by tag values to identify which teams, apps, or
environments are responsible for cost changes.

Example:
  cost-blame blame --last 30d --tag-key team --threshold 50`,
	RunE: runBlame,
}

func init() {
	rootCmd.AddCommand(blameCmd)

	blameCmd.Flags().String("last", "30d", "Time window (48h, 7d, 30d)")
	blameCmd.Flags().String("granularity", "DAILY", "Granularity: DAILY or HOURLY")
	blameCmd.Flags().String("tag-key", "", "Tag key to group by (required)")
	blameCmd.Flags().StringSlice("tag-values", nil, "Optional filter for specific tag values")
	blameCmd.Flags().Float64("threshold", 0, "Minimum USD delta to report")
	blameCmd.Flags().Int("top", 20, "Number of results to show")
	blameCmd.Flags().Bool("json", false, "Output as JSON")

	blameCmd.MarkFlagRequired("tag-key")
}

func runBlame(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	log := getLogger()

	// Parse flags
	lastWindow, _ := cmd.Flags().GetString("last")
	granularity, _ := cmd.Flags().GetString("granularity")
	tagKey, _ := cmd.Flags().GetString("tag-key")
	tagValues, _ := cmd.Flags().GetStringSlice("tag-values")
	threshold, _ := cmd.Flags().GetFloat64("threshold")
	topN, _ := cmd.Flags().GetInt("top")
	asJSON, _ := cmd.Flags().GetBool("json")

	// Parse time window
	window, err := timewin.Parse(lastWindow)
	if err != nil {
		return fmt.Errorf("invalid time window: %w", err)
	}

	// Create AWS clients
	clients, err := awsx.New(ctx, awsx.Options{
		Profile: viper.GetString("profile"),
		Region:  viper.GetString("region"),
	})
	if err != nil {
		return fmt.Errorf("failed to create AWS clients: %w", err)
	}

	// Query cost data grouped by service AND tag
	log.Info("querying cost attribution by tag...", zap.String("tag_key", tagKey))
	deltas, err := cost.Query(ctx, clients.CostExplorer, cost.QueryParams{
		Window:      window,
		Granularity: granularity,
		GroupBy:     "service",
		TagKey:      tagKey,
		TagValues:   tagValues,
	})
	if err != nil {
		return fmt.Errorf("cost query failed: %w", err)
	}

	log.Debug("attribution results", zap.Int("count", len(deltas)))

	// Output results
	return output.PrintDeltas(deltas, threshold, topN, asJSON, window.IncludesToday())
}
