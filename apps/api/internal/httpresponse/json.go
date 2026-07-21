package httpresponse

import (
	"encoding/json"
	"log"
	"net/http"
)

func JSON(
	w http.ResponseWriter,
	status int,
	value any,
) {
	w.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)

	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf(
			"encode HTTP JSON response: %v",
			err,
		)
	}
}

func Error(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
) {
	JSON(
		w,
		status,
		ErrorResponse{
			Error: ErrorDetail{
				Code:    code,
				Message: message,
			},
		},
	)
}
