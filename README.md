# cost-blame

[![Test](https://github.com/pfrederiksen/cost-blame/actions/workflows/test.yml/badge.svg)](https://github.com/pfrederiksen/cost-blame/actions/workflows/test.yml)
[![Release](https://img.shields.io/github/v/release/pfrederiksen/cost-blame)](https://github.com/pfrederiksen/cost-blame/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/pfrederiksen/cost-blame)](https://goreportcard.com/report/github.com/pfrederiksen/cost-blame)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

> Local-first CLI to attribute AWS cost spikes to services, tags, and likely resources.

**cost-blame** helps CloudOps and FinOps teams answer the critical question: **"Who or what caused our AWS spend to rise?"**

Built with Go, it runs entirely locally using AWS Cost Explorer and Resource APIs—no backend required.

## Features

- **Spike Detection**: Identify cost increases across services, regions, accounts, or tags
- **Anomaly Detection**: Statistical analysis with z-score for detecting unusual cost patterns
- **New Spender Identification**: Find resources that just started incurring costs
- **Tag-based Attribution**: Blame cost changes on teams, apps, or environments via tags
- **Resource Drilldown**: Map cost spikes to specific EC2, RDS, S3, Lambda, CloudFront, ECS, EKS resources
- **Multi-Account Support**: Query across AWS Organizations or filter specific accounts
- **Export Options**: CSV export and Slack webhook integration for alerts
- **Flexible Output**: Human-readable tables or JSON for automation
- **AWS SDK v2**: Fast, modern AWS integration with proper pagination and rate limiting

## Installation

### Homebrew (Recommended)

```bash
# Add the tap
brew tap pfrederiksen/tap

# Install cost-blame
brew install cost-blame

# Upgrade to latest version
brew upgrade cost-blame
```

### Download Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/pfrederiksen/cost-blame/releases):

**macOS (Intel)**
```bash
curl -L https://github.com/pfrederiksen/cost-blame/releases/latest/download/cost-blame_0.2.2_Darwin_x86_64.tar.gz | tar xz
sudo mv cost-blame /usr/local/bin/
```

**macOS (Apple Silicon)**
```bash
curl -L https://github.com/pfrederiksen/cost-blame/releases/latest/download/cost-blame_0.2.2_Darwin_arm64.tar.gz | tar xz
sudo mv cost-blame /usr/local/bin/
```

**Linux (amd64)**
```bash
curl -L https://github.com/pfrederiksen/cost-blame/releases/latest/download/cost-blame_0.2.2_Linux_x86_64.tar.gz | tar xz
sudo mv cost-blame /usr/local/bin/
```

**Linux (arm64)**
```bash
curl -L https://github.com/pfrederiksen/cost-blame/releases/latest/download/cost-blame_0.2.2_Linux_arm64.tar.gz | tar xz
sudo mv cost-blame /usr/local/bin/
```

**Windows**
Download the `.zip` file from the releases page and extract it to your PATH.

### Using Go

```bash
go install github.com/pfrederiksen/cost-blame@latest
```

### From Source

```bash
git clone https://github.com/pfrederiksen/cost-blame.git
cd cost-blame
make build
# Binary will be in ./bin/cost-blame
```

### Requirements

- AWS credentials configured (via `~/.aws/credentials` or environment variables)
- AWS Cost Explorer API access (requires payer account for multi-account visibility)
- For building from source: Go 1.22+

## Quick Start

### Find top cost increases in the last 7 days

```bash
cost-blame spike --last 7d --threshold 100 --group-by service
```

### Attribute costs by team tag

```bash
cost-blame blame --last 30d --tag-key team
```

### Drill down into EC2 cost spike

```bash
cost-blame drilldown AmazonEC2 --last 48h --region us-west-2 --tag-key team
```

## Commands

### `cost-blame spike`

Detect cost spikes by comparing two equal time periods.

**New in v0.2.0**: Multi-account support, CSV export, Slack webhooks

**Flags:**
- `--last`: Time window (`48h`, `7d`, `30d`)
- `--granularity`: `DAILY` or `HOURLY` (default: `DAILY`)
- `--threshold`: Minimum USD delta to report (default: `0`)
- `--group-by`: `service`, `linked_account`, `region`, or `usage_type` (default: `service`)
- `--tag-key`: Optional tag dimension to group by
- `--accounts`: Filter to specific account IDs (comma-separated)
- `--all-accounts`: Query all accounts in AWS Organization
- `--top`: Number of results (default: `10`)
- `--json`: Output as JSON
- `--csv`: Export results to CSV file
- `--slack-webhook`: Send alerts to Slack webhook URL
- `--profile`: AWS profile
- `--region`: AWS region (default: `us-east-1`)

**Example:**

```bash
cost-blame spike --last 7d --threshold 200 --group-by service --top 5
```

### `cost-blame new-spend`

Find resources that recently started spending.

**Flags:** Same as `spike`, plus:
- `--min-current`: Minimum current spend to consider (default: `50`)

**Example:**

```bash
cost-blame new-spend --last 30d --min-current 100 --group-by service
```

### `cost-blame blame`

Attribute cost changes by tag values.

**Flags:**
- `--tag-key`: Tag to group by (required)
- `--tag-values`: Optional CSV filter for specific values
- Other flags same as `spike`

**Example:**

```bash
cost-blame blame --last 30d --tag-key team --threshold 50
```

### `cost-blame drilldown`

Map a service cost spike to likely resources.

**Arguments:**
- `SERVICE`: AWS service name (e.g., `AmazonEC2`, `AmazonRDS`)

**Flags:**
- `--last`: Time window
- `--region`: Filter to specific region
- `--account`: Filter to specific account ID
- `--tag-key`: Filter by tag key
- `--json`: Output as JSON

**Example:**

```bash
cost-blame drilldown AmazonEC2 --last 48h --region us-west-2
```

## How It Works

1. **Cost Explorer Queries**: Fetches cost data for current and prior periods
2. **Delta Calculation**: Computes absolute and percentage changes
3. **Resource Correlation**: Uses Resource Groups Tagging API and service-specific APIs (EC2, RDS) to identify resources
4. **Attribution**: Maps costs to tags, teams, or specific resource ARNs

## Permissions Required

Minimum IAM permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ce:GetCostAndUsage",
        "tag:GetResources",
        "ec2:DescribeInstances",
        "ec2:DescribeVolumes",
        "ec2:DescribeNatGateways",
        "ec2:DescribeAddresses",
        "ec2:DescribeSnapshots",
        "rds:DescribeDBInstances"
      ],
      "Resource": "*"
    }
  ]
}
```

**Note**: Cost Explorer requires **payer account** credentials for consolidated billing visibility across linked accounts.

## Configuration

cost-blame respects standard AWS SDK configuration:

- `AWS_PROFILE` / `--profile`
- `AWS_REGION` / `--region`
- `~/.aws/credentials` and `~/.aws/config`

## Edge Cases & Limitations

- **Incomplete Data**: Costs for the current day are not final; tool warns when window includes today
- **Hourly Granularity**: Cost Explorer has constraints on hourly data retention
- **Currency**: Assumes USD but displays actual currency from API responses
- **Permissions**: Continues with partial results if some APIs are inaccessible

## Development

```bash
# Build
make build

# Run tests
make test

# Lint (requires golangci-lint)
make lint

# Run locally
go run main.go spike --last 7d --debug
```

## CI/CD

This project uses GitHub Actions for automated testing and releases:

- **Automated Testing**: Every PR and push runs the full test suite
- **Multi-platform Releases**: Automated builds for Linux, macOS, and Windows (amd64 + arm64)
- **Semantic Versioning**: Tag a version (e.g., `v0.1.0`) to trigger an automated release

### Creating a Release

```bash
# Tag the version
git tag v0.1.0

# Push the tag to trigger release workflow
git push origin v0.1.0
```

The release workflow will automatically:
- Run all tests
- Build binaries for all platforms
- Generate checksums
- Create a GitHub release with installation instructions

## Contributing

Contributions welcome! This is an open-source project under the Apache 2.0 license.

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Commit your changes (`git commit -m 'feat: add amazing feature'`)
6. Push to the branch (`git push origin feat/amazing-feature`)
7. Open a Pull Request

All PRs automatically run tests via GitHub Actions.

## License

Apache License 2.0 - see [LICENSE](LICENSE)

## Roadmap

- ✅ Multi-account support via AWS Organizations (v0.2.0)
- ✅ S3, Lambda, CloudFront, ECS, EKS drilldown (v0.2.0)
- ✅ Anomaly detection with statistical analysis (v0.2.0)
- ✅ CSV export and Slack webhook integration (v0.2.0)
- Interactive TUI mode with bubble tea
- PagerDuty integration
- Cost forecasting and trend analysis
- Budget threshold alerts

---

Built with ❤️ for CloudOps teams tired of mystery AWS bills.

### `cost-blame anomaly`

Detect cost anomalies using statistical analysis (z-score).

**Arguments:**
- `--group-by`: `service` or `linked_account` (default: `service`)
- `--historical-days`: Number of days for baseline (default: `30`)
- `--threshold`: Z-score threshold for anomaly detection (default: `2.0`)
- `--min-data-points`: Minimum data points required (default: `7`)
- `--anomalies-only`: Show only detected anomalies
- `--top`: Number of results (default: `20`)
- `--json`: Output as JSON

**Example:**

```bash
cost-blame anomaly --historical-days 30 --threshold 2.5 --anomalies-only
```
