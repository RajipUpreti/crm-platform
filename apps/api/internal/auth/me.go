package auth

import (
	"net/http"
	"time"

	"github.com/rajipupreti/crm-platform/apps/api/internal/httpresponse"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/tenant"
	"github.com/rajipupreti/crm-platform/apps/api/internal/requestcontext"
	"github.com/rajipupreti/crm-platform/apps/api/internal/user"
)

type CurrentUserResponse struct {
	User user.User `json:"user"`

	Tenant tenant.Tenant `json:"tenant"`

	Membership membership.Membership `json:"membership"`

	ExpiresAt time.Time `json:"expiresAt"`
}

// Me returns the currently authenticated CRM user.
//
//	@Summary		Get current user
//	@Description	Returns the user associated with the Redis-backed application session.
//	@Tags			Authentication
//	@Produce		json
//	@Security		CookieAuth
//	@Success		200	{object}	CurrentUserResponse
//	@Failure		401	{object}	SwaggerErrorResponse
//	@Failure		403	{object}	SwaggerErrorResponse
//	@Failure		500	{object}	SwaggerErrorResponse
//	@Router			/auth/me [get]
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

			Tenant: authentication.Tenant,

			Membership: authentication.Membership,

			ExpiresAt: authentication.
				Session.
				ExpiresAt,
		},
	)
}
