package iamhttp

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/rajipupreti/crm-platform/apps/api/internal/httpresponse"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/invitation"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
	"github.com/rajipupreti/crm-platform/apps/api/internal/requestcontext"
)

type InvitationHandler struct {
	service *invitation.Service
}

func NewInvitationHandler(
	service *invitation.Service,
) (*InvitationHandler, error) {
	if service == nil {
		return nil, errors.New(
			"invitation service is required",
		)
	}

	return &InvitationHandler{
		service: service,
	}, nil
}

type CreateInvitationRequest struct {
	Email string `json:"email" format:"email"`

	Role membership.Role `json:"role" enums:"ADMIN,MEMBER"`
}

type AcceptInvitationRequest struct {
	Token string `json:"token"`
}

// CreateInvitation creates a pending tenant invitation.
//
//	@Summary		Create tenant invitation
//	@Description	Invites a user to the current tenant. OWNER or ADMIN membership is required.
//	@Tags			Invitations
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			request	body		CreateInvitationRequest	true	"Invitation details"
//	@Success		201		{object}	invitation.CreatedInvitation
//	@Failure		400		{object}	httpresponse.ErrorResponse
//	@Failure		401		{object}	httpresponse.ErrorResponse
//	@Failure		403		{object}	httpresponse.ErrorResponse
//	@Failure		409		{object}	httpresponse.ErrorResponse
//	@Failure		500		{object}	httpresponse.ErrorResponse
//	@Router			/api/v1/tenant/invitations [post]
func (h *InvitationHandler) CreateInvitation(
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

	var request CreateInvitationRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		httpresponse.Error(
			w,
			http.StatusBadRequest,
			"invalid_request",
			"request body is invalid",
		)
		return
	}

	created, err := h.service.Create(
		r.Context(),
		authentication.Tenant.ID,
		authentication.User.ID,
		request.Email,
		request.Role,
	)
	if err != nil {
		switch {
		case errors.Is(
			err,
			invitation.ErrInvalidInput,
		):
			httpresponse.Error(
				w,
				http.StatusBadRequest,
				"invalid_invitation",
				"invitation details are invalid",
			)

		case errors.Is(
			err,
			invitation.ErrAlreadyPending,
		):
			httpresponse.Error(
				w,
				http.StatusConflict,
				"invitation_already_pending",
				"a pending invitation already exists",
			)

		default:
			log.Printf(
				"create invitation: %v",
				err,
			)

			httpresponse.Error(
				w,
				http.StatusInternalServerError,
				"invitation_creation_failed",
				"could not create invitation",
			)
		}

		return
	}

	httpresponse.JSON(
		w,
		http.StatusCreated,
		created,
	)
}

// AcceptInvitation accepts an invitation for the authenticated user.
//
//	@Summary		Accept tenant invitation
//	@Description	Accepts an invitation when its email matches the authenticated CRM user.
//	@Tags			Invitations
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			request	body		AcceptInvitationRequest	true	"Invitation token"
//	@Success		200		{object}	membership.Membership
//	@Failure		400		{object}	httpresponse.ErrorResponse
//	@Failure		401		{object}	httpresponse.ErrorResponse
//	@Failure		403		{object}	httpresponse.ErrorResponse
//	@Failure		404		{object}	httpresponse.ErrorResponse
//	@Failure		409		{object}	httpresponse.ErrorResponse
//	@Failure		410		{object}	httpresponse.ErrorResponse
//	@Failure		500		{object}	httpresponse.ErrorResponse
//	@Router			/api/v1/invitations/accept [post]
func (h *InvitationHandler) AcceptInvitation(
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

	var request AcceptInvitationRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		httpresponse.Error(
			w,
			http.StatusBadRequest,
			"invalid_request",
			"request body is invalid",
		)
		return
	}

	request.Token = strings.TrimSpace(
		request.Token,
	)

	createdMembership, err := h.service.Accept(
		r.Context(),
		invitation.AcceptInput{
			Token: request.Token,

			UserID: authentication.User.ID,

			UserEmail: authentication.User.Email,
		},
	)
	if err != nil {
		switch {
		case errors.Is(
			err,
			invitation.ErrInvalidInput,
		):
			httpresponse.Error(
				w,
				http.StatusBadRequest,
				"invalid_invitation",
				"invitation token is required",
			)

		case errors.Is(
			err,
			invitation.ErrExpired,
		):
			httpresponse.Error(
				w,
				http.StatusGone,
				"invitation_expired",
				"invitation has expired",
			)

		case errors.Is(
			err,
			invitation.ErrEmailMismatch,
		):
			httpresponse.Error(
				w,
				http.StatusForbidden,
				"invitation_email_mismatch",
				"invitation belongs to another email",
			)

		case errors.Is(
			err,
			invitation.ErrNotFound,
		):
			httpresponse.Error(
				w,
				http.StatusNotFound,
				"invitation_not_found",
				"invitation was not found",
			)

		case errors.Is(
			err,
			invitation.ErrAlreadyAccepted,
		):
			httpresponse.Error(
				w,
				http.StatusConflict,
				"invitation_already_accepted",
				"invitation has already been accepted",
			)

		case errors.Is(
			err,
			invitation.ErrRevoked,
		):
			httpresponse.Error(
				w,
				http.StatusGone,
				"invitation_revoked",
				"invitation has been revoked",
			)

		default:
			log.Printf(
				"accept invitation: %v",
				err,
			)

			httpresponse.Error(
				w,
				http.StatusInternalServerError,
				"invitation_acceptance_failed",
				"could not accept invitation",
			)
		}

		return
	}

	httpresponse.JSON(
		w,
		http.StatusOK,
		createdMembership,
	)
}
