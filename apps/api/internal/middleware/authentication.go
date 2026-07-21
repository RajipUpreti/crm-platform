package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/rajipupreti/crm-platform/apps/api/internal/httpresponse"
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

type AuthenticationMiddleware struct {
	sessions SessionReader
	cookies  SessionCookieReader
	users    UserReader
}

func NewAuthenticationMiddleware(
	sessions SessionReader,
	cookies SessionCookieReader,
	users UserReader,
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

	return &AuthenticationMiddleware{
		sessions: sessions,
		cookies:  cookies,
		users:    users,
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

			storedSession, err :=
				m.sessions.Find(
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

			currentUser, err :=
				m.users.FindByID(
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

			ctx :=
				requestcontext.WithAuthentication(
					r.Context(),
					requestcontext.Authentication{
						Session: storedSession,
						User:    currentUser,
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
