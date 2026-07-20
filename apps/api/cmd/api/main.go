package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rajipupreti/crm-platform/apps/api/internal/auth"
	"github.com/rajipupreti/crm-platform/apps/api/internal/config"
	"github.com/rajipupreti/crm-platform/apps/api/internal/database"
	"github.com/rajipupreti/crm-platform/apps/api/internal/redisclient"
	"github.com/rajipupreti/crm-platform/apps/api/internal/server"
)

func main() {
	if err := run(); err != nil {
		log.Printf("application stopped with error: %v", err)
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

	postgresPool, err :=
		database.OpenPostgres(
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

	defer postgresPool.Close()

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

	loginTransactionStore, err :=
		auth.NewRedisLoginTransactionStore(
			redisClient,
			cfg.RedisKeyPrefix,
		)
	if err != nil {
		return err
	}

	authHandler, err := auth.NewHandler(
		oidcClient,
		loginTransactionStore,
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
	)

	httpServer := appServer.HTTPServer()

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf(
			"CRM API listening on %s",
			cfg.HTTPAddress,
		)

		serverErrors <- httpServer.ListenAndServe()
	}()

	shutdownSignal := make(chan os.Signal, 1)

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
		if !errors.Is(serverErr, http.ErrServerClosed) {
			return serverErr
		}
	}

	shutdownContext, shutdownCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownContext); err != nil {
		return err
	}

	log.Println("CRM API stopped cleanly")

	return nil
}
