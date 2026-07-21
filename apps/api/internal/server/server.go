package server

import (
	"net/http"
	"time"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/rajipupreti/crm-platform/apps/api/internal/auth"
	"github.com/rajipupreti/crm-platform/apps/api/internal/httpresponse"
	"github.com/rajipupreti/crm-platform/apps/api/internal/middleware"
)

type Server struct {
	httpServer     *http.Server
	oidcClient     *auth.OIDCClient
	authHandler    *auth.Handler
	authMiddleware *middleware.AuthenticationMiddleware
}

func New(
	address string,
	oidcClient *auth.OIDCClient,
	authHandler *auth.Handler,
	authMiddleware *middleware.AuthenticationMiddleware,
) *Server {
	server := &Server{
		oidcClient:     oidcClient,
		authHandler:    authHandler,
		authMiddleware: authMiddleware,
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

	mux.Handle(
		"GET /auth/me",
		server.authMiddleware.Require(
			http.HandlerFunc(
				server.authHandler.Me,
			),
		),
	)

	mux.HandleFunc(
		"POST /auth/logout",
		server.authHandler.Logout,
	)
	mux.HandleFunc(
		"GET /swagger",
		func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(
				w,
				r,
				"/swagger/index.html",
				http.StatusTemporaryRedirect,
			)
		},
	)

	mux.Handle(
		"GET /swagger/",
		httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"),
		),
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

// health reports whether the API process is running.
//
//	@Summary	API health
//	@Tags		Health
//	@Produce	json
//	@Success	200	{object}	httpresponse.HealthResponse
//	@Router		/health [get]
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

// authHealth reports the configured OpenID Connect client metadata.
//
//	@Summary	Authentication health
//	@Tags		Health
//	@Produce	json
//	@Success	200	{object}	httpresponse.DependencyHealthResponse
//	@Router		/health/auth [get]
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
