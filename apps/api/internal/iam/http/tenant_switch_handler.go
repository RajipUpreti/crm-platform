package iamhttp

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/rajipupreti/crm-platform/apps/api/internal/httpresponse"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/tenant"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/tenantswitch"
	"github.com/rajipupreti/crm-platform/apps/api/internal/requestcontext"
	"github.com/rajipupreti/crm-platform/apps/api/internal/session"
)

type TenantSwitchHandler struct {
	service *tenantswitch.Service

	cookies *session.CookieManager
}

func NewTenantSwitchHandler(
	service *tenantswitch.Service,
	cookies *session.CookieManager,
) (*TenantSwitchHandler, error) {
	if service == nil {
		return nil, errors.New(
			"tenant switch service is required",
		)
	}

	if cookies == nil {
		return nil, errors.New(
			"session cookie manager is required",
		)
	}

	return &TenantSwitchHandler{
		service: service,
		cookies: cookies,
	}, nil
}

type TenantSwitchResponse struct {
	Tenant tenant.Tenant `json:"tenant"`

	Membership membership.Membership `json:"membership"`
}

// SwitchTenant changes the selected tenant for the current session.
//
//	@Summary		Switch current tenant
//	@Description	Validates active membership, rotates the application session, and selects the requested tenant.
//	@Tags			Tenants
//	@Produce		json
//	@Param			tenantId	path		string	true	"Tenant ID"	format(uuid)
//	@Success		200			{object}	TenantSwitchResponse
//	@Failure		400			{object}	httpresponse.ErrorResponse
//	@Failure		401			{object}	httpresponse.ErrorResponse
//	@Failure		403			{object}	httpresponse.ErrorResponse
//	@Failure		404			{object}	httpresponse.ErrorResponse
//	@Failure		500			{object}	httpresponse.ErrorResponse
//	@Router			/api/v1/tenants/{tenantId}/switch [post]
func (h *TenantSwitchHandler) SwitchTenant(
	w http.ResponseWriter,
	r *http.Request,
) {
	tenantID := strings.TrimSpace(
		r.PathValue("tenantId"),
	)

	if tenantID == "" {
		httpresponse.Error(
			w,
			http.StatusBadRequest,
			"invalid_tenant_id",
			"tenant ID is required",
		)
		return
	}

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

	currentToken, err := h.cookies.Read(r)
	if err != nil {
		httpresponse.Error(
			w,
			http.StatusUnauthorized,
			"authentication_required",
			"authentication is required",
		)
		return
	}

	if tenantID == authentication.Tenant.ID {
		httpresponse.JSON(
			w,
			http.StatusOK,
			TenantSwitchResponse{
				Tenant: authentication.Tenant,

				Membership: authentication.Membership,
			},
		)
		return
	}

	result, err := h.service.Switch(
		r.Context(),
		authentication.User.ID,
		currentToken,
		tenantID,
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
			httpresponse.Error(
				w,
				http.StatusNotFound,
				"tenant_not_found",
				"workspace was not found",
			)

		case errors.Is(
			err,
			tenant.ErrSuspended,
		):
			httpresponse.Error(
				w,
				http.StatusForbidden,
				"tenant_suspended",
				"workspace is suspended",
			)

		case errors.Is(
			err,
			membership.ErrNotFound,
		),
			errors.Is(
				err,
				membership.ErrInactive,
			):
			httpresponse.Error(
				w,
				http.StatusForbidden,
				"tenant_access_denied",
				"you do not have access to this workspace",
			)

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
				"switch tenant: %v",
				err,
			)

			httpresponse.Error(
				w,
				http.StatusInternalServerError,
				"tenant_switch_failed",
				"could not switch workspace",
			)
		}

		return
	}

	h.cookies.Set(
		w,
		result.Session.Token,
		result.Session.ExpiresAt,
	)

	httpresponse.JSON(
		w,
		http.StatusOK,
		TenantSwitchResponse{
			Tenant: result.Tenant,

			Membership: result.Membership,
		},
	)
}
