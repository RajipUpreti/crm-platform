package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rajipupreti/crm-platform/apps/api/internal/auth"
)

type Server struct {
	httpServer *http.Server
	oidcClient *auth.OIDCClient
}

func New(
	address string,
	oidcClient *auth.OIDCClient,
) *Server {
	mux := http.NewServeMux()

	server := &Server{
		oidcClient: oidcClient,
	}

	mux.HandleFunc("GET /health", server.health)
	mux.HandleFunc("GET /health/auth", server.authHealth)

	server.httpServer = &http.Server{
		Addr:              address,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return server
}

func (s *Server) HTTPServer() *http.Server {
	return s.httpServer
}

func (s *Server) health(
	w http.ResponseWriter,
	r *http.Request,
) {
	writeJSON(
		w,
		http.StatusOK,
		map[string]string{
			"status": "ok",
		},
	)
}

func (s *Server) authHealth(
	w http.ResponseWriter,
	r *http.Request,
) {
	writeJSON(
		w,
		http.StatusOK,
		map[string]string{
			"status":      "ok",
			"issuer":      s.oidcClient.IssuerURL,
			"clientId":    s.oidcClient.OAuth2Config.ClientID,
			"redirectUrl": s.oidcClient.OAuth2Config.RedirectURL,
		},
	)
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	value any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		return
	}
}
