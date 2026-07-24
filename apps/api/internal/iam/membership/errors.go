package membership

import "errors"

var (
	ErrNotFound = errors.New(
		"membership not found",
	)

	ErrInvalidInput = errors.New(
		"invalid membership input",
	)

	ErrAlreadyExists = errors.New(
		"membership already exists",
	)

	ErrSuspended = errors.New(
		"membership is suspended",
	)

	ErrInactive = errors.New(
		"membership is inactive",
	)

	ErrForbidden = errors.New(
		"membership modification is forbidden",
	)

	ErrLastOwner = errors.New(
		"the final active owner cannot be modified",
	)

	ErrSelfModification = errors.New(
		"members cannot modify themselves through the administration API",
	)
)
