package onboarding

import "errors"

var (
	ErrInvalidUser = errors.New(
		"invalid onboarding user",
	)

	ErrProvisioningFailed = errors.New(
		"tenant provisioning failed",
	)
)
