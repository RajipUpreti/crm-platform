package user

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRepository struct {
	upsertResult User
	upsertError  error

	receivedIdentity Identity
}

func (f *fakeRepository) UpsertFromIdentity(
	ctx context.Context,
	identity Identity,
) (User, error) {
	f.receivedIdentity = identity
	return f.upsertResult, f.upsertError
}

func (f *fakeRepository) FindByID(
	ctx context.Context,
	id string,
) (User, error) {
	return User{}, ErrNotFound
}

func (f *fakeRepository) FindByProviderIdentity(
	ctx context.Context,
	provider string,
	providerUserID string,
) (User, error) {
	return User{}, ErrNotFound
}

func TestSynchronizeIdentityNormalizesIdentity(
	t *testing.T,
) {
	t.Parallel()

	repository := &fakeRepository{
		upsertResult: User{
			ID:        "user-id",
			Email:     "developer@example.com",
			Status:    StatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	service, err := NewService(repository)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	_, err = service.SynchronizeIdentity(
		context.Background(),
		Identity{
			Provider:       " Keycloak ",
			ProviderUserID: " provider-user-id ",
			Email:          " Developer@Example.COM ",
			EmailVerified:  true,
			FirstName:      " Local ",
			LastName:       " Developer ",
			DisplayName:    " Local Developer ",
		},
	)
	if err != nil {
		t.Fatalf(
			"SynchronizeIdentity() error = %v",
			err,
		)
	}

	if repository.receivedIdentity.Provider != "keycloak" {
		t.Fatalf(
			"provider = %q; expected keycloak",
			repository.receivedIdentity.Provider,
		)
	}

	if repository.receivedIdentity.Email !=
		"developer@example.com" {
		t.Fatalf(
			"email = %q; expected normalized email",
			repository.receivedIdentity.Email,
		)
	}

	if repository.receivedIdentity.FirstName != "Local" {
		t.Fatalf(
			"first name = %q; expected Local",
			repository.receivedIdentity.FirstName,
		)
	}
}

func TestSynchronizeIdentityRejectsUnverifiedEmail(
	t *testing.T,
) {
	t.Parallel()

	repository := &fakeRepository{}

	service, err := NewService(repository)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	_, err = service.SynchronizeIdentity(
		context.Background(),
		Identity{
			Provider:       "keycloak",
			ProviderUserID: "provider-user-id",
			Email:          "developer@example.com",
			EmailVerified:  false,
		},
	)

	if !errors.Is(err, ErrInvalidIdentity) {
		t.Fatalf(
			"error = %v; expected ErrInvalidIdentity",
			err,
		)
	}
}

func TestSynchronizeIdentityRejectsSuspendedUser(
	t *testing.T,
) {
	t.Parallel()

	repository := &fakeRepository{
		upsertResult: User{
			ID:     "user-id",
			Status: StatusSuspended,
		},
	}

	service, err := NewService(repository)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	_, err = service.SynchronizeIdentity(
		context.Background(),
		Identity{
			Provider:       "keycloak",
			ProviderUserID: "provider-user-id",
			Email:          "developer@example.com",
			EmailVerified:  true,
		},
	)

	if !errors.Is(err, ErrSuspended) {
		t.Fatalf(
			"error = %v; expected ErrSuspended",
			err,
		)
	}
}
