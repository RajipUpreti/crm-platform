package auth

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rajipupreti/crm-platform/apps/api/internal/httpresponse"
	"golang.org/x/oauth2"
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

	code := strings.TrimSpace(
		r.URL.Query().Get("code"),
	)
	if code == "" {
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

	if transaction.State != state {
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
		code,
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

	httpresponse.JSON(
		w,
		http.StatusOK,
		AuthenticatedIdentityResponse{
			Identity: identity.Response(),
			ReturnTo: transaction.ReturnTo,
		},
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
	code := "authentication_failed"

	switch {
	case errors.Is(err, ErrMissingState):
		code = "missing_state"

	case errors.Is(
		err,
		ErrMissingAuthorizationCode,
	):
		code = "missing_authorization_code"

	case errors.Is(err, ErrInvalidState):
		code = "invalid_state"
	}

	httpresponse.Error(
		w,
		status,
		code,
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
