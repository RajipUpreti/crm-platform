package membership

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRepository struct {
	createInput          CreateInput
	createResult         Membership
	createError          error
	detailedResult       Member
	detailedError        error
	detailedList         []Member
	ownerCount           int
	ownerCountError      error
	updateRoleResult     Membership
	updateRoleError      error
	updateRoleTenantID   string
	updateRoleID         string
	updateRoleValue      Role
	updateStatusResult   Membership
	updateStatusError    error
	updateStatusTenantID string
	updateStatusID       string
	updateStatusValue    Status
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

func (f *fakeRepository) ListDetailedByTenantID(
	ctx context.Context,
	tenantID string,
) ([]Member, error) {
	return f.detailedList, nil
}

func (f *fakeRepository) FindDetailedByID(
	ctx context.Context,
	id string,
) (Member, error) {
	return f.detailedResult, f.detailedError
}

func (f *fakeRepository) CountActiveOwners(
	ctx context.Context,
	tenantID string,
) (int, error) {
	return f.ownerCount, f.ownerCountError
}

func (f *fakeRepository) UpdateRole(
	ctx context.Context,
	id string,
	role Role,
) (Membership, error) {
	return Membership{}, ErrNotFound
}

func (f *fakeRepository) UpdateRoleForTenant(
	ctx context.Context,
	id string,
	tenantID string,
	role Role,
) (Membership, error) {
	f.updateRoleID = id
	f.updateRoleTenantID = tenantID
	f.updateRoleValue = role
	return f.updateRoleResult, f.updateRoleError
}

func (f *fakeRepository) UpdateStatus(
	ctx context.Context,
	id string,
	status Status,
) (Membership, error) {
	return Membership{}, ErrNotFound
}

func (f *fakeRepository) UpdateStatusForTenant(
	ctx context.Context,
	id string,
	tenantID string,
	status Status,
) (Membership, error) {
	f.updateStatusID = id
	f.updateStatusTenantID = tenantID
	f.updateStatusValue = status
	return f.updateStatusResult, f.updateStatusError
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

func TestChangeRoleForTenantUsesTenantScopedUpdate(
	t *testing.T,
) {
	t.Parallel()

	repository := &fakeRepository{
		detailedResult: Member{
			Membership: DetailedMembership{
				ID:       "target-membership",
				TenantID: "tenant-id",
				UserID:   "target-user",
				Role:     RoleMember,
				Status:   StatusActive,
			},
		},
		updateRoleResult: Membership{
			ID:       "target-membership",
			TenantID: "tenant-id",
			Role:     RoleAdmin,
		},
	}
	service, err := NewService(repository)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.ChangeRoleForTenant(
		context.Background(),
		"tenant-id",
		"owner-user",
		RoleOwner,
		"target-membership",
		RoleAdmin,
	)
	if err != nil {
		t.Fatalf("ChangeRoleForTenant() error = %v", err)
	}

	if repository.updateRoleTenantID != "tenant-id" ||
		repository.updateRoleID != "target-membership" ||
		repository.updateRoleValue != RoleAdmin {
		t.Fatalf(
			"tenant-scoped update = (%q, %q, %q)",
			repository.updateRoleTenantID,
			repository.updateRoleID,
			repository.updateRoleValue,
		)
	}
}

func TestChangeRoleForTenantRejectsCrossTenantMembership(
	t *testing.T,
) {
	t.Parallel()

	repository := &fakeRepository{
		detailedResult: Member{
			Membership: DetailedMembership{
				ID:       "target-membership",
				TenantID: "other-tenant",
				UserID:   "target-user",
				Role:     RoleMember,
			},
		},
	}
	service, _ := NewService(repository)

	_, err := service.ChangeRoleForTenant(
		context.Background(),
		"current-tenant",
		"owner-user",
		RoleOwner,
		"target-membership",
		RoleAdmin,
	)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v; expected ErrNotFound", err)
	}
	if repository.updateRoleID != "" {
		t.Fatal("cross-tenant membership was updated")
	}
}

func TestAdminCannotModifyOwner(
	t *testing.T,
) {
	t.Parallel()

	repository := &fakeRepository{
		detailedResult: Member{
			Membership: DetailedMembership{
				ID:       "owner-membership",
				TenantID: "tenant-id",
				UserID:   "owner-user",
				Role:     RoleOwner,
				Status:   StatusActive,
			},
		},
	}
	service, _ := NewService(repository)

	err := service.RemoveForTenant(
		context.Background(),
		"tenant-id",
		"admin-user",
		RoleAdmin,
		"owner-membership",
	)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("error = %v; expected ErrForbidden", err)
	}
}

func TestCannotRemoveFinalActiveOwner(
	t *testing.T,
) {
	t.Parallel()

	repository := &fakeRepository{
		detailedResult: Member{
			Membership: DetailedMembership{
				ID:       "owner-membership",
				TenantID: "tenant-id",
				UserID:   "other-owner",
				Role:     RoleOwner,
				Status:   StatusActive,
			},
		},
		ownerCount: 1,
	}
	service, _ := NewService(repository)

	err := service.RemoveForTenant(
		context.Background(),
		"tenant-id",
		"owner-user",
		RoleOwner,
		"owner-membership",
	)
	if !errors.Is(err, ErrLastOwner) {
		t.Fatalf("error = %v; expected ErrLastOwner", err)
	}
}

func TestCannotRemoveSelfThroughAdminEndpoint(
	t *testing.T,
) {
	t.Parallel()

	repository := &fakeRepository{
		detailedResult: Member{
			Membership: DetailedMembership{
				ID:       "self-membership",
				TenantID: "tenant-id",
				UserID:   "actor-user",
				Role:     RoleAdmin,
				Status:   StatusActive,
			},
		},
	}
	service, _ := NewService(repository)

	err := service.RemoveForTenant(
		context.Background(),
		"tenant-id",
		"actor-user",
		RoleAdmin,
		"self-membership",
	)
	if !errors.Is(err, ErrSelfModification) {
		t.Fatalf(
			"error = %v; expected ErrSelfModification",
			err,
		)
	}
}
