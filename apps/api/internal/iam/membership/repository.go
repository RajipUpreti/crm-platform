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

	UpdateRole(
		ctx context.Context,
		id string,
		role Role,
	) (Membership, error)

	UpdateStatus(
		ctx context.Context,
		id string,
		status Status,
	) (Membership, error)
}
