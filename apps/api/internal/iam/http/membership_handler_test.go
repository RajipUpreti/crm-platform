package iamhttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/tenant"
	"github.com/rajipupreti/crm-platform/apps/api/internal/requestcontext"
	"github.com/rajipupreti/crm-platform/apps/api/internal/user"
)

type membershipHandlerRepository struct {
	members            []membership.Member
	target             membership.Member
	updatedRole        membership.Membership
	updatedStatus      membership.Membership
	updatedStatusValue membership.Status
}

func (f *membershipHandlerRepository) Create(
	context.Context,
	membership.CreateInput,
) (membership.Membership, error) {
	return membership.Membership{}, nil
}

func (f *membershipHandlerRepository) FindByID(
	context.Context,
	string,
) (membership.Membership, error) {
	return membership.Membership{}, membership.ErrNotFound
}

func (f *membershipHandlerRepository) FindByTenantAndUser(
	context.Context,
	string,
	string,
) (membership.Membership, error) {
	return membership.Membership{}, membership.ErrNotFound
}

func (f *membershipHandlerRepository) ListByTenantID(
	context.Context,
	string,
) ([]membership.Membership, error) {
	return nil, nil
}

func (f *membershipHandlerRepository) ListByUserID(
	context.Context,
	string,
) ([]membership.Membership, error) {
	return nil, nil
}

func (f *membershipHandlerRepository) ListDetailedByTenantID(
	context.Context,
	string,
) ([]membership.Member, error) {
	return f.members, nil
}

func (f *membershipHandlerRepository) FindDetailedByID(
	context.Context,
	string,
) (membership.Member, error) {
	return f.target, nil
}

func (f *membershipHandlerRepository) CountActiveOwners(
	context.Context,
	string,
) (int, error) {
	return 2, nil
}

func (f *membershipHandlerRepository) UpdateRole(
	context.Context,
	string,
	membership.Role,
) (membership.Membership, error) {
	return membership.Membership{}, nil
}

func (f *membershipHandlerRepository) UpdateRoleForTenant(
	context.Context,
	string,
	string,
	membership.Role,
) (membership.Membership, error) {
	return f.updatedRole, nil
}

func (f *membershipHandlerRepository) UpdateStatus(
	context.Context,
	string,
	membership.Status,
) (membership.Membership, error) {
	return membership.Membership{}, nil
}

func (f *membershipHandlerRepository) UpdateStatusForTenant(
	_ context.Context,
	_ string,
	_ string,
	status membership.Status,
) (membership.Membership, error) {
	f.updatedStatusValue = status
	return f.updatedStatus, nil
}

func TestMembershipHandlerListsDetailedMembers(
	t *testing.T,
) {
	t.Parallel()

	repository := &membershipHandlerRepository{
		members: []membership.Member{
			{
				Membership: membership.DetailedMembership{
					ID:   "membership-id",
					Role: membership.RoleAdmin,
				},
				User: membership.DetailedUser{
					ID:    "user-id",
					Email: "member@example.com",
				},
			},
		},
	}
	handler := newTestMembershipHandler(t, repository)
	request := authenticatedMembershipRequest(
		http.MethodGet,
		"/api/v1/members",
		"",
	)
	response := httptest.NewRecorder()

	handler.ListMembers(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d; expected 200", response.Code)
	}
	if !strings.Contains(
		response.Body.String(),
		`"email":"member@example.com"`,
	) {
		t.Fatalf("response = %s", response.Body.String())
	}
}

func TestMembershipHandlerRejectsOwnerRoleTarget(
	t *testing.T,
) {
	t.Parallel()

	repository := &membershipHandlerRepository{
		target: membership.Member{
			Membership: membership.DetailedMembership{
				ID:       "target-membership",
				TenantID: "tenant-id",
				UserID:   "target-user",
				Role:     membership.RoleMember,
				Status:   membership.StatusActive,
			},
		},
	}
	handler := newTestMembershipHandler(t, repository)
	request := authenticatedMembershipRequest(
		http.MethodPatch,
		"/api/v1/members/target-membership/role",
		`{"role":"OWNER"}`,
	)
	request.SetPathValue("membershipId", "target-membership")
	response := httptest.NewRecorder()

	handler.UpdateMemberRole(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; expected 400", response.Code)
	}
}

func TestMembershipHandlerRemoveMarksMembershipLeft(
	t *testing.T,
) {
	t.Parallel()

	repository := &membershipHandlerRepository{
		target: membership.Member{
			Membership: membership.DetailedMembership{
				ID:       "target-membership",
				TenantID: "tenant-id",
				UserID:   "target-user",
				Role:     membership.RoleMember,
				Status:   membership.StatusActive,
			},
		},
	}
	handler := newTestMembershipHandler(t, repository)
	request := authenticatedMembershipRequest(
		http.MethodDelete,
		"/api/v1/members/target-membership",
		"",
	)
	request.SetPathValue("membershipId", "target-membership")
	response := httptest.NewRecorder()

	handler.RemoveMember(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d; expected 204", response.Code)
	}
	if repository.updatedStatusValue != membership.StatusLeft {
		t.Fatalf(
			"status update = %q; expected LEFT",
			repository.updatedStatusValue,
		)
	}
}

func newTestMembershipHandler(
	t *testing.T,
	repository membership.Repository,
) *MembershipHandler {
	t.Helper()

	service, err := membership.NewService(repository)
	if err != nil {
		t.Fatal(err)
	}

	handler, err := NewMembershipHandler(service)
	if err != nil {
		t.Fatal(err)
	}

	return handler
}

func authenticatedMembershipRequest(
	method string,
	target string,
	body string,
) *http.Request {
	request := httptest.NewRequest(
		method,
		target,
		strings.NewReader(body),
	)

	return request.WithContext(
		requestcontext.WithAuthentication(
			request.Context(),
			requestcontext.Authentication{
				User: user.User{ID: "actor-user"},
				Tenant: tenant.Tenant{
					ID: "tenant-id",
				},
				Membership: membership.Membership{
					Role: membership.RoleOwner,
				},
			},
		),
	)
}
