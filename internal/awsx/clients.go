package awsx

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Clients holds AWS SDK v2 service clients
type Clients struct {
	CostExplorer  *costexplorer.Client
	Organizations *organizations.Client
	Tagging       *resourcegroupstaggingapi.Client
	EC2           *ec2.Client
	RDS           *rds.Client
	Lambda        *lambda.Client
	S3            *s3.Client
	CloudFront    *cloudfront.Client
	ECS           *ecs.Client
	EKS           *eks.Client
	Config        aws.Config
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
		CostExplorer:  costexplorer.NewFromConfig(cfg),
		Organizations: organizations.NewFromConfig(cfg),
		Tagging:       resourcegroupstaggingapi.NewFromConfig(cfg),
		EC2:           ec2.NewFromConfig(cfg),
		RDS:           rds.NewFromConfig(cfg),
		Lambda:        lambda.NewFromConfig(cfg),
		S3:            s3.NewFromConfig(cfg),
		CloudFront:    cloudfront.NewFromConfig(cfg),
		ECS:           ecs.NewFromConfig(cfg),
		EKS:           eks.NewFromConfig(cfg),
		Config:        cfg,
	}, nil
}
