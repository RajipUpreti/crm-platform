package iamhttp

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/rajipupreti/crm-platform/apps/api/internal/httpresponse"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
	"github.com/rajipupreti/crm-platform/apps/api/internal/requestcontext"
)

type MembershipHandler struct {
	service *membership.Service
}

func NewMembershipHandler(
	service *membership.Service,
) (*MembershipHandler, error) {
	if service == nil {
		return nil, errors.New(
			"membership service is required",
		)
	}

	return &MembershipHandler{
		service: service,
	}, nil
}

type MemberListResponse struct {
	Members []membership.Member `json:"members"`
}

type UpdateMemberRoleRequest struct {
	Role membership.Role `json:"role" enums:"ADMIN,MEMBER"`
}

type UpdateMemberStatusRequest struct {
	Status membership.Status `json:"status" enums:"ACTIVE,SUSPENDED"`
}

// ListMembers lists members of the current tenant.
//
//	@Summary		List tenant members
//	@Description	Returns membership and user details for members of the current tenant.
//	@Tags			Members
//	@Produce		json
//	@Security		CookieAuth
//	@Success		200	{object}	MemberListResponse
//	@Failure		401	{object}	httpresponse.ErrorResponse
//	@Failure		403	{object}	httpresponse.ErrorResponse
//	@Failure		500	{object}	httpresponse.ErrorResponse
//	@Router			/api/v1/members [get]
func (h *MembershipHandler) ListMembers(
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

	members, err := h.service.ListDetailedByTenantID(
		r.Context(),
		authentication.Tenant.ID,
	)
	if err != nil {
		log.Printf("list tenant members: %v", err)
		httpresponse.Error(
			w,
			http.StatusInternalServerError,
			"member_listing_failed",
			"could not list tenant members",
		)
		return
	}

	httpresponse.JSON(
		w,
		http.StatusOK,
		MemberListResponse{Members: members},
	)
}

// UpdateMemberRole changes a tenant member's role.
//
//	@Summary		Change member role
//	@Description	Changes a membership role to ADMIN or MEMBER while enforcing tenant and ownership rules.
//	@Tags			Members
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			membershipId	path		string					true	"Membership ID"	format(uuid)
//	@Param			request			body		UpdateMemberRoleRequest	true	"New membership role"
//	@Success		200				{object}	membership.Membership
//	@Failure		400				{object}	httpresponse.ErrorResponse
//	@Failure		401				{object}	httpresponse.ErrorResponse
//	@Failure		403				{object}	httpresponse.ErrorResponse
//	@Failure		404				{object}	httpresponse.ErrorResponse
//	@Failure		409				{object}	httpresponse.ErrorResponse
//	@Failure		500				{object}	httpresponse.ErrorResponse
//	@Router			/api/v1/members/{membershipId}/role [patch]
func (h *MembershipHandler) UpdateMemberRole(
	w http.ResponseWriter,
	r *http.Request,
) {
	authentication, ok := membershipAuthentication(w, r)
	if !ok {
		return
	}

	var request UpdateMemberRoleRequest
	if !decodeMembershipRequest(w, r, &request) {
		return
	}

	updated, err := h.service.ChangeRoleForTenant(
		r.Context(),
		authentication.Tenant.ID,
		authentication.User.ID,
		authentication.Membership.Role,
		strings.TrimSpace(r.PathValue("membershipId")),
		request.Role,
	)
	if err != nil {
		writeMembershipError(w, err)
		return
	}

	httpresponse.JSON(w, http.StatusOK, updated)
}

// UpdateMemberStatus suspends or reactivates a tenant membership.
//
//	@Summary		Change member status
//	@Description	Suspends or reactivates a membership while enforcing tenant and ownership rules.
//	@Tags			Members
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			membershipId	path		string						true	"Membership ID"	format(uuid)
//	@Param			request			body		UpdateMemberStatusRequest	true	"New membership status"
//	@Success		200				{object}	membership.Membership
//	@Failure		400				{object}	httpresponse.ErrorResponse
//	@Failure		401				{object}	httpresponse.ErrorResponse
//	@Failure		403				{object}	httpresponse.ErrorResponse
//	@Failure		404				{object}	httpresponse.ErrorResponse
//	@Failure		409				{object}	httpresponse.ErrorResponse
//	@Failure		500				{object}	httpresponse.ErrorResponse
//	@Router			/api/v1/members/{membershipId}/status [patch]
func (h *MembershipHandler) UpdateMemberStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	authentication, ok := membershipAuthentication(w, r)
	if !ok {
		return
	}

	var request UpdateMemberStatusRequest
	if !decodeMembershipRequest(w, r, &request) {
		return
	}

	updated, err := h.service.ChangeStatusForTenant(
		r.Context(),
		authentication.Tenant.ID,
		authentication.User.ID,
		authentication.Membership.Role,
		strings.TrimSpace(r.PathValue("membershipId")),
		request.Status,
	)
	if err != nil {
		writeMembershipError(w, err)
		return
	}

	httpresponse.JSON(w, http.StatusOK, updated)
}

// RemoveMember removes a member from the current tenant.
//
//	@Summary		Remove tenant member
//	@Description	Marks a membership as LEFT while enforcing tenant, self-removal, hierarchy, and ownership rules.
//	@Tags			Members
//	@Security		CookieAuth
//	@Param			membershipId	path	string	true	"Membership ID"	format(uuid)
//	@Success		204
//	@Failure		400	{object}	httpresponse.ErrorResponse
//	@Failure		401	{object}	httpresponse.ErrorResponse
//	@Failure		403	{object}	httpresponse.ErrorResponse
//	@Failure		404	{object}	httpresponse.ErrorResponse
//	@Failure		409	{object}	httpresponse.ErrorResponse
//	@Failure		500	{object}	httpresponse.ErrorResponse
//	@Router			/api/v1/members/{membershipId} [delete]
func (h *MembershipHandler) RemoveMember(
	w http.ResponseWriter,
	r *http.Request,
) {
	authentication, ok := membershipAuthentication(w, r)
	if !ok {
		return
	}

	err := h.service.RemoveForTenant(
		r.Context(),
		authentication.Tenant.ID,
		authentication.User.ID,
		authentication.Membership.Role,
		strings.TrimSpace(r.PathValue("membershipId")),
	)
	if err != nil {
		writeMembershipError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func membershipAuthentication(
	w http.ResponseWriter,
	r *http.Request,
) (requestcontext.Authentication, bool) {
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
		return requestcontext.Authentication{}, false
	}

	return authentication, true
}

func decodeMembershipRequest(
	w http.ResponseWriter,
	r *http.Request,
	value any,
) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		httpresponse.Error(
			w,
			http.StatusBadRequest,
			"invalid_request",
			"request body is invalid",
		)
		return false
	}

	return true
}

func writeMembershipError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, membership.ErrInvalidInput):
		httpresponse.Error(
			w,
			http.StatusBadRequest,
			"invalid_membership",
			"membership request is invalid",
		)

	case errors.Is(err, membership.ErrNotFound):
		httpresponse.Error(
			w,
			http.StatusNotFound,
			"membership_not_found",
			"membership was not found",
		)

	case errors.Is(err, membership.ErrLastOwner):
		httpresponse.Error(
			w,
			http.StatusConflict,
			"last_owner_required",
			"the final active owner cannot be modified",
		)

	case errors.Is(err, membership.ErrSelfModification):
		httpresponse.Error(
			w,
			http.StatusForbidden,
			"self_modification_forbidden",
			"use the dedicated account workflow to modify your own membership",
		)

	case errors.Is(err, membership.ErrForbidden):
		httpresponse.Error(
			w,
			http.StatusForbidden,
			"membership_modification_forbidden",
			"you cannot modify this membership",
		)

	default:
		log.Printf("manage tenant membership: %v", err)
		httpresponse.Error(
			w,
			http.StatusInternalServerError,
			"membership_update_failed",
			"could not update membership",
		)
	}
}
