package tenantswitch

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/tenant"
	"github.com/rajipupreti/crm-platform/apps/api/internal/session"
)

type fakeTenantReader struct {
	result tenant.Tenant
	err    error
}

func (f *fakeTenantReader) FindByID(
	ctx context.Context,
	id string,
) (tenant.Tenant, error) {
	return f.result, f.err
}

type fakeMembershipReader struct {
	result membership.Membership
	err    error
}

func (f *fakeMembershipReader) FindActiveByTenantAndUser(
	ctx context.Context,
	tenantID string,
	userID string,
) (membership.Membership, error) {
	return f.result, f.err
}

type fakeSessionManager struct {
	createInput  session.CreateInput
	createResult session.CreatedSession
	createError  error

	deletedTokens []string
	deleteErrors  map[string]error
}

func (f *fakeSessionManager) Create(
	ctx context.Context,
	input session.CreateInput,
) (session.CreatedSession, error) {
	f.createInput = input
	return f.createResult, f.createError
}

func (f *fakeSessionManager) Delete(
	ctx context.Context,
	token string,
) error {
	f.deletedTokens = append(
		f.deletedTokens,
		token,
	)

	if f.deleteErrors == nil {
		return nil
	}

	return f.deleteErrors[token]
}

func TestSwitchCreatesNewSessionAndDeletesOldSession(
	t *testing.T,
) {
	t.Parallel()

	sessionManager := &fakeSessionManager{
		createResult: session.CreatedSession{
			Token: "new-token",
			ExpiresAt: time.Now().
				Add(time.Hour),
		},
	}

	service, err := NewService(
		&fakeTenantReader{
			result: tenant.Tenant{
				ID:     "target-tenant",
				Status: tenant.StatusActive,
			},
		},
		&fakeMembershipReader{
			result: membership.Membership{
				ID:       "target-membership",
				TenantID: "target-tenant",
				UserID:   "user-id",
				Role:     membership.RoleMember,
				Status:   membership.StatusActive,
			},
		},
		sessionManager,
		sessionManager,
	)
	if err != nil {
		t.Fatalf(
			"NewService() error = %v",
			err,
		)
	}

	result, err := service.Switch(
		context.Background(),
		"user-id",
		"old-token",
		"target-tenant",
	)
	if err != nil {
		t.Fatalf(
			"Switch() error = %v",
			err,
		)
	}

	if sessionManager.createInput.UserID !=
		"user-id" {
		t.Fatalf(
			"user ID = %q",
			sessionManager.createInput.UserID,
		)
	}

	if sessionManager.createInput.TenantID !=
		"target-tenant" {
		t.Fatalf(
			"tenant ID = %q",
			sessionManager.createInput.TenantID,
		)
	}

	if sessionManager.createInput.MembershipID !=
		"target-membership" {
		t.Fatalf(
			"membership ID = %q",
			sessionManager.createInput.MembershipID,
		)
	}

	if len(sessionManager.deletedTokens) != 1 ||
		sessionManager.deletedTokens[0] !=
			"old-token" {
		t.Fatalf(
			"deleted tokens = %v",
			sessionManager.deletedTokens,
		)
	}

	if result.Session.Token != "new-token" {
		t.Fatalf(
			"new token = %q",
			result.Session.Token,
		)
	}
}

func TestSwitchRejectsMissingMembership(
	t *testing.T,
) {
	t.Parallel()

	sessionManager := &fakeSessionManager{}

	service, err := NewService(
		&fakeTenantReader{
			result: tenant.Tenant{
				ID:     "target-tenant",
				Status: tenant.StatusActive,
			},
		},
		&fakeMembershipReader{
			err: membership.ErrNotFound,
		},
		sessionManager,
		sessionManager,
	)
	if err != nil {
		t.Fatalf(
			"NewService() error = %v",
			err,
		)
	}

	_, err = service.Switch(
		context.Background(),
		"user-id",
		"old-token",
		"target-tenant",
	)

	if !errors.Is(
		err,
		membership.ErrNotFound,
	) {
		t.Fatalf(
			"error = %v; expected ErrNotFound",
			err,
		)
	}

	if sessionManager.createInput.UserID != "" {
		t.Fatal(
			"new session should not be created",
		)
	}
}

func TestSwitchDeletesNewSessionWhenOldDeletionFails(
	t *testing.T,
) {
	t.Parallel()

	sessionManager := &fakeSessionManager{
		createResult: session.CreatedSession{
			Token: "new-token",
		},
		deleteErrors: map[string]error{
			"old-token": errors.New(
				"Redis unavailable",
			),
		},
	}

	service, err := NewService(
		&fakeTenantReader{
			result: tenant.Tenant{
				ID:     "target-tenant",
				Status: tenant.StatusActive,
			},
		},
		&fakeMembershipReader{
			result: membership.Membership{
				ID:       "target-membership",
				TenantID: "target-tenant",
				UserID:   "user-id",
				Status:   membership.StatusActive,
			},
		},
		sessionManager,
		sessionManager,
	)
	if err != nil {
		t.Fatalf(
			"NewService() error = %v",
			err,
		)
	}

	_, err = service.Switch(
		context.Background(),
		"user-id",
		"old-token",
		"target-tenant",
	)
	if err == nil {
		t.Fatal(
			"expected tenant switch error",
		)
	}

	if len(sessionManager.deletedTokens) != 2 {
		t.Fatalf(
			"deleted tokens = %v; expected old and new token",
			sessionManager.deletedTokens,
		)
	}

	if sessionManager.deletedTokens[1] !=
		"new-token" {
		t.Fatalf(
			"cleanup token = %q; expected new-token",
			sessionManager.deletedTokens[1],
		)
	}
}
