package invitation

import "errors"

var (
	ErrNotFound = errors.New(
		"invitation not found",
	)

	ErrInvalidInput = errors.New(
		"invalid invitation input",
	)

	ErrAlreadyPending = errors.New(
		"pending invitation already exists",
	)

	ErrExpired = errors.New(
		"invitation expired",
	)

	ErrAlreadyAccepted = errors.New(
		"invitation already accepted",
	)

	ErrRevoked = errors.New(
		"invitation revoked",
	)

	ErrEmailMismatch = errors.New(
		"invitation email does not match authenticated user",
	)
)
