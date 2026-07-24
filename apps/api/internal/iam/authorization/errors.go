package authorization

import "errors"

var (
	ErrInvalidRole = errors.New(
		"invalid membership role",
	)

	ErrInvalidPermission = errors.New(
		"invalid permission",
	)

	ErrPermissionDenied = errors.New(
		"permission denied",
	)
)
