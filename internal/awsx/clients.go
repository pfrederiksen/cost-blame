package awsx

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
)

// Clients holds AWS SDK v2 service clients
type Clients struct {
	CostExplorer *costexplorer.Client
	Tagging      *resourcegroupstaggingapi.Client
	EC2          *ec2.Client
	RDS          *rds.Client
	Config       aws.Config
}

// Options for AWS client configuration
type Options struct {
	Profile string
	Region  string
}

// New creates AWS clients with the given options
func New(ctx context.Context, opts Options) (*Clients, error) {
	var configOpts []func(*config.LoadOptions) error

	if opts.Profile != "" {
		configOpts = append(configOpts, config.WithSharedConfigProfile(opts.Profile))
	}

	if opts.Region != "" {
		configOpts = append(configOpts, config.WithRegion(opts.Region))
	}

	cfg, err := config.LoadDefaultConfig(ctx, configOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Clients{
		CostExplorer: costexplorer.NewFromConfig(cfg),
		Tagging:      resourcegroupstaggingapi.NewFromConfig(cfg),
		EC2:          ec2.NewFromConfig(cfg),
		RDS:          rds.NewFromConfig(cfg),
		Config:       cfg,
	}, nil
}
