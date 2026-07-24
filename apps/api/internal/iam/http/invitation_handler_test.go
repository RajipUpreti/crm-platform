package iamhttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/invitation"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/tenant"
	"github.com/rajipupreti/crm-platform/apps/api/internal/requestcontext"
	"github.com/rajipupreti/crm-platform/apps/api/internal/user"
)

type invitationHandlerRepository struct {
	listResult     []invitation.Invitation
	findResult     invitation.Invitation
	findError      error
	revokeResult   invitation.Invitation
	replaceResult  invitation.Invitation
	receivedStatus *invitation.Status
	receivedTenant string
}

func (f *invitationHandlerRepository) Create(
	context.Context,
	invitation.CreateInput,
) (invitation.Invitation, error) {
	return invitation.Invitation{}, nil
}

func (f *invitationHandlerRepository) FindByTokenDigest(
	context.Context,
	string,
) (invitation.Invitation, error) {
	return invitation.Invitation{}, invitation.ErrNotFound
}

func (f *invitationHandlerRepository) FindByIDAndTenant(
	_ context.Context,
	_ string,
	tenantID string,
) (invitation.Invitation, error) {
	f.receivedTenant = tenantID
	return f.findResult, f.findError
}

func (f *invitationHandlerRepository) ListByTenantID(
	_ context.Context,
	tenantID string,
	status *invitation.Status,
) ([]invitation.Invitation, error) {
	f.receivedTenant = tenantID
	f.receivedStatus = status
	return f.listResult, nil
}

func (f *invitationHandlerRepository) Accept(
	context.Context,
	string,
	string,
) (membership.Membership, error) {
	return membership.Membership{}, nil
}

func (f *invitationHandlerRepository) RevokeForTenant(
	context.Context,
	string,
	string,
) (invitation.Invitation, error) {
	return f.revokeResult, nil
}

func (f *invitationHandlerRepository) ReplaceToken(
	context.Context,
	string,
	string,
	string,
	time.Time,
) (invitation.Invitation, error) {
	return f.replaceResult, nil
}

func TestInvitationHandlerListsWithStatusFilter(
	t *testing.T,
) {
	t.Parallel()

	repository := &invitationHandlerRepository{
		listResult: []invitation.Invitation{
			{
				ID:        "invitation-id",
				Email:     "newuser@example.com",
				Status:    invitation.StatusPending,
				ExpiresAt: time.Now().Add(time.Hour),
			},
		},
	}
	handler := newTestInvitationHandler(t, repository)
	request := authenticatedInvitationRequest(
		http.MethodGet,
		"/api/v1/invitations?status=PENDING",
	)
	response := httptest.NewRecorder()

	handler.ListInvitations(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d; expected 200", response.Code)
	}
	if repository.receivedStatus == nil ||
		*repository.receivedStatus != invitation.StatusPending {
		t.Fatalf(
			"filter = %v; expected PENDING",
			repository.receivedStatus,
		)
	}
	if !strings.Contains(
		response.Body.String(),
		`"email":"newuser@example.com"`,
	) {
		t.Fatalf("response = %s", response.Body.String())
	}
}

func TestInvitationHandlerRejectsInvalidStatusFilter(
	t *testing.T,
) {
	t.Parallel()

	handler := newTestInvitationHandler(
		t,
		&invitationHandlerRepository{},
	)
	request := authenticatedInvitationRequest(
		http.MethodGet,
		"/api/v1/invitations?status=UNKNOWN",
	)
	response := httptest.NewRecorder()

	handler.ListInvitations(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; expected 400", response.Code)
	}
}

func TestInvitationHandlerReturnsNotFoundForOtherTenant(
	t *testing.T,
) {
	t.Parallel()

	repository := &invitationHandlerRepository{
		findError: invitation.ErrNotFound,
	}
	handler := newTestInvitationHandler(t, repository)
	request := authenticatedInvitationRequest(
		http.MethodPost,
		"/api/v1/invitations/other/revoke",
	)
	request.SetPathValue("invitationId", "other")
	response := httptest.NewRecorder()

	handler.RevokeInvitation(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d; expected 404", response.Code)
	}
	if repository.receivedTenant != "tenant-id" {
		t.Fatalf(
			"tenant = %q; expected tenant-id",
			repository.receivedTenant,
		)
	}
}

func newTestInvitationHandler(
	t *testing.T,
	repository invitation.Repository,
) *InvitationHandler {
	t.Helper()

	service, err := invitation.NewService(
		repository,
		7*24*time.Hour,
	)
	if err != nil {
		t.Fatal(err)
	}

	handler, err := NewInvitationHandler(service)
	if err != nil {
		t.Fatal(err)
	}

	return handler
}

func authenticatedInvitationRequest(
	method string,
	target string,
) *http.Request {
	request := httptest.NewRequest(method, target, nil)

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
