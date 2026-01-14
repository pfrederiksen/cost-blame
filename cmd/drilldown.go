package cmd

import (
	"context"
	"fmt"

	"github.com/pfrederiksen/cost-blame/internal/awsx"
	"github.com/pfrederiksen/cost-blame/internal/inventory"
	"github.com/pfrederiksen/cost-blame/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var drilldownCmd = &cobra.Command{
	Use:   "drilldown SERVICE",
	Short: "Map a service cost spike to likely resources",
	Long: `Drill down from a service-level cost spike to identify specific resources
that may be responsible for the cost change.

Supported services: AmazonEC2, AmazonRDS (with fallbacks via Tagging API for others)

Example:
  cost-blame drilldown AmazonEC2 --last 48h --region us-west-2 --tag-key team`,
	Args: cobra.ExactArgs(1),
	RunE: runDrilldown,
}

func init() {
	rootCmd.AddCommand(drilldownCmd)

	drilldownCmd.Flags().String("last", "48h", "Time window (48h, 7d, 30d)")
	drilldownCmd.Flags().String("account", "", "Filter to specific account ID")
	drilldownCmd.Flags().String("tag-key", "", "Filter by tag key")
	drilldownCmd.Flags().Bool("json", false, "Output as JSON")
}

func runDrilldown(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	log := getLogger()

	service := args[0]

	// Parse flags
	region := viper.GetString("region")
	tagKey, _ := cmd.Flags().GetString("tag-key")
	asJSON, _ := cmd.Flags().GetBool("json")

	log.Info("drilling down into service",
		zap.String("service", service),
		zap.String("region", region))

	// Create AWS clients
	clients, err := awsx.New(ctx, awsx.Options{
		Profile: viper.GetString("profile"),
		Region:  region,
	})
	if err != nil {
		return fmt.Errorf("failed to create AWS clients: %w", err)
	}

	// Find resources
	finder := inventory.NewFinder(
		clients.Tagging,
		clients.EC2,
		clients.RDS,
		clients.Lambda,
		clients.S3,
		clients.CloudFront,
		clients.ECS,
		clients.EKS,
	)
	resources, err := finder.FindByService(ctx, service, region, tagKey)
	if err != nil {
		log.Warn("partial results due to error", zap.Error(err))
	}

	log.Debug("found resources", zap.Int("count", len(resources)))

	// Output results
	return output.PrintResources(resources, service, asJSON)
}
