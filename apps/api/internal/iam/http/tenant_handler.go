package iamhttp

import (
	"errors"
	"log"
	"net/http"

	"github.com/rajipupreti/crm-platform/apps/api/internal/httpresponse"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/tenant"
	"github.com/rajipupreti/crm-platform/apps/api/internal/requestcontext"
)

type TenantHandler struct {
	service *tenant.Service
}

func NewTenantHandler(
	service *tenant.Service,
) (*TenantHandler, error) {
	if service == nil {
		return nil, errors.New(
			"tenant service is required",
		)
	}

	return &TenantHandler{
		service: service,
	}, nil
}

type TenantListResponse struct {
	Tenants []tenant.Access `json:"tenants"`
}

// ListTenants returns all active tenants available to the authenticated user.
//
//	@Summary		List available tenants
//	@Description	Returns active tenant memberships for the authenticated user.
//	@Tags			Tenants
//	@Produce		json
//	@Security		CookieAuth
//	@Success		200	{object}	TenantListResponse
//	@Failure		401	{object}	httpresponse.ErrorResponse
//	@Failure		500	{object}	httpresponse.ErrorResponse
//	@Router			/api/v1/tenants [get]
func (h *TenantHandler) ListTenants(
	w http.ResponseWriter,
	r *http.Request,
) {
	authentication, err :=
		requestcontext.AuthenticationFromContext(
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

	accesses, err :=
		h.service.ListAccessByUserID(
			r.Context(),
			authentication.User.ID,
			authentication.Tenant.ID,
		)
	if err != nil {
		log.Printf(
			"list tenant access: %v",
			err,
		)

		httpresponse.Error(
			w,
			http.StatusInternalServerError,
			"tenant_listing_failed",
			"could not list available workspaces",
		)
		return
	}

	httpresponse.JSON(
		w,
		http.StatusOK,
		TenantListResponse{
			Tenants: accesses,
		},
	)
}
