package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rajipupreti/crm-platform/apps/api/internal/requestcontext"
	"github.com/rajipupreti/crm-platform/apps/api/internal/session"
	"github.com/rajipupreti/crm-platform/apps/api/internal/user"
)

type fakeSessionReader struct {
	result session.Session
	err    error
}

func (f *fakeSessionReader) Find(
	ctx context.Context,
	token string,
) (session.Session, error) {
	return f.result, f.err
}

type fakeCookieReader struct {
	token string
	err   error
}

func (f *fakeCookieReader) Read(
	r *http.Request,
) (string, error) {
	return f.token, f.err
}

type fakeUserReader struct {
	result user.User
	err    error
}

func (f *fakeUserReader) FindByID(
	ctx context.Context,
	id string,
) (user.User, error) {
	return f.result, f.err
}

func TestAuthenticationMiddlewareAllowsValidSession(
	t *testing.T,
) {
	t.Parallel()

	expiresAt := time.Now().
		Add(time.Hour)

	middleware, err := NewAuthenticationMiddleware(
		&fakeSessionReader{
			result: session.Session{
				UserID:    "user-id",
				CreatedAt: time.Now(),
				ExpiresAt: expiresAt,
			},
		},
		&fakeCookieReader{
			token: "session-token",
		},
		&fakeUserReader{
			result: user.User{
				ID:     "user-id",
				Email:  "developer@example.com",
				Status: user.StatusActive,
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"NewAuthenticationMiddleware() error = %v",
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

			authentication, err := requestcontext.
				AuthenticationFromContext(
					r.Context(),
				)
			if err != nil {
				t.Fatalf(
					"authentication context error = %v",
					err,
				)
			}

			if authentication.User.ID !=
				"user-id" {
				t.Fatalf(
					"user ID = %q; expected user-id",
					authentication.User.ID,
				)
			}

			w.WriteHeader(http.StatusOK)
		},
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/auth/me",
		nil,
	)

	response := httptest.NewRecorder()

	middleware.Require(next).
		ServeHTTP(
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
			"status = %d; expected %d",
			response.Code,
			http.StatusOK,
		)
	}
}

func TestAuthenticationMiddlewareRejectsMissingCookie(
	t *testing.T,
) {
	t.Parallel()

	middleware, err := NewAuthenticationMiddleware(
		&fakeSessionReader{},
		&fakeCookieReader{
			err: session.ErrNotFound,
		},
		&fakeUserReader{},
	)
	if err != nil {
		t.Fatalf(
			"NewAuthenticationMiddleware() error = %v",
			err,
		)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/auth/me",
		nil,
	)

	response := httptest.NewRecorder()

	middleware.Require(
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
			"status = %d; expected %d",
			response.Code,
			http.StatusUnauthorized,
		)
	}
}

func TestAuthenticationMiddlewareRejectsExpiredSession(
	t *testing.T,
) {
	t.Parallel()

	middleware, err := NewAuthenticationMiddleware(
		&fakeSessionReader{
			err: session.ErrExpired,
		},
		&fakeCookieReader{
			token: "expired-token",
		},
		&fakeUserReader{},
	)
	if err != nil {
		t.Fatalf(
			"NewAuthenticationMiddleware() error = %v",
			err,
		)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/auth/me",
		nil,
	)

	response := httptest.NewRecorder()

	middleware.Require(
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
			"status = %d; expected %d",
			response.Code,
			http.StatusUnauthorized,
		)
	}
}

func TestAuthenticationMiddlewareRejectsSuspendedUser(
	t *testing.T,
) {
	t.Parallel()

	middleware, err := NewAuthenticationMiddleware(
		&fakeSessionReader{
			result: session.Session{
				UserID: "user-id",
				ExpiresAt: time.Now().
					Add(time.Hour),
			},
		},
		&fakeCookieReader{
			token: "session-token",
		},
		&fakeUserReader{
			err: user.ErrSuspended,
		},
	)
	if err != nil {
		t.Fatalf(
			"NewAuthenticationMiddleware() error = %v",
			err,
		)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/auth/me",
		nil,
	)

	response := httptest.NewRecorder()

	middleware.Require(
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
			"status = %d; expected %d",
			response.Code,
			http.StatusForbidden,
		)
	}
}

func TestAuthenticationMiddlewareHandlesStoreFailure(
	t *testing.T,
) {
	t.Parallel()

	middleware, err := NewAuthenticationMiddleware(
		&fakeSessionReader{
			err: errors.New(
				"Redis unavailable",
			),
		},
		&fakeCookieReader{
			token: "session-token",
		},
		&fakeUserReader{},
	)
	if err != nil {
		t.Fatalf(
			"NewAuthenticationMiddleware() error = %v",
			err,
		)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/auth/me",
		nil,
	)

	response := httptest.NewRecorder()

	middleware.Require(
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
			"status = %d; expected %d",
			response.Code,
			http.StatusInternalServerError,
		)
	}
}
