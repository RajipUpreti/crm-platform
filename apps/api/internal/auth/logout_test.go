package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rajipupreti/crm-platform/apps/api/internal/session"
)

type fakeSessionDestroyer struct {
	deletedToken string
	err          error
}

func (f *fakeSessionDestroyer) Delete(
	ctx context.Context,
	token string,
) error {
	f.deletedToken = token
	return f.err
}

func TestLogoutDeletesSessionAndClearsCookie(
	t *testing.T,
) {
	t.Parallel()

	destroyer := &fakeSessionDestroyer{}

	cookieManager, err := session.NewCookieManager(
		session.CookieConfig{
			Name:     "crm_session",
			Path:     "/",
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		},
	)
	if err != nil {
		t.Fatalf(
			"NewCookieManager() error = %v",
			err,
		)
	}

	handler := &Handler{
		sessionDestroyer:     destroyer,
		sessionCookieManager: cookieManager,
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/auth/logout",
		nil,
	)

	request.AddCookie(
		&http.Cookie{
			Name:  "crm_session",
			Value: "raw-session-token",
		},
	)

	response := httptest.NewRecorder()

	handler.Logout(
		response,
		request,
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"status = %d; expected %d",
			response.Code,
			http.StatusOK,
		)
	}

	if destroyer.deletedToken !=
		"raw-session-token" {
		t.Fatalf(
			"deleted token = %q",
			destroyer.deletedToken,
		)
	}

	cookies := response.Result().Cookies()

	if len(cookies) == 0 {
		t.Fatal(
			"logout did not clear the cookie",
		)
	}

	if cookies[0].MaxAge != -1 {
		t.Fatalf(
			"cookie MaxAge = %d; expected -1",
			cookies[0].MaxAge,
		)
	}
}
