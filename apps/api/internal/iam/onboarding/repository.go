package onboarding

import "context"

type Repository interface {
	FindPrimaryAccess(
		ctx context.Context,
		userID string,
	) (Result, error)

	ProvisionPersonalTenant(
		ctx context.Context,
		userID string,
		tenantName string,
		tenantSlug string,
	) (Result, error)
}
