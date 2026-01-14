package cmd

import (
	"context"
	"fmt"

	"os"

	"github.com/pfrederiksen/cost-blame/internal/awsx"
	"github.com/pfrederiksen/cost-blame/internal/cost"
	"github.com/pfrederiksen/cost-blame/internal/export"
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
	spikeCmd.Flags().StringSlice("accounts", nil, "Filter to specific account IDs (comma-separated)")
	spikeCmd.Flags().Bool("all-accounts", false, "Query all accounts in organization")
	spikeCmd.Flags().Int("top", 10, "Number of results to show")
	spikeCmd.Flags().Bool("json", false, "Output as JSON")
	spikeCmd.Flags().String("csv", "", "Export to CSV file (path)")
	spikeCmd.Flags().String("slack-webhook", "", "Send alerts to Slack webhook URL")
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
	accounts, _ := cmd.Flags().GetStringSlice("accounts")
	allAccounts, _ := cmd.Flags().GetBool("all-accounts")
	topN, _ := cmd.Flags().GetInt("top")
	asJSON, _ := cmd.Flags().GetBool("json")
	csvPath, _ := cmd.Flags().GetString("csv")
	slackWebhook, _ := cmd.Flags().GetString("slack-webhook")

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

	// Resolve account IDs if --all-accounts is specified
	if allAccounts {
		log.Info("fetching all accounts from organization...")
		accounts, err = clients.ListAccounts(ctx)
		if err != nil {
			log.Warn("failed to list accounts, proceeding without filter", zap.Error(err))
			accounts = nil
		} else {
			log.Info("found accounts", zap.Int("count", len(accounts)))
		}
	}

	// Query cost data
	log.Info("querying cost data...")
	deltas, err := cost.Query(ctx, clients.CostExplorer, cost.QueryParams{
		Window:      window,
		Granularity: granularity,
		GroupBy:     groupBy,
		TagKey:      tagKey,
		AccountIDs:  accounts,
	})
	if err != nil {
		return fmt.Errorf("cost query failed: %w", err)
	}

	// Export to CSV if requested
	if csvPath != "" {
		f, err := os.Create(csvPath)
		if err != nil {
			return fmt.Errorf("failed to create CSV file: %w", err)
		}
		defer f.Close()

		if err := export.WriteCSV(f, deltas); err != nil {
			return fmt.Errorf("failed to write CSV: %w", err)
		}
		log.Info("exported to CSV", zap.String("path", csvPath))
	}

	// Send to Slack if webhook provided
	if slackWebhook != "" {
		if err := export.SendToSlack(slackWebhook, deltas, topN); err != nil {
			log.Warn("failed to send to Slack", zap.Error(err))
		} else {
			log.Info("sent alert to Slack")
		}
	}

	// Output results to console
	return output.PrintDeltas(deltas, threshold, topN, asJSON, window.IncludesToday())
}
