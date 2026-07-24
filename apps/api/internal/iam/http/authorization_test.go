package iamhttp

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/authorization"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/permission"
	"github.com/rajipupreti/crm-platform/apps/api/internal/requestcontext"
)

type fakePermissionAuthorizer struct {
	receivedRole       membership.Role
	receivedPermission permission.Permission
	err                error
}

func (f *fakePermissionAuthorizer) Authorize(
	role membership.Role,
	requiredPermission permission.Permission,
) error {
	f.receivedRole = role
	f.receivedPermission = requiredPermission

	return f.err
}

func TestAuthorizationGuardAllowsPermission(
	t *testing.T,
) {
	t.Parallel()

	authorizer := &fakePermissionAuthorizer{}

	guard, err := NewAuthorizationGuard(
		authorizer,
	)
	if err != nil {
		t.Fatalf(
			"NewAuthorizationGuard() error = %v",
			err,
		)
	}

	nextCalled := false

	next := http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		},
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/tenant/invitations",
		nil,
	)

	request = request.WithContext(
		requestcontext.WithAuthentication(
			request.Context(),
			requestcontext.Authentication{
				Membership: membership.Membership{
					Role: membership.RoleAdmin,
				},
			},
		),
	)

	response := httptest.NewRecorder()

	guard.Require(
		permission.MemberInvite,
		next,
	).ServeHTTP(
		response,
		request,
	)

	if !nextCalled {
		t.Fatal(
			"next handler was not called",
		)
	}

	if response.Code != http.StatusOK {
		t.Fatalf(
			"status = %d; expected 200",
			response.Code,
		)
	}

	if authorizer.receivedRole !=
		membership.RoleAdmin {
		t.Fatalf(
			"role = %q; expected ADMIN",
			authorizer.receivedRole,
		)
	}
}

func TestAuthorizationGuardReturnsForbidden(
	t *testing.T,
) {
	t.Parallel()

	guard, err := NewAuthorizationGuard(
		&fakePermissionAuthorizer{
			err: authorization.
				ErrPermissionDenied,
		},
	)
	if err != nil {
		t.Fatalf(
			"NewAuthorizationGuard() error = %v",
			err,
		)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/tenant/invitations",
		nil,
	)

	request = request.WithContext(
		requestcontext.WithAuthentication(
			request.Context(),
			requestcontext.Authentication{
				Membership: membership.Membership{
					Role: membership.RoleMember,
				},
			},
		),
	)

	response := httptest.NewRecorder()

	guard.Require(
		permission.MemberInvite,
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				t.Fatal(
					"next handler should not run",
				)
			},
		),
	).ServeHTTP(
		response,
		request,
	)

	if response.Code !=
		http.StatusForbidden {
		t.Fatalf(
			"status = %d; expected 403",
			response.Code,
		)
	}
}

func TestAuthorizationGuardReturnsUnauthorizedWithoutContext(
	t *testing.T,
) {
	t.Parallel()

	guard, err := NewAuthorizationGuard(
		&fakePermissionAuthorizer{},
	)
	if err != nil {
		t.Fatalf(
			"NewAuthorizationGuard() error = %v",
			err,
		)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/tenant/invitations",
		nil,
	)

	response := httptest.NewRecorder()

	guard.Require(
		permission.MemberInvite,
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				t.Fatal(
					"next handler should not run",
				)
			},
		),
	).ServeHTTP(
		response,
		request,
	)

	if response.Code !=
		http.StatusUnauthorized {
		t.Fatalf(
			"status = %d; expected 401",
			response.Code,
		)
	}
}

func TestAuthorizationGuardHandlesUnexpectedFailure(
	t *testing.T,
) {
	t.Parallel()

	guard, err := NewAuthorizationGuard(
		&fakePermissionAuthorizer{
			err: errors.New(
				"unexpected failure",
			),
		},
	)
	if err != nil {
		t.Fatalf(
			"NewAuthorizationGuard() error = %v",
			err,
		)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/tenant/invitations",
		nil,
	)

	request = request.WithContext(
		requestcontext.WithAuthentication(
			request.Context(),
			requestcontext.Authentication{
				Membership: membership.Membership{
					Role: membership.RoleAdmin,
				},
			},
		),
	)

	response := httptest.NewRecorder()

	guard.Require(
		permission.MemberInvite,
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				t.Fatal(
					"next handler should not run",
				)
			},
		),
	).ServeHTTP(
		response,
		request,
	)

	if response.Code !=
		http.StatusInternalServerError {
		t.Fatalf(
			"status = %d; expected 500",
			response.Code,
		)
	}
}
