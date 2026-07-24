package tenant

import "context"

type Repository interface {
	Create(
		ctx context.Context,
		input CreateInput,
	) (Tenant, error)

	Update(
		ctx context.Context,
		id string,
		input UpdateInput,
	) (Tenant, error)

	FindByID(
		ctx context.Context,
		id string,
	) (Tenant, error)

	FindBySlug(
		ctx context.Context,
		slug string,
	) (Tenant, error)

	ListByUserID(
		ctx context.Context,
		userID string,
	) ([]Tenant, error)

	ListAccessByUserID(
		ctx context.Context,
		userID string,
	) ([]Access, error)
}
