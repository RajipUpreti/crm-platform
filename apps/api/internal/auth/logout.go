package auth

import (
	"log"
	"net/http"

	"github.com/rajipupreti/crm-platform/apps/api/internal/httpresponse"
)

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
		struct {
			LoggedOut bool `json:"loggedOut"`
		}{
			LoggedOut: true,
		},
	)
}
