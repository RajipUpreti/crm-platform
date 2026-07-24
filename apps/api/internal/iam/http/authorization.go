package iamhttp

import (
	"context"
	"errors"
	"net/http"

	"github.com/rajipupreti/crm-platform/apps/api/internal/httpresponse"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/authorization"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/permission"
	"github.com/rajipupreti/crm-platform/apps/api/internal/requestcontext"
)

type PermissionAuthorizer interface {
	Authorize(
		role membership.Role,
		requiredPermission permission.Permission,
	) error
}

type AuthorizationGuard struct {
	authorizer PermissionAuthorizer
}

func NewAuthorizationGuard(
	authorizer PermissionAuthorizer,
) (*AuthorizationGuard, error) {
	if authorizer == nil {
		return nil, errors.New(
			"permission authorizer is required",
		)
	}

	return &AuthorizationGuard{
		authorizer: authorizer,
	}, nil
}

func (g *AuthorizationGuard) Require(
	requiredPermission permission.Permission,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			authentication, err := requestcontext.
				AuthenticationFromContext(
					r.Context(),
				)
			if err != nil {
				httpresponse.Error(
					w,
					http.StatusUnauthorized,
					"authentication_required",
					"authentication is required",
				)
				return
			}

			err = g.authorizer.Authorize(
				authentication.
					Membership.
					Role,
				requiredPermission,
			)
			if err != nil {
				if errors.Is(
					err,
					authorization.
						ErrPermissionDenied,
				) {
					httpresponse.Error(
						w,
						http.StatusForbidden,
						"permission_denied",
						"you do not have permission to perform this action",
					)
					return
				}

				httpresponse.Error(
					w,
					http.StatusInternalServerError,
					"authorization_failed",
					"could not evaluate access permission",
				)
				return
			}

			next.ServeHTTP(
				w,
				r,
			)
		},
	)
}

func AuthorizationFromContext(
	ctx context.Context,
) (
	requestcontext.Authentication,
	error,
) {
	return requestcontext.
		AuthenticationFromContext(ctx)
}
