package user

import "errors"

var (
	ErrNotFound = errors.New(
		"user not found",
	)

	ErrInvalidIdentity = errors.New(
		"invalid external identity",
	)

	ErrSuspended = errors.New(
		"user is suspended",
	)
)
