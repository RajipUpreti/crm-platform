package auth

import (
	"net/http"
	"time"

	"github.com/rajipupreti/crm-platform/apps/api/internal/httpresponse"
	"github.com/rajipupreti/crm-platform/apps/api/internal/requestcontext"
	"github.com/rajipupreti/crm-platform/apps/api/internal/user"
)

type CurrentUserResponse struct {
	User      user.User `json:"user"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func (h *Handler) Me(
	w http.ResponseWriter,
	r *http.Request,
) {
	authentication, err := requestcontext.AuthenticationFromContext(
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

	httpresponse.JSON(
		w,
		http.StatusOK,
		CurrentUserResponse{
			User: authentication.User,
			ExpiresAt: authentication.
				Session.
				ExpiresAt,
		},
	)
}
