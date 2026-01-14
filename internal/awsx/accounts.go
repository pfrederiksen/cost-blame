package awsx

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
)

// ListAccounts retrieves all active accounts in the organization
func (c *Clients) ListAccounts(ctx context.Context) ([]string, error) {
	var accountIDs []string

	paginator := organizations.NewListAccountsPaginator(c.Organizations, &organizations.ListAccountsInput{})
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list accounts: %w", err)
		}

		for _, account := range output.Accounts {
			// Only include active accounts
			if account.Status == types.AccountStatusActive {
				accountIDs = append(accountIDs, aws.ToString(account.Id))
			}
		}
	}

	return accountIDs, nil
}

// GetAccountName retrieves the name for a given account ID
func (c *Clients) GetAccountName(ctx context.Context, accountID string) (string, error) {
	output, err := c.Organizations.DescribeAccount(ctx, &organizations.DescribeAccountInput{
		AccountId: aws.String(accountID),
	})
	if err != nil {
		return accountID, nil // Fallback to ID if describe fails
	}

	return aws.ToString(output.Account.Name), nil
}
