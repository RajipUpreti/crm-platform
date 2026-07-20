package user

import (
	"fmt"
)

type Handler struct {
	service *Service
}

func NewHandler(
	service *Service,
) (*Handler, error) {
	if service == nil {
		return nil, fmt.Errorf(
			"user service is required",
		)
	}

	return &Handler{
		service: service,
	}, nil
}
