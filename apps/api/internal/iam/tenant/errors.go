package tenant

import "errors"

var (
	ErrNotFound = errors.New(
		"tenant not found",
	)

	ErrInvalidInput = errors.New(
		"invalid tenant input",
	)

	ErrSlugAlreadyExists = errors.New(
		"tenant slug already exists",
	)

	ErrSuspended = errors.New(
		"tenant is suspended",
	)

	ErrDeleted = errors.New(
		"tenant is deleted",
	)
)
