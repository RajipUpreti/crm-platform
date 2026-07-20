package auth

import "errors"

var (
	ErrProviderRejectedAuthentication = errors.New(
		"identity provider rejected authentication",
	)

	ErrMissingAuthorizationCode = errors.New(
		"authorization code is missing",
	)

	ErrMissingState = errors.New(
		"authentication state is missing",
	)

	ErrInvalidState = errors.New(
		"authentication state is invalid",
	)

	ErrAuthorizationCodeExchange = errors.New(
		"authorization code exchange failed",
	)

	ErrMissingIDToken = errors.New(
		"ID token is missing",
	)

	ErrInvalidIDToken = errors.New(
		"ID token is invalid",
	)

	ErrInvalidNonce = errors.New(
		"ID token nonce is invalid",
	)

	ErrInvalidIdentityClaims = errors.New(
		"identity claims are invalid",
	)
)
