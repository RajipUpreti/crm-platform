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

	appServer := server.New(
		cfg.HTTPAddress,
		oidcClient,
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
