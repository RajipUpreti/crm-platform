package user

import "context"

type Repository interface {
	UpsertFromIdentity(
		ctx context.Context,
		identity Identity,
	) (User, error)

	FindByID(
		ctx context.Context,
		id string,
	) (User, error)

	FindByProviderIdentity(
		ctx context.Context,
		provider string,
		providerUserID string,
	) (User, error)
}
