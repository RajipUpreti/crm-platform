package server

import (
	"net/http"
	"time"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/rajipupreti/crm-platform/apps/api/internal/auth"
	"github.com/rajipupreti/crm-platform/apps/api/internal/httpresponse"
	iamhttp "github.com/rajipupreti/crm-platform/apps/api/internal/iam/http"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/permission"
	"github.com/rajipupreti/crm-platform/apps/api/internal/middleware"
)

type Server struct {
	httpServer *http.Server

	oidcClient          *auth.OIDCClient
	authHandler         *auth.Handler
	authMiddleware      *middleware.AuthenticationMiddleware
	authorizationGuard  *iamhttp.AuthorizationGuard
	tenantHandler       *iamhttp.TenantHandler
	invitationHandler   *iamhttp.InvitationHandler
	tenantSwitchHandler *iamhttp.TenantSwitchHandler
}

func New(
	httpAddress string,
	oidcClient *auth.OIDCClient,
	authHandler *auth.Handler,
	authMiddleware *middleware.AuthenticationMiddleware,
	authorizationGuard *iamhttp.AuthorizationGuard,
	tenantHandler *iamhttp.TenantHandler,
	invitationHandler *iamhttp.InvitationHandler,
	tenantSwitchHandler *iamhttp.TenantSwitchHandler,
) *Server {
	server := &Server{
		oidcClient:          oidcClient,
		authHandler:         authHandler,
		authMiddleware:      authMiddleware,
		authorizationGuard:  authorizationGuard,
		tenantHandler:       tenantHandler,
		invitationHandler:   invitationHandler,
		tenantSwitchHandler: tenantSwitchHandler,
	}

	mux := http.NewServeMux()

	server.registerHealthRoutes(mux)
	server.registerAuthenticationRoutes(mux)
	server.registerIAMRoutes(mux)
	server.registerSwaggerRoutes(mux)

	server.httpServer = &http.Server{
		Addr:              httpAddress,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return server
}

func (s *Server) registerHealthRoutes(
	mux *http.ServeMux,
) {
	mux.HandleFunc(
		"GET /health",
		s.health,
	)

	mux.HandleFunc(
		"GET /health/auth",
		s.authHealth,
	)
}

func (s *Server) registerAuthenticationRoutes(
	mux *http.ServeMux,
) {
	mux.HandleFunc(
		"GET /auth/login",
		s.authHandler.Login,
	)

	mux.HandleFunc(
		"GET /auth/callback",
		s.authHandler.Callback,
	)

	mux.Handle(
		"GET /auth/me",
		s.authMiddleware.Require(
			http.HandlerFunc(
				s.authHandler.Me,
			),
		),
	)

	mux.HandleFunc(
		"POST /auth/logout",
		s.authHandler.Logout,
	)
}

func (s *Server) registerIAMRoutes(
	mux *http.ServeMux,
) {
	mux.Handle(
		"POST /api/v1/tenant/invitations",
		s.authMiddleware.Require(
			s.authorizationGuard.Require(
				permission.MemberInvite,
				http.HandlerFunc(
					s.invitationHandler.
						CreateInvitation,
				),
			),
		),
	)

	mux.Handle(
		"POST /api/v1/tenants/{tenantId}/switch",
		s.authMiddleware.Require(
			http.HandlerFunc(
				s.tenantSwitchHandler.SwitchTenant,
			),
		),
	)

	mux.Handle(
		"GET /api/v1/tenants",
		s.authMiddleware.Require(
			http.HandlerFunc(
				s.tenantHandler.ListTenants,
			),
		),
	)

	mux.Handle(
		"POST /api/v1/invitations/accept",
		s.authMiddleware.Require(
			http.HandlerFunc(
				s.invitationHandler.AcceptInvitation,
			),
		),
	)
}

func (s *Server) registerSwaggerRoutes(
	mux *http.ServeMux,
) {
	mux.HandleFunc(
		"GET /swagger",
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
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
			httpSwagger.URL(
				"/swagger/doc.json",
			),
		),
	)
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
		httpresponse.HealthResponse{
			Status: "ok",
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
		httpresponse.DependencyHealthResponse{
			Status:      "ok",
			Issuer:      s.oidcClient.IssuerURL,
			ClientID:    s.oidcClient.OAuth2Config.ClientID,
			RedirectURL: s.oidcClient.OAuth2Config.RedirectURL,
		},
	)
}
