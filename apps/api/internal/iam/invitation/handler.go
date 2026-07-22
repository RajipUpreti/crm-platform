package invitation

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/rajipupreti/crm-platform/apps/api/internal/httpresponse"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
	"github.com/rajipupreti/crm-platform/apps/api/internal/requestcontext"
)

type Handler struct {
	service *Service
}

func NewHandler(
	service *Service,
) (*Handler, error) {
	if service == nil {
		return nil, errors.New(
			"invitation service is required",
		)
	}

	return &Handler{
		service: service,
	}, nil
}

type CreateRequest struct {
	Email string `json:"email" format:"email"`

	Role membership.Role `json:"role" enums:"ADMIN,MEMBER"`
}

type AcceptRequest struct {
	Token string `json:"token"`
}

func (h *Handler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	authn, err := requestcontext.AuthenticationFromContext(
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

	if authn.Membership.Role !=
		membership.RoleOwner &&
		authn.Membership.Role !=
			membership.RoleAdmin {
		httpresponse.Error(
			w,
			http.StatusForbidden,
			"insufficient_permission",
			"owner or admin access is required",
		)
		return
	}

	var request CreateRequest

	if err := json.NewDecoder(
		r.Body,
	).Decode(&request); err != nil {
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
		authn.Tenant.ID,
		authn.User.ID,
		request.Email,
		request.Role,
	)
	if err != nil {
		switch {
		case errors.Is(
			err,
			ErrAlreadyPending,
		):
			httpresponse.Error(
				w,
				http.StatusConflict,
				"invitation_already_pending",
				"a pending invitation already exists",
			)

		default:
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

func (h *Handler) Accept(
	w http.ResponseWriter,
	r *http.Request,
) {
	authn, err := requestcontext.AuthenticationFromContext(
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

	var request AcceptRequest

	if err := json.NewDecoder(
		r.Body,
	).Decode(&request); err != nil {
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
		AcceptInput{
			Token:     request.Token,
			UserID:    authn.User.ID,
			UserEmail: authn.User.Email,
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrExpired):
			httpresponse.Error(
				w,
				http.StatusGone,
				"invitation_expired",
				"invitation has expired",
			)

		case errors.Is(err, ErrEmailMismatch):
			httpresponse.Error(
				w,
				http.StatusForbidden,
				"invitation_email_mismatch",
				"invitation belongs to another email",
			)

		case errors.Is(err, ErrNotFound):
			httpresponse.Error(
				w,
				http.StatusNotFound,
				"invitation_not_found",
				"invitation was not found",
			)

		default:
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
