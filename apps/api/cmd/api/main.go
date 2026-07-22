// Package main starts the CRM Platform API.
//
//	@title						CRM Platform API
//	@version					1.0
//	@description				Backend API for CRM authentication, users, organisations, contacts, deals, and related resources.
//	@description				Authentication uses a Redis-backed opaque session stored in an HttpOnly cookie.
//	@termsOfService				https://example.com/terms
//
//	@contact.name				Rajip Upreti
//	@contact.email				upretirajeev0@gmail.com
//
//	@license.name				Proprietary
//
//	@host						localhost:8080
//	@BasePath					/
//	@schemes					http
//
//	@securityDefinitions.apikey	CookieAuth
//	@in						header
//	@name					Cookie
//
//	@tag.name					Health
//	@tag.description			Service health and dependency diagnostics.
//
//	@tag.name					Authentication
//	@tag.description			OIDC login, current-user session, and logout operations.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rajipupreti/crm-platform/apps/api/internal/auth"
	"github.com/rajipupreti/crm-platform/apps/api/internal/config"
	"github.com/rajipupreti/crm-platform/apps/api/internal/database"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/onboarding"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/tenant"
	"github.com/rajipupreti/crm-platform/apps/api/internal/middleware"
	"github.com/rajipupreti/crm-platform/apps/api/internal/redisclient"
	"github.com/rajipupreti/crm-platform/apps/api/internal/server"
	"github.com/rajipupreti/crm-platform/apps/api/internal/session"
	"github.com/rajipupreti/crm-platform/apps/api/internal/user"

	_ "github.com/rajipupreti/crm-platform/apps/api/docs"
)

func main() {
	if err := run(); err != nil {
		log.Printf(
			"application stopped with error: %v",
			err,
		)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	startupContext, startupCancel := context.WithTimeout(
		context.Background(),
		20*time.Second,
	)
	defer startupCancel()

	postgresPool, err := database.OpenPostgres(
		startupContext,
		database.PostgresConfig{
			URL:             cfg.DatabaseURL,
			MaxConnections:  10,
			MinConnections:  1,
			MaxConnLifetime: 30 * time.Minute,
			MaxConnIdleTime: 5 * time.Minute,
		},
	)
	if err != nil {
		return err
	}
	tenantRepository, err := tenant.NewPostgresRepository(
		postgresPool,
	)
	if err != nil {
		return fmt.Errorf(
			"create tenant repository: %w",
			err,
		)
	}

	tenantService, err := tenant.NewService(
		tenantRepository,
	)
	if err != nil {
		return fmt.Errorf(
			"create tenant service: %w",
			err,
		)
	}
	membershipRepository, err := membership.NewPostgresRepository(
		postgresPool,
	)
	if err != nil {
		return fmt.Errorf(
			"create membership repository: %w",
			err,
		)
	}

	membershipService, err := membership.NewService(
		membershipRepository,
	)
	if err != nil {
		return fmt.Errorf(
			"create membership service: %w",
			err,
		)
	}
	defer postgresPool.Close()
	onboardingRepository, err := onboarding.NewPostgresRepository(
		postgresPool,
	)
	if err != nil {
		return fmt.Errorf(
			"create onboarding repository: %w",
			err,
		)
	}

	onboardingService, err := onboarding.NewService(
		onboardingRepository,
	)
	if err != nil {
		return fmt.Errorf(
			"create onboarding service: %w",
			err,
		)
	}
	userRepository, err := user.NewPostgresRepository(
		postgresPool,
	)
	if err != nil {
		return err
	}

	userService, err := user.NewService(
		userRepository,
	)
	if err != nil {
		return err
	}

	oidcClient, err := auth.NewOIDCClient(
		startupContext,
		auth.OIDCConfig{
			IssuerURL:             cfg.OIDCIssuerURL,
			ClientID:              cfg.OIDCClientID,
			ClientSecret:          cfg.OIDCClientSecret,
			RedirectURL:           cfg.OIDCRedirectURL,
			DockerKeycloakAddress: cfg.OIDCDockerKeycloakAddress,
		},
	)
	if err != nil {
		return err
	}

	redisClient, err := redisclient.New(
		startupContext,
		redisclient.Config{
			Address:  cfg.RedisAddress,
			Password: cfg.RedisPassword,
			Database: cfg.RedisDatabase,
		},
	)
	if err != nil {
		return err
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf(
				"close Redis client: %v",
				err,
			)
		}
	}()

	loginTransactionStore, err := auth.NewRedisLoginTransactionStore(
		redisClient,
		cfg.RedisKeyPrefix,
	)
	if err != nil {
		return err
	}

	sessionStore, err := session.NewRedisStore(
		redisClient,
		cfg.Session.RedisKeyPrefix,
	)
	if err != nil {
		return err
	}

	sessionService, err := session.NewService(
		sessionStore,
		cfg.Session.TTL,
	)
	if err != nil {
		return err
	}

	sessionCookieManager, err := session.NewCookieManager(
		session.CookieConfig{
			Name:     cfg.Session.CookieName,
			Path:     "/",
			Domain:   cfg.Session.CookieDomain,
			Secure:   cfg.Session.CookieSecure,
			SameSite: cfg.Session.CookieSameSite,
		},
	)
	if err != nil {
		return err
	}

	authenticationMiddleware, err := middleware.NewAuthenticationMiddleware(
		sessionService,
		sessionCookieManager,
		userService,
		tenantService,
		membershipService,
	)
	if err != nil {
		return fmt.Errorf(
			"create authentication middleware: %w",
			err,
		)
	}

	authHandler, err := auth.NewHandler(
		oidcClient,
		loginTransactionStore,
		userService,
		onboardingService,
		sessionService,
		sessionService,
		sessionCookieManager,
		auth.HandlerConfig{
			FrontendURL:             cfg.FrontendURL,
			LoginTransactionTTL:     10 * time.Minute,
			DefaultLoginDestination: "/dashboard",
		},
	)
	if err != nil {
		return err
	}

	appServer := server.New(
		cfg.HTTPAddress,
		oidcClient,
		authHandler,
		authenticationMiddleware,
	)

	httpServer := appServer.HTTPServer()
	httpHandler := httpServer.Handler

	corsMiddleware := middleware.NewCORS(
		cfg.CORSAllowedOrigins,
	)

	httpHandler = corsMiddleware.Wrap(
		httpHandler,
	)

	httpServer.Handler = httpHandler

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf(
			"CRM API listening on %s",
			cfg.HTTPAddress,
		)

		serverErrors <- httpServer.ListenAndServe()
	}()

	shutdownSignal := make(
		chan os.Signal,
		1,
	)

	signal.Notify(
		shutdownSignal,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	select {
	case receivedSignal := <-shutdownSignal:
		log.Printf(
			"received shutdown signal: %s",
			receivedSignal,
		)

	case serverErr := <-serverErrors:
		if !errors.Is(
			serverErr,
			http.ErrServerClosed,
		) {
			return serverErr
		}
	}

	shutdownContext, shutdownCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer shutdownCancel()

	if err := httpServer.Shutdown(
		shutdownContext,
	); err != nil {
		return err
	}

	log.Println("CRM API stopped cleanly")

	return nil
}
