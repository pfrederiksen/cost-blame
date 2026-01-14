package anomaly

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/pfrederiksen/cost-blame/internal/timewin"
)

// Anomaly represents a detected cost anomaly
type Anomaly struct {
	Key              string
	CurrentCost      float64
	HistoricalMean   float64
	HistoricalStdDev float64
	ZScore           float64
	PercentDeviation float64
	IsAnomaly        bool
	Severity         string // LOW, MEDIUM, HIGH, CRITICAL
}

// DetectorConfig holds configuration for anomaly detection
type DetectorConfig struct {
	HistoricalDays int     // Number of days of historical data to analyze
	ZScoreThreshold float64 // Z-score threshold for anomaly (default: 2.0)
	MinDataPoints  int     // Minimum data points required
}

// Detect identifies cost anomalies using statistical analysis
func Detect(ctx context.Context, client *costexplorer.Client, groupBy string, config DetectorConfig) ([]Anomaly, error) {
	// Set defaults
	if config.HistoricalDays == 0 {
		config.HistoricalDays = 30
	}
	if config.ZScoreThreshold == 0 {
		config.ZScoreThreshold = 2.0
	}
	if config.MinDataPoints == 0 {
		config.MinDataPoints = 7
	}

	// Query historical cost data (last N days)
	endDate := time.Now().UTC().Truncate(24 * time.Hour)
	startDate := endDate.Add(-time.Duration(config.HistoricalDays) * 24 * time.Hour)

	// Build group definition
	var groupDef types.GroupDefinition
	switch groupBy {
	case "service":
		groupDef = types.GroupDefinition{
			Type: types.GroupDefinitionType("DIMENSION"),
			Key:  aws.String("SERVICE"),
		}
	case "linked_account":
		groupDef = types.GroupDefinition{
			Type: types.GroupDefinitionType("DIMENSION"),
			Key:  aws.String("LINKED_ACCOUNT"),
		}
	default:
		return nil, fmt.Errorf("unsupported group-by for anomaly detection: %s", groupBy)
	}

	input := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(timewin.FormatCE(startDate)),
			End:   aws.String(timewin.FormatCE(endDate)),
		},
		Granularity: types.GranularityDaily,
		Metrics:     []string{"UnblendedCost"},
		GroupBy:     []types.GroupDefinition{groupDef},
	}

	// Fetch all historical data
	historicalData := make(map[string][]float64)
	var nextToken *string

	for {
		if nextToken != nil {
			input.NextPageToken = nextToken
		}

		output, err := client.GetCostAndUsage(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch historical costs: %w", err)
		}

		// Process each time period
		for _, result := range output.ResultsByTime {
			for _, group := range result.Groups {
				key := group.Keys[0]
				if len(group.Metrics) > 0 {
					if unblended, ok := group.Metrics["UnblendedCost"]; ok && unblended.Amount != nil {
						cost := parseFloat(aws.ToString(unblended.Amount))
						historicalData[key] = append(historicalData[key], cost)
					}
				}
			}
		}

		nextToken = output.NextPageToken
		if nextToken == nil {
			break
		}
	}

	// Compute anomalies
	var anomalies []Anomaly
	for key, costs := range historicalData {
		if len(costs) < config.MinDataPoints {
			continue
		}

		// Current cost is the most recent data point
		currentCost := costs[len(costs)-1]

		// Historical baseline is all but the last data point
		historical := costs[:len(costs)-1]
		mean, stdDev := computeStats(historical)

		// Calculate z-score
		zScore := 0.0
		if stdDev > 0 {
			zScore = (currentCost - mean) / stdDev
		}

		percentDeviation := 0.0
		if mean > 0 {
			percentDeviation = ((currentCost - mean) / mean) * 100
		}

		isAnomaly := math.Abs(zScore) >= config.ZScoreThreshold
		severity := getSeverity(zScore)

		anomalies = append(anomalies, Anomaly{
			Key:              key,
			CurrentCost:      currentCost,
			HistoricalMean:   mean,
			HistoricalStdDev: stdDev,
			ZScore:           zScore,
			PercentDeviation: percentDeviation,
			IsAnomaly:        isAnomaly,
			Severity:         severity,
		})
	}

	// Sort by z-score (absolute value, descending)
	sort.Slice(anomalies, func(i, j int) bool {
		return math.Abs(anomalies[i].ZScore) > math.Abs(anomalies[j].ZScore)
	})

	return anomalies, nil
}

func computeStats(values []float64) (mean, stdDev float64) {
	if len(values) == 0 {
		return 0, 0
	}

	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean = sum / float64(len(values))

	// Calculate standard deviation
	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(values))
	stdDev = math.Sqrt(variance)

	return mean, stdDev
}

func getSeverity(zScore float64) string {
	absZ := math.Abs(zScore)
	switch {
	case absZ >= 4.0:
		return "CRITICAL"
	case absZ >= 3.0:
		return "HIGH"
	case absZ >= 2.0:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
