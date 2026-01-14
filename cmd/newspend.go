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

var newSpendCmd = &cobra.Command{
	Use:   "new-spend",
	Short: "Find resources that recently started spending",
	Long: `Identify services or resources that had minimal spend in the prior period
but now have significant costs in the current period.

Example:
  cost-blame new-spend --last 30d --min-current 100 --group-by service`,
	RunE: runNewSpend,
}

func init() {
	rootCmd.AddCommand(newSpendCmd)

	newSpendCmd.Flags().String("last", "30d", "Time window (48h, 7d, 30d)")
	newSpendCmd.Flags().String("granularity", "DAILY", "Granularity: DAILY or HOURLY")
	newSpendCmd.Flags().Float64("min-current", 50, "Minimum current spend to consider")
	newSpendCmd.Flags().String("group-by", "service", "Group by: service, linked_account, region, usage_type")
	newSpendCmd.Flags().String("tag-key", "", "Optional tag dimension to group by")
	newSpendCmd.Flags().Int("top", 20, "Number of results to show")
	newSpendCmd.Flags().Bool("json", false, "Output as JSON")
}

func runNewSpend(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	log := getLogger()

	// Parse flags
	lastWindow, _ := cmd.Flags().GetString("last")
	granularity, _ := cmd.Flags().GetString("granularity")
	minCurrent, _ := cmd.Flags().GetFloat64("min-current")
	groupBy, _ := cmd.Flags().GetString("group-by")
	tagKey, _ := cmd.Flags().GetString("tag-key")
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

	// Query cost data
	log.Info("querying for new spenders...")
	deltas, err := cost.Query(ctx, clients.CostExplorer, cost.QueryParams{
		Window:      window,
		Granularity: granularity,
		GroupBy:     groupBy,
		TagKey:      tagKey,
	})
	if err != nil {
		return fmt.Errorf("cost query failed: %w", err)
	}

	// Filter for new spenders only
	newSpenders := make([]cost.Delta, 0)
	for _, d := range deltas {
		if d.IsNewSpender && d.CurrentCost >= minCurrent {
			newSpenders = append(newSpenders, d)
		}
	}

	log.Debug("found new spenders", zap.Int("count", len(newSpenders)))

	// Output results
	return output.PrintDeltas(newSpenders, 0, topN, asJSON, window.IncludesToday())
}
