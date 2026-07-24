package membership

import "context"

type Repository interface {
	Create(
		ctx context.Context,
		input CreateInput,
	) (Membership, error)

	FindByID(
		ctx context.Context,
		id string,
	) (Membership, error)

	FindByTenantAndUser(
		ctx context.Context,
		tenantID string,
		userID string,
	) (Membership, error)

	ListByTenantID(
		ctx context.Context,
		tenantID string,
	) ([]Membership, error)

	ListByUserID(
		ctx context.Context,
		userID string,
	) ([]Membership, error)

	ListDetailedByTenantID(
		ctx context.Context,
		tenantID string,
	) ([]Member, error)

	FindDetailedByID(
		ctx context.Context,
		id string,
	) (Member, error)

	CountActiveOwners(
		ctx context.Context,
		tenantID string,
	) (int, error)

	UpdateRoleForTenant(
		ctx context.Context,
		id string,
		tenantID string,
		role Role,
	) (Membership, error)

	UpdateStatusForTenant(
		ctx context.Context,
		id string,
		tenantID string,
		status Status,
	) (Membership, error)
}
