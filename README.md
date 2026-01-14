# cost-blame

> Local-first CLI to attribute AWS cost spikes to services, tags, and likely resources.

**cost-blame** helps CloudOps and FinOps teams answer the critical question: **"Who or what caused our AWS spend to rise?"**

Built with Go, it runs entirely locally using AWS Cost Explorer and Resource APIs—no backend required.

## Features

- **Spike Detection**: Identify cost increases across services, regions, accounts, or tags
- **New Spender Identification**: Find resources that just started incurring costs
- **Tag-based Attribution**: Blame cost changes on teams, apps, or environments via tags
- **Resource Drilldown**: Map cost spikes to specific EC2 instances, RDS databases, and more
- **Flexible Output**: Human-readable tables or JSON for automation
- **AWS SDK v2**: Fast, modern AWS integration with proper pagination and rate limiting

## Installation

### From Source

```bash
git clone https://github.com/pfrederiksen/cost-blame.git
cd cost-blame
make build
# Binary will be in ./bin/cost-blame
```

### Requirements

- Go 1.22+
- AWS credentials configured (via `~/.aws/credentials` or environment variables)
- AWS Cost Explorer API access (requires payer account for multi-account visibility)

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

**Flags:**
- `--last`: Time window (`48h`, `7d`, `30d`)
- `--granularity`: `DAILY` or `HOURLY` (default: `DAILY`)
- `--threshold`: Minimum USD delta to report (default: `0`)
- `--group-by`: `service`, `linked_account`, `region`, or `usage_type` (default: `service`)
- `--tag-key`: Optional tag dimension to group by
- `--top`: Number of results (default: `10`)
- `--json`: Output as JSON
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

## Contributing

Contributions welcome! This is an open-source project under the Apache 2.0 license.

## License

Apache License 2.0 - see [LICENSE](LICENSE)

## Roadmap

- Multi-account support via AWS Organizations
- S3 and Lambda-specific drilldown
- Anomaly detection beyond simple deltas
- Export to CSV, Slack, or PagerDuty
- Interactive TUI mode with bubble tea

---

Built with ❤️ for CloudOps teams tired of mystery AWS bills.
