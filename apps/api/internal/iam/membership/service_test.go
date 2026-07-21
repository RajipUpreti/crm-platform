package membership

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRepository struct {
	createInput  CreateInput
	createResult Membership
	createError  error
}

func (f *fakeRepository) Create(
	ctx context.Context,
	input CreateInput,
) (Membership, error) {
	f.createInput = input
	return f.createResult, f.createError
}

func (f *fakeRepository) FindByID(
	ctx context.Context,
	id string,
) (Membership, error) {
	return Membership{}, ErrNotFound
}

func (f *fakeRepository) FindByTenantAndUser(
	ctx context.Context,
	tenantID string,
	userID string,
) (Membership, error) {
	return Membership{}, ErrNotFound
}

func (f *fakeRepository) ListByTenantID(
	ctx context.Context,
	tenantID string,
) ([]Membership, error) {
	return nil, nil
}

func (f *fakeRepository) ListByUserID(
	ctx context.Context,
	userID string,
) ([]Membership, error) {
	return nil, nil
}

func (f *fakeRepository) UpdateRole(
	ctx context.Context,
	id string,
	role Role,
) (Membership, error) {
	return Membership{}, ErrNotFound
}

func (f *fakeRepository) UpdateStatus(
	ctx context.Context,
	id string,
	status Status,
) (Membership, error) {
	return Membership{}, ErrNotFound
}

func TestServiceCreateActiveMembershipSetsJoinedAt(
	t *testing.T,
) {
	t.Parallel()

	repository := &fakeRepository{
		createResult: Membership{
			ID:       "membership-id",
			TenantID: "tenant-id",
			UserID:   "user-id",
			Role:     RoleOwner,
			Status:   StatusActive,
		},
	}

	service, err := NewService(repository)
	if err != nil {
		t.Fatalf(
			"NewService() error = %v",
			err,
		)
	}

	fixedNow := time.Date(
		2026,
		time.July,
		21,
		1,
		0,
		0,
		0,
		time.UTC,
	)

	service.now = func() time.Time {
		return fixedNow
	}

	_, err = service.Create(
		context.Background(),
		CreateInput{
			TenantID: "tenant-id",
			UserID:   "user-id",
			Role:     RoleOwner,
			Status:   StatusActive,
		},
	)
	if err != nil {
		t.Fatalf(
			"Create() error = %v",
			err,
		)
	}

	if repository.createInput.JoinedAt == nil {
		t.Fatal(
			"joinedAt was not populated",
		)
	}

	if !repository.createInput.JoinedAt.Equal(
		fixedNow,
	) {
		t.Fatalf(
			"joinedAt = %v; expected %v",
			repository.createInput.JoinedAt,
			fixedNow,
		)
	}
}

func TestServiceCreateRejectsInvalidRole(
	t *testing.T,
) {
	t.Parallel()

	service, err := NewService(
		&fakeRepository{},
	)
	if err != nil {
		t.Fatalf(
			"NewService() error = %v",
			err,
		)
	}

	_, err = service.Create(
		context.Background(),
		CreateInput{
			TenantID: "tenant-id",
			UserID:   "user-id",
			Role:     Role("SUPERADMIN"),
			Status:   StatusActive,
		},
	)

	if !errors.Is(
		err,
		ErrInvalidInput,
	) {
		t.Fatalf(
			"error = %v; expected ErrInvalidInput",
			err,
		)
	}
}
