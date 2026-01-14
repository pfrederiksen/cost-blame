package cost

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/pfrederiksen/cost-blame/internal/timewin"
)

// Delta represents cost change between two periods
type Delta struct {
	Key            string  // Group key (service name, tag value, etc.)
	CurrentCost    float64
	PriorCost      float64
	AbsoluteDelta  float64
	PercentChange  float64
	IsNewSpender   bool
	Currency       string
}

// QueryParams holds parameters for Cost Explorer queries
type QueryParams struct {
	Window       *timewin.Window
	Granularity  string   // DAILY or HOURLY
	GroupBy      string   // service, linked_account, region, usage_type
	TagKey       string   // optional tag dimension
	TagValues    []string // optional filter for specific tag values
	AccountIDs   []string // optional filter for specific accounts
}

// Query fetches cost data for current and prior periods and computes deltas
func Query(ctx context.Context, client *costexplorer.Client, params QueryParams) ([]Delta, error) {
	// Build GroupBy dimensions
	var groupDefs []types.GroupDefinition

	switch params.GroupBy {
	case "service":
		groupDefs = append(groupDefs, types.GroupDefinition{
			Type: types.GroupDefinitionType("DIMENSION"),
			Key:  aws.String("SERVICE"),
		})
	case "linked_account":
		groupDefs = append(groupDefs, types.GroupDefinition{
			Type: types.GroupDefinitionType("DIMENSION"),
			Key:  aws.String("LINKED_ACCOUNT"),
		})
	case "region":
		groupDefs = append(groupDefs, types.GroupDefinition{
			Type: types.GroupDefinitionType("DIMENSION"),
			Key:  aws.String("REGION"),
		})
	case "usage_type":
		groupDefs = append(groupDefs, types.GroupDefinition{
			Type: types.GroupDefinitionType("DIMENSION"),
			Key:  aws.String("USAGE_TYPE"),
		})
	default:
		return nil, fmt.Errorf("unsupported group-by: %s", params.GroupBy)
	}

	// Add tag grouping if specified
	if params.TagKey != "" {
		groupDefs = append(groupDefs, types.GroupDefinition{
			Type: types.GroupDefinitionType("TAG"),
			Key:  aws.String(params.TagKey),
		})
	}

	// Determine granularity
	var gran types.Granularity
	switch params.Granularity {
	case "HOURLY":
		gran = types.GranularityHourly
	case "DAILY":
		gran = types.GranularityDaily
	default:
		gran = types.GranularityDaily
	}

	// Query current period
	currentCosts, err := queryCostAndUsage(ctx, client,
		timewin.FormatCE(params.Window.CurrentStart),
		timewin.FormatCE(params.Window.CurrentEnd),
		gran, groupDefs, params.AccountIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query current period: %w", err)
	}

	// Query prior period
	priorCosts, err := queryCostAndUsage(ctx, client,
		timewin.FormatCE(params.Window.PriorStart),
		timewin.FormatCE(params.Window.PriorEnd),
		gran, groupDefs, params.AccountIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query prior period: %w", err)
	}

	// Compute deltas
	deltas := computeDeltas(currentCosts, priorCosts)

	// Sort by absolute delta descending
	sort.Slice(deltas, func(i, j int) bool {
		return deltas[i].AbsoluteDelta > deltas[j].AbsoluteDelta
	})

	return deltas, nil
}

func queryCostAndUsage(ctx context.Context, client *costexplorer.Client, start, end string, gran types.Granularity, groupDefs []types.GroupDefinition, accountIDs []string) (map[string]float64, error) {
	input := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(start),
			End:   aws.String(end),
		},
		Granularity: gran,
		Metrics:     []string{"UnblendedCost"},
		GroupBy:     groupDefs,
	}

	// Add account filter if specified
	if len(accountIDs) > 0 {
		values := make([]string, len(accountIDs))
		copy(values, accountIDs)
		input.Filter = &types.Expression{
			Dimensions: &types.DimensionValues{
				Key:    types.DimensionLinkedAccount,
				Values: values,
			},
		}
	}

	costs := make(map[string]float64)

	// Handle pagination manually
	var nextToken *string
	for {
		if nextToken != nil {
			input.NextPageToken = nextToken
		}

		output, err := client.GetCostAndUsage(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, result := range output.ResultsByTime {
			for _, group := range result.Groups {
				// Build composite key from all group dimensions
				key := buildGroupKey(group.Keys)

				// Sum costs across time periods
				if len(group.Metrics) > 0 {
					if unblended, ok := group.Metrics["UnblendedCost"]; ok && unblended.Amount != nil {
						amount, _ := strconv.ParseFloat(*unblended.Amount, 64)
						costs[key] += amount
					}
				}
			}
		}

		nextToken = output.NextPageToken
		if nextToken == nil {
			break
		}
	}

	return costs, nil
}

func buildGroupKey(keys []string) string {
	if len(keys) == 0 {
		return "Unknown"
	}
	// Join multiple dimensions with " | "
	result := keys[0]
	for i := 1; i < len(keys); i++ {
		result += " | " + keys[i]
	}
	return result
}

func computeDeltas(current, prior map[string]float64) []Delta {
	allKeys := make(map[string]bool)
	for k := range current {
		allKeys[k] = true
	}
	for k := range prior {
		allKeys[k] = true
	}

	var deltas []Delta
	for key := range allKeys {
		curr := current[key]
		prev := prior[key]
		delta := curr - prev

		var pctChange float64
		if prev > 0 {
			pctChange = (delta / prev) * 100
		} else if curr > 0 {
			pctChange = 9999 // Effectively infinite for new spenders
		}

		isNew := prev < 0.01 && curr >= 0.01

		deltas = append(deltas, Delta{
			Key:           key,
			CurrentCost:   curr,
			PriorCost:     prev,
			AbsoluteDelta: delta,
			PercentChange: pctChange,
			IsNewSpender:  isNew,
			Currency:      "USD",
		})
	}

	return deltas
}
