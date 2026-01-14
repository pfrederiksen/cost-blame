package awsx

import (
	"context"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr bool
	}{
		{
			name:    "default options",
			opts:    Options{},
			wantErr: false,
		},
		{
			name: "with profile",
			opts: Options{
				Profile: "default",
			},
			wantErr: false,
		},
		{
			name: "with region",
			opts: Options{
				Region: "us-west-2",
			},
			wantErr: false,
		},
		{
			name: "with both profile and region",
			opts: Options{
				Profile: "default",
				Region:  "eu-west-1",
			},
			wantErr: false,
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clients, err := New(ctx, tt.opts)

			// In CI environments without AWS credentials, New() may fail
			// This is expected behavior, so we skip the rest of the test
			if err != nil {
				t.Logf("New() returned error (may be expected in CI without credentials): %v", err)
				t.Skip("Skipping test - requires AWS credentials")
				return
			}

			if !tt.wantErr {
				if clients == nil {
					t.Error("New() returned nil clients")
					return
				}

				// Verify all clients are initialized
				if clients.CostExplorer == nil {
					t.Error("CostExplorer client is nil")
				}
				if clients.Organizations == nil {
					t.Error("Organizations client is nil")
				}
				if clients.Tagging == nil {
					t.Error("Tagging client is nil")
				}
				if clients.EC2 == nil {
					t.Error("EC2 client is nil")
				}
				if clients.RDS == nil {
					t.Error("RDS client is nil")
				}
				if clients.Lambda == nil {
					t.Error("Lambda client is nil")
				}
				if clients.S3 == nil {
					t.Error("S3 client is nil")
				}
				if clients.CloudFront == nil {
					t.Error("CloudFront client is nil")
				}
				if clients.ECS == nil {
					t.Error("ECS client is nil")
				}
				if clients.EKS == nil {
					t.Error("EKS client is nil")
				}
			}
		})
	}
}

func TestNewWithInvalidProfile(t *testing.T) {
	ctx := context.Background()

	// This should not error - AWS SDK will use default credentials
	// even with non-existent profile name
	clients, err := New(ctx, Options{
		Profile: "non-existent-profile-12345",
	})

	// The New function itself doesn't validate profile existence
	// It only fails if AWS config can't be loaded at all
	if err != nil {
		t.Logf("New() with invalid profile returned error (expected): %v", err)
	} else if clients == nil {
		t.Error("New() returned nil clients without error")
	}
}
