package inventory

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	tagtypes "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
)

// Resource represents an AWS resource with tags
type Resource struct {
	ARN     string
	ID      string
	Type    string
	Tags    map[string]string
	Region  string
	Account string
}

// Finder helps locate resources for cost attribution
type Finder struct {
	taggingClient *resourcegroupstaggingapi.Client
	ec2Client     *ec2.Client
	rdsClient     *rds.Client
}

// NewFinder creates a new resource finder
func NewFinder(tagging *resourcegroupstaggingapi.Client, ec2 *ec2.Client, rds *rds.Client) *Finder {
	return &Finder{
		taggingClient: tagging,
		ec2Client:     ec2,
		rdsClient:     rds,
	}
}

// FindByService finds resources for a given service
func (f *Finder) FindByService(ctx context.Context, service, region, tagKey string) ([]Resource, error) {
	// Normalize service name
	normalizedService := strings.ToLower(service)

	switch {
	case strings.Contains(normalizedService, "ec2"):
		return f.findEC2Resources(ctx, region, tagKey)
	case strings.Contains(normalizedService, "rds"):
		return f.findRDSResources(ctx, region, tagKey)
	default:
		// Try generic tagging API
		return f.findViaTaggingAPI(ctx, service, tagKey)
	}
}

func (f *Finder) findEC2Resources(ctx context.Context, region, tagKey string) ([]Resource, error) {
	var resources []Resource

	// Describe instances
	instancesResp, err := f.ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
	if err == nil {
		for _, reservation := range instancesResp.Reservations {
			for _, instance := range reservation.Instances {
				tags := extractEC2Tags(instance.Tags)
				resources = append(resources, Resource{
					ARN:    fmt.Sprintf("arn:aws:ec2:%s::instance/%s", region, aws.ToString(instance.InstanceId)),
					ID:     aws.ToString(instance.InstanceId),
					Type:   "EC2 Instance",
					Tags:   tags,
					Region: region,
				})
			}
		}
	}

	// Describe volumes
	volumesResp, err := f.ec2Client.DescribeVolumes(ctx, &ec2.DescribeVolumesInput{})
	if err == nil {
		for _, volume := range volumesResp.Volumes {
			tags := extractEC2Tags(volume.Tags)
			resources = append(resources, Resource{
				ARN:    fmt.Sprintf("arn:aws:ec2:%s::volume/%s", region, aws.ToString(volume.VolumeId)),
				ID:     aws.ToString(volume.VolumeId),
				Type:   "EBS Volume",
				Tags:   tags,
				Region: region,
			})
		}
	}

	// Describe NAT Gateways
	natResp, err := f.ec2Client.DescribeNatGateways(ctx, &ec2.DescribeNatGatewaysInput{})
	if err == nil {
		for _, nat := range natResp.NatGateways {
			tags := extractEC2Tags(nat.Tags)
			resources = append(resources, Resource{
				ARN:    fmt.Sprintf("arn:aws:ec2:%s::natgateway/%s", region, aws.ToString(nat.NatGatewayId)),
				ID:     aws.ToString(nat.NatGatewayId),
				Type:   "NAT Gateway",
				Tags:   tags,
				Region: region,
			})
		}
	}

	return resources, nil
}

func (f *Finder) findRDSResources(ctx context.Context, region, tagKey string) ([]Resource, error) {
	var resources []Resource

	resp, err := f.rdsClient.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe RDS instances: %w", err)
	}

	for _, instance := range resp.DBInstances {
		tags := make(map[string]string)
		if instance.DBInstanceArn != nil {
			// Fetch tags for this instance
			tagResp, err := f.rdsClient.ListTagsForResource(ctx, &rds.ListTagsForResourceInput{
				ResourceName: instance.DBInstanceArn,
			})
			if err == nil {
				for _, tag := range tagResp.TagList {
					tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
				}
			}
		}

		resources = append(resources, Resource{
			ARN:    aws.ToString(instance.DBInstanceArn),
			ID:     aws.ToString(instance.DBInstanceIdentifier),
			Type:   "RDS Instance",
			Tags:   tags,
			Region: region,
		})
	}

	return resources, nil
}

func (f *Finder) findViaTaggingAPI(ctx context.Context, service, tagKey string) ([]Resource, error) {
	var resources []Resource

	input := &resourcegroupstaggingapi.GetResourcesInput{
		ResourceTypeFilters: []string{service},
	}

	if tagKey != "" {
		input.TagFilters = []tagtypes.TagFilter{
			{
				Key: aws.String(tagKey),
			},
		}
	}

	paginator := resourcegroupstaggingapi.NewGetResourcesPaginator(f.taggingClient, input)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get resources: %w", err)
		}

		for _, resource := range output.ResourceTagMappingList {
			tags := make(map[string]string)
			for _, tag := range resource.Tags {
				tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
			}

			resources = append(resources, Resource{
				ARN:  aws.ToString(resource.ResourceARN),
				Tags: tags,
			})
		}
	}

	return resources, nil
}

func extractEC2Tags(tags []ec2types.Tag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		result[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	return result
}
