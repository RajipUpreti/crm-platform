package auth

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/rajipupreti/crm-platform/apps/api/internal/httpresponse"
	"github.com/rajipupreti/crm-platform/apps/api/internal/user"
)

func (h *Handler) Callback(
	w http.ResponseWriter,
	r *http.Request,
) {
	if providerError := strings.TrimSpace(
		r.URL.Query().Get("error"),
	); providerError != "" {
		h.handleProviderError(
			w,
			r,
			providerError,
		)
		return
	}

	state := strings.TrimSpace(
		r.URL.Query().Get("state"),
	)
	if state == "" {
		h.writeCallbackError(
			w,
			ErrMissingState,
			http.StatusBadRequest,
		)
		return
	}

	authorizationCode := strings.TrimSpace(
		r.URL.Query().Get("code"),
	)
	if authorizationCode == "" {
		h.writeCallbackError(
			w,
			ErrMissingAuthorizationCode,
			http.StatusBadRequest,
		)
		return
	}

	transaction, err :=
		h.transactionStore.Consume(
			r.Context(),
			state,
		)
	if err != nil {
		switch {
		case errors.Is(
			err,
			ErrLoginTransactionNotFound,
		),
			errors.Is(
				err,
				ErrLoginTransactionExpired,
			):
			h.writeCallbackError(
				w,
				ErrInvalidState,
				http.StatusBadRequest,
			)

		default:
			log.Printf(
				"consume OIDC login transaction: %v",
				err,
			)

			httpresponse.Error(
				w,
				http.StatusServiceUnavailable,
				"authentication_unavailable",
				"authentication is temporarily unavailable",
			)
		}

		return
	}

	if !constantTimeEqual(
		transaction.State,
		state,
	) {
		log.Printf(
			"OIDC transaction state mismatch",
		)

		h.writeCallbackError(
			w,
			ErrInvalidState,
			http.StatusBadRequest,
		)
		return
	}

	identity, err := h.exchangeAndVerify(
		r,
		authorizationCode,
		transaction,
	)
	if err != nil {
		log.Printf(
			"OIDC callback verification failed: %v",
			err,
		)

		httpresponse.Error(
			w,
			http.StatusUnauthorized,
			"authentication_failed",
			"authentication failed",
		)
		return
	}

	crmUser, err :=
		h.identitySynchronizer.SynchronizeIdentity(
			r.Context(),
			user.Identity{
				Provider:       "keycloak",
				ProviderUserID: identity.Subject,
				Email:          identity.Email,
				EmailVerified:  identity.EmailVerified,
				FirstName:      identity.GivenName,
				LastName:       identity.FamilyName,
				DisplayName:    identity.Name,
			},
		)
	if err != nil {
		log.Printf(
			"synchronize authenticated CRM user: %v",
			err,
		)

		status := http.StatusInternalServerError
		responseCode := "user_synchronization_failed"
		message := "could not prepare authenticated user"

		switch {
		case errors.Is(err, user.ErrSuspended):
			status = http.StatusForbidden
			responseCode = "user_suspended"
			message = "user access is suspended"

		case errors.Is(err, user.ErrInvalidIdentity):
			status = http.StatusUnauthorized
			responseCode = "invalid_identity"
			message = "authenticated identity is not valid"
		}

		httpresponse.Error(
			w,
			status,
			responseCode,
			message,
		)
		return
	}

	destination, err :=
		h.frontendDestination(
			transaction.ReturnTo,
		)
	if err != nil {
		log.Printf(
			"build frontend redirect: %v",
			err,
		)

		httpresponse.Error(
			w,
			http.StatusInternalServerError,
			"redirect_failed",
			"could not complete authentication",
		)
		return
	}

	createdSession, err :=
		h.sessionCreator.Create(
			r.Context(),
			crmUser.ID,
		)
	if err != nil {
		log.Printf(
			"create application session: %v",
			err,
		)

		httpresponse.Error(
			w,
			http.StatusInternalServerError,
			"session_creation_failed",
			"could not establish application session",
		)
		return
	}

	h.sessionCookieManager.Set(
		w,
		createdSession.Token,
		createdSession.ExpiresAt,
	)

	http.Redirect(
		w,
		r,
		destination,
		http.StatusSeeOther,
	)
}

func (h *Handler) exchangeAndVerify(
	r *http.Request,
	code string,
	transaction LoginTransaction,
) (IdentityClaims, error) {
	oidcContext := oidc.ClientContext(
		r.Context(),
		h.oidcClient.HTTPClient,
	)

	oauthToken, err :=
		h.oidcClient.OAuth2Config.Exchange(
			oidcContext,
			code,
			oauth2.VerifierOption(
				transaction.CodeVerifier,
			),
		)
	if err != nil {
		return IdentityClaims{}, fmt.Errorf(
			"%w: %v",
			ErrAuthorizationCodeExchange,
			err,
		)
	}

	rawIDToken, ok :=
		oauthToken.Extra("id_token").(string)

	if !ok ||
		strings.TrimSpace(rawIDToken) == "" {
		return IdentityClaims{},
			ErrMissingIDToken
	}

	idToken, err :=
		h.oidcClient.Verifier.Verify(
			oidcContext,
			rawIDToken,
		)
	if err != nil {
		return IdentityClaims{}, fmt.Errorf(
			"%w: %v",
			ErrInvalidIDToken,
			err,
		)
	}

	if !constantTimeEqual(
		idToken.Nonce,
		transaction.Nonce,
	) {
		return IdentityClaims{},
			ErrInvalidNonce
	}

	var claims IdentityClaims

	if err := idToken.Claims(&claims); err != nil {
		return IdentityClaims{}, fmt.Errorf(
			"%w: decode claims: %v",
			ErrInvalidIdentityClaims,
			err,
		)
	}

	if err := claims.Validate(); err != nil {
		return IdentityClaims{}, err
	}

	return claims, nil
}

func (h *Handler) handleProviderError(
	w http.ResponseWriter,
	r *http.Request,
	providerError string,
) {
	errorDescription := strings.TrimSpace(
		r.URL.Query().Get("error_description"),
	)

	log.Printf(
		"OIDC provider returned error=%q description=%q",
		providerError,
		errorDescription,
	)

	message := "authentication was not completed"

	if providerError == "access_denied" {
		message = "authentication was cancelled"
	}

	httpresponse.Error(
		w,
		http.StatusUnauthorized,
		"provider_authentication_failed",
		message,
	)
}

func (h *Handler) writeCallbackError(
	w http.ResponseWriter,
	err error,
	status int,
) {
	responseCode := "authentication_failed"

	switch {
	case errors.Is(err, ErrMissingState):
		responseCode = "missing_state"

	case errors.Is(
		err,
		ErrMissingAuthorizationCode,
	):
		responseCode = "missing_authorization_code"

	case errors.Is(err, ErrInvalidState):
		responseCode = "invalid_state"
	}

	httpresponse.Error(
		w,
		status,
		responseCode,
		err.Error(),
	)
}

func constantTimeEqual(
	actual string,
	expected string,
) bool {
	if actual == "" || expected == "" {
		return false
	}

	if len(actual) != len(expected) {
		return false
	}

	return subtle.ConstantTimeCompare(
		[]byte(actual),
		[]byte(expected),
	) == 1
}
