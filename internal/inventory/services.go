package inventory

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (f *Finder) findLambdaResources(ctx context.Context, region string) ([]Resource, error) {
	var resources []Resource

	paginator := lambda.NewListFunctionsPaginator(f.lambdaClient, &lambda.ListFunctionsInput{})
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list Lambda functions: %w", err)
		}

		for _, function := range output.Functions {
			// Fetch tags for this function
			tags := make(map[string]string)
			if function.FunctionArn != nil {
				tagResp, err := f.lambdaClient.ListTags(ctx, &lambda.ListTagsInput{
					Resource: function.FunctionArn,
				})
				if err == nil && tagResp.Tags != nil {
					tags = tagResp.Tags
				}
			}

			resources = append(resources, Resource{
				ARN:    aws.ToString(function.FunctionArn),
				ID:     aws.ToString(function.FunctionName),
				Type:   "Lambda Function",
				Tags:   tags,
				Region: region,
			})
		}
	}

	return resources, nil
}

func (f *Finder) findS3Resources(ctx context.Context) ([]Resource, error) {
	var resources []Resource

	output, err := f.s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list S3 buckets: %w", err)
	}

	for _, bucket := range output.Buckets {
		bucketName := aws.ToString(bucket.Name)

		// Fetch tags for this bucket
		tags := make(map[string]string)
		tagResp, err := f.s3Client.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{
			Bucket: bucket.Name,
		})
		if err == nil {
			for _, tag := range tagResp.TagSet {
				tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
			}
		}

		// Get bucket region
		locationResp, err := f.s3Client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
			Bucket: bucket.Name,
		})
		bucketRegion := "us-east-1" // Default
		if err == nil && locationResp.LocationConstraint != "" {
			bucketRegion = string(locationResp.LocationConstraint)
		}

		resources = append(resources, Resource{
			ARN:    fmt.Sprintf("arn:aws:s3:::%s", bucketName),
			ID:     bucketName,
			Type:   "S3 Bucket",
			Tags:   tags,
			Region: bucketRegion,
		})
	}

	return resources, nil
}

func (f *Finder) findCloudFrontResources(ctx context.Context) ([]Resource, error) {
	var resources []Resource

	paginator := cloudfront.NewListDistributionsPaginator(f.cloudfrontClient, &cloudfront.ListDistributionsInput{})
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list CloudFront distributions: %w", err)
		}

		if output.DistributionList == nil {
			continue
		}

		for _, dist := range output.DistributionList.Items {
			// Fetch tags for this distribution
			tags := make(map[string]string)
			if dist.ARN != nil {
				tagResp, err := f.cloudfrontClient.ListTagsForResource(ctx, &cloudfront.ListTagsForResourceInput{
					Resource: dist.ARN,
				})
				if err == nil && tagResp.Tags != nil {
					for _, tag := range tagResp.Tags.Items {
						tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
					}
				}
			}

			resources = append(resources, Resource{
				ARN:    aws.ToString(dist.ARN),
				ID:     aws.ToString(dist.Id),
				Type:   "CloudFront Distribution",
				Tags:   tags,
				Region: "global",
			})
		}
	}

	return resources, nil
}

func (f *Finder) findECSResources(ctx context.Context, region string) ([]Resource, error) {
	var resources []Resource

	// List clusters
	clusterPaginator := ecs.NewListClustersPaginator(f.ecsClient, &ecs.ListClustersInput{})
	for clusterPaginator.HasMorePages() {
		output, err := clusterPaginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list ECS clusters: %w", err)
		}

		if len(output.ClusterArns) == 0 {
			continue
		}

		// Describe clusters to get details
		describeResp, err := f.ecsClient.DescribeClusters(ctx, &ecs.DescribeClustersInput{
			Clusters: output.ClusterArns,
			Include:  []ecsTypes.ClusterField{ecsTypes.ClusterFieldTags},
		})
		if err != nil {
			continue
		}

		for _, cluster := range describeResp.Clusters {
			tags := make(map[string]string)
			for _, tag := range cluster.Tags {
				tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
			}

			resources = append(resources, Resource{
				ARN:    aws.ToString(cluster.ClusterArn),
				ID:     aws.ToString(cluster.ClusterName),
				Type:   "ECS Cluster",
				Tags:   tags,
				Region: region,
			})
		}
	}

	return resources, nil
}

func (f *Finder) findEKSResources(ctx context.Context, region string) ([]Resource, error) {
	var resources []Resource

	paginator := eks.NewListClustersPaginator(f.eksClient, &eks.ListClustersInput{})
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list EKS clusters: %w", err)
		}

		for _, clusterName := range output.Clusters {
			// Describe cluster to get ARN and tags
			describeResp, err := f.eksClient.DescribeCluster(ctx, &eks.DescribeClusterInput{
				Name: aws.String(clusterName),
			})
			if err != nil {
				continue
			}

			tags := make(map[string]string)
			if describeResp.Cluster != nil && describeResp.Cluster.Tags != nil {
				tags = describeResp.Cluster.Tags
			}

			resources = append(resources, Resource{
				ARN:    aws.ToString(describeResp.Cluster.Arn),
				ID:     clusterName,
				Type:   "EKS Cluster",
				Tags:   tags,
				Region: region,
			})
		}
	}

	return resources, nil
}
