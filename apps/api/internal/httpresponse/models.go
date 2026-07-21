package httpresponse

// ErrorDetail describes a stable machine-readable API error.
type ErrorDetail struct {
	Code    string `json:"code" example:"authentication_required"`
	Message string `json:"message" example:"authentication is required"`
}

// ErrorResponse is returned when an API operation fails.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// HealthResponse describes basic API process health.
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

// DependencyHealthResponse describes external dependency health.
type DependencyHealthResponse struct {
	Status      string `json:"status" example:"ok"`
	Issuer      string `json:"issuer" format:"uri" example:"http://localhost:8081/realms/crm"`
	ClientID    string `json:"clientId" example:"crm-backend"`
	RedirectURL string `json:"redirectUrl" format:"uri" example:"http://localhost:8080/auth/callback"`
}
