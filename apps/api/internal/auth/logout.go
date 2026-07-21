package auth

import (
	"log"
	"net/http"

	"github.com/rajipupreti/crm-platform/apps/api/internal/httpresponse"
)

// Logout destroys the current application session.
//
//	@Summary		Log out
//	@Description	Deletes the current session when present and always clears the session cookie.
//	@Tags			Authentication
//	@Produce		json
//	@Success		200	{object}	LogoutResponse
//	@Router			/auth/logout [post]
func (h *Handler) Logout(
	w http.ResponseWriter,
	r *http.Request,
) {
	token, err := h.sessionCookieManager.Read(r)

	if err == nil {
		if err := h.sessionDestroyer.Delete(
			r.Context(),
			token,
		); err != nil {
			log.Printf(
				"delete application session: %v",
				err,
			)
		}
	}

	h.sessionCookieManager.Clear(w)
	httpresponse.JSON(
		w,
		http.StatusOK,
		LogoutResponse{
			LoggedOut: true,
		},
	)
}

type LogoutResponse struct {
	LoggedOut bool `json:"loggedOut" example:"true"`
}
