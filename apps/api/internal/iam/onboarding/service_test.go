package onboarding

import (
	"context"
	"errors"
	"testing"

	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/tenant"
	"github.com/rajipupreti/crm-platform/apps/api/internal/user"
)

type fakeRepository struct {
	findResult Result
	findError  error

	provisionResult Result
	provisionErrors []error

	receivedUserID string
	receivedName   string
	receivedSlugs  []string
}

func (f *fakeRepository) FindPrimaryAccess(
	ctx context.Context,
	userID string,
) (Result, error) {
	f.receivedUserID = userID

	return f.findResult,
		f.findError
}

func (f *fakeRepository) ProvisionPersonalTenant(
	ctx context.Context,
	userID string,
	tenantName string,
	tenantSlug string,
) (Result, error) {
	f.receivedUserID = userID
	f.receivedName = tenantName
	f.receivedSlugs = append(
		f.receivedSlugs,
		tenantSlug,
	)

	index := len(f.receivedSlugs) - 1

	if index < len(f.provisionErrors) &&
		f.provisionErrors[index] != nil {
		return Result{},
			f.provisionErrors[index]
	}

	return f.provisionResult, nil
}

func TestEnsureTenantAccessReturnsExistingAccess(
	t *testing.T,
) {
	t.Parallel()

	expected := Result{
		Tenant: tenant.Tenant{
			ID:     "tenant-id",
			Status: tenant.StatusActive,
		},
		Membership: membership.Membership{
			ID:     "membership-id",
			Role:   membership.RoleOwner,
			Status: membership.StatusActive,
		},
		Created: false,
	}

	repository := &fakeRepository{
		findResult: expected,
	}

	service, err := NewService(repository)
	if err != nil {
		t.Fatalf(
			"NewService() error = %v",
			err,
		)
	}

	result, err := service.EnsureTenantAccess(
		context.Background(),
		user.User{
			ID: "user-id",
		},
	)
	if err != nil {
		t.Fatalf(
			"EnsureTenantAccess() error = %v",
			err,
		)
	}

	if result.Tenant.ID !=
		expected.Tenant.ID {
		t.Fatalf(
			"tenant ID = %q; expected %q",
			result.Tenant.ID,
			expected.Tenant.ID,
		)
	}

	if len(repository.receivedSlugs) != 0 {
		t.Fatal(
			"tenant provisioning should not run",
		)
	}
}

func TestEnsureTenantAccessCreatesPersonalTenant(
	t *testing.T,
) {
	t.Parallel()

	repository := &fakeRepository{
		findError: membership.ErrNotFound,
		provisionResult: Result{
			Tenant: tenant.Tenant{
				ID:   "tenant-id",
				Name: "Local Developer's Workspace",
				Slug: "local-developer-3ea55be2",
			},
			Membership: membership.Membership{
				ID:     "membership-id",
				Role:   membership.RoleOwner,
				Status: membership.StatusActive,
			},
			Created: true,
		},
	}

	service, err := NewService(repository)
	if err != nil {
		t.Fatalf(
			"NewService() error = %v",
			err,
		)
	}

	result, err := service.EnsureTenantAccess(
		context.Background(),
		user.User{
			ID:          "3ea55be2-3e57-4c85-9d19-f6048c405e44",
			Email:       "developer@example.com",
			DisplayName: "Local Developer",
		},
	)
	if err != nil {
		t.Fatalf(
			"EnsureTenantAccess() error = %v",
			err,
		)
	}

	if !result.Created {
		t.Fatal(
			"expected newly created onboarding result",
		)
	}

	if repository.receivedName !=
		"Local Developer's Workspace" {
		t.Fatalf(
			"tenant name = %q",
			repository.receivedName,
		)
	}

	if repository.receivedSlugs[0] !=
		"local-developer-3ea55be2" {
		t.Fatalf(
			"tenant slug = %q",
			repository.receivedSlugs[0],
		)
	}
}

func TestEnsureTenantAccessRetriesSlugCollision(
	t *testing.T,
) {
	t.Parallel()

	repository := &fakeRepository{
		findError: membership.ErrNotFound,
		provisionErrors: []error{
			tenant.ErrSlugAlreadyExists,
			nil,
		},
		provisionResult: Result{
			Created: true,
		},
	}

	service, err := NewService(repository)
	if err != nil {
		t.Fatalf(
			"NewService() error = %v",
			err,
		)
	}

	_, err = service.EnsureTenantAccess(
		context.Background(),
		user.User{
			ID:          "3ea55be2-3e57-4c85-9d19-f6048c405e44",
			DisplayName: "Local Developer",
		},
	)
	if err != nil {
		t.Fatalf(
			"EnsureTenantAccess() error = %v",
			err,
		)
	}

	if len(repository.receivedSlugs) != 2 {
		t.Fatalf(
			"slug attempts = %d; expected 2",
			len(repository.receivedSlugs),
		)
	}

	if repository.receivedSlugs[1] !=
		"local-developer-3ea55be2-2" {
		t.Fatalf(
			"retry slug = %q",
			repository.receivedSlugs[1],
		)
	}
}

func TestEnsureTenantAccessRejectsMissingUserID(
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

	_, err = service.EnsureTenantAccess(
		context.Background(),
		user.User{},
	)

	if !errors.Is(
		err,
		ErrInvalidUser,
	) {
		t.Fatalf(
			"error = %v; expected ErrInvalidUser",
			err,
		)
	}
}
