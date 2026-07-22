package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/rajipupreti/crm-platform/apps/api/internal/httpresponse"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/tenant"
	"github.com/rajipupreti/crm-platform/apps/api/internal/requestcontext"
	"github.com/rajipupreti/crm-platform/apps/api/internal/session"
	"github.com/rajipupreti/crm-platform/apps/api/internal/user"
)

type SessionReader interface {
	Find(
		ctx context.Context,
		token string,
	) (session.Session, error)
}

type SessionCookieReader interface {
	Read(
		r *http.Request,
	) (string, error)
}

type UserReader interface {
	FindByID(
		ctx context.Context,
		id string,
	) (user.User, error)
}
type TenantReader interface {
	FindByID(
		ctx context.Context,
		id string,
	) (tenant.Tenant, error)
}

type MembershipReader interface {
	FindActiveByTenantAndUser(
		ctx context.Context,
		tenantID string,
		userID string,
	) (membership.Membership, error)
}

type AuthenticationMiddleware struct {
	sessions SessionReader
	cookies  SessionCookieReader
	users    UserReader

	tenants TenantReader

	memberships MembershipReader
}

func NewAuthenticationMiddleware(
	sessions SessionReader,
	cookies SessionCookieReader,
	users UserReader,
	tenants TenantReader,
	memberships MembershipReader,
) (*AuthenticationMiddleware, error) {
	if sessions == nil {
		return nil, errors.New(
			"session reader is required",
		)
	}

	if cookies == nil {
		return nil, errors.New(
			"session cookie reader is required",
		)
	}

	if users == nil {
		return nil, errors.New(
			"user reader is required",
		)
	}

	if tenants == nil {
		return nil, errors.New(
			"tenant reader is required",
		)
	}

	if memberships == nil {
		return nil, errors.New(
			"membership reader is required",
		)
	}

	return &AuthenticationMiddleware{
		sessions:    sessions,
		cookies:     cookies,
		users:       users,
		tenants:     tenants,
		memberships: memberships,
	}, nil
}

func (m *AuthenticationMiddleware) Require(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			token, err := m.cookies.Read(r)
			if err != nil {
				writeUnauthorized(w)
				return
			}

			storedSession, err := m.sessions.Find(
				r.Context(),
				token,
			)
			if err != nil {
				switch {
				case errors.Is(
					err,
					session.ErrNotFound,
				),
					errors.Is(
						err,
						session.ErrExpired,
					),
					errors.Is(
						err,
						session.ErrInvalid,
					):
					writeUnauthorized(w)

				default:
					log.Printf(
						"read application session: %v",
						err,
					)

					httpresponse.Error(
						w,
						http.StatusInternalServerError,
						"session_lookup_failed",
						"could not validate application session",
					)
				}

				return
			}

			currentUser, err := m.users.FindByID(
				r.Context(),
				storedSession.UserID,
			)
			if err != nil {
				switch {
				case errors.Is(
					err,
					user.ErrNotFound,
				):
					writeUnauthorized(w)

				case errors.Is(
					err,
					user.ErrSuspended,
				):
					httpresponse.Error(
						w,
						http.StatusForbidden,
						"user_suspended",
						"user access is suspended",
					)

				default:
					log.Printf(
						"load authenticated user: %v",
						err,
					)

					httpresponse.Error(
						w,
						http.StatusInternalServerError,
						"user_lookup_failed",
						"could not load authenticated user",
					)
				}

				return
			}
			currentTenant, err := m.tenants.FindByID(
				r.Context(),
				storedSession.TenantID,
			)
			if err != nil {
				switch {
				case errors.Is(
					err,
					tenant.ErrNotFound,
				),
					errors.Is(
						err,
						tenant.ErrDeleted,
					):
					writeUnauthorized(w)

				case errors.Is(
					err,
					tenant.ErrSuspended,
				):
					httpresponse.Error(
						w,
						http.StatusForbidden,
						"tenant_suspended",
						"workspace access is suspended",
					)

				default:
					log.Printf(
						"load authenticated tenant: %v",
						err,
					)

					httpresponse.Error(
						w,
						http.StatusInternalServerError,
						"tenant_lookup_failed",
						"could not load current workspace",
					)
				}

				return
			}
			currentMembership, err := m.memberships.
				FindActiveByTenantAndUser(
					r.Context(),
					currentTenant.ID,
					currentUser.ID,
				)
			if err != nil {
				switch {
				case errors.Is(
					err,
					membership.ErrNotFound,
				),
					errors.Is(
						err,
						membership.ErrInactive,
					):
					writeUnauthorized(w)

				case errors.Is(
					err,
					membership.ErrSuspended,
				):
					httpresponse.Error(
						w,
						http.StatusForbidden,
						"membership_suspended",
						"workspace membership is suspended",
					)

				default:
					log.Printf(
						"load authenticated membership: %v",
						err,
					)

					httpresponse.Error(
						w,
						http.StatusInternalServerError,
						"membership_lookup_failed",
						"could not validate workspace membership",
					)
				}

				return
			}
			if currentMembership.ID !=
				storedSession.MembershipID {
				writeUnauthorized(w)
				return
			}
			ctx := requestcontext.WithAuthentication(
				r.Context(),
				requestcontext.Authentication{
					Session: storedSession,

					User: currentUser,

					Tenant: currentTenant,

					Membership: currentMembership,
				},
			)
			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		},
	)
}

func writeUnauthorized(
	w http.ResponseWriter,
) {
	httpresponse.Error(
		w,
		http.StatusUnauthorized,
		"authentication_required",
		"authentication is required",
	)
}
