package server

import (
	"net/http"
	"time"

	"github.com/rajipupreti/crm-platform/apps/api/internal/auth"
	"github.com/rajipupreti/crm-platform/apps/api/internal/httpresponse"
)

type Server struct {
	httpServer  *http.Server
	oidcClient  *auth.OIDCClient
	authHandler *auth.Handler
}

func New(
	address string,
	oidcClient *auth.OIDCClient,
	authHandler *auth.Handler,
) *Server {
	server := &Server{
		oidcClient:  oidcClient,
		authHandler: authHandler,
	}

	mux := http.NewServeMux()

	mux.HandleFunc(
		"GET /health",
		server.health,
	)

	mux.HandleFunc(
		"GET /health/auth",
		server.authHealth,
	)

	mux.HandleFunc(
		"GET /auth/login",
		server.authHandler.Login,
	)

	mux.HandleFunc(
		"GET /auth/callback",
		server.authHandler.Callback,
	)

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
	httpresponse.JSON(
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
	httpresponse.JSON(
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
