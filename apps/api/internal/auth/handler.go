package auth

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

type HandlerConfig struct {
	FrontendURL             string
	LoginTransactionTTL     time.Duration
	DefaultLoginDestination string
}

type Handler struct {
	oidcClient       *OIDCClient
	transactionStore LoginTransactionStore

	identitySynchronizer IdentitySynchronizer

	frontendURL             *url.URL
	loginTransactionTTL     time.Duration
	defaultLoginDestination string
	now                     func() time.Time
}

func NewHandler(
	oidcClient *OIDCClient,
	transactionStore LoginTransactionStore,
	identitySynchronizer IdentitySynchronizer,
	cfg HandlerConfig,
) (*Handler, error) {
	if oidcClient == nil {
		return nil, fmt.Errorf("OIDC client is required")
	}

	if transactionStore == nil {
		return nil, fmt.Errorf(
			"login transaction store is required",
		)
	}

	if identitySynchronizer == nil {
		return nil, fmt.Errorf(
			"identity synchronizer is required",
		)
	}
	frontendURL, err := url.Parse(cfg.FrontendURL)
	if err != nil {
		return nil, fmt.Errorf(
			"parse frontend URL: %w",
			err,
		)
	}

	if frontendURL.Scheme == "" || frontendURL.Host == "" {
		return nil, fmt.Errorf(
			"frontend URL must be absolute",
		)
	}

	transactionTTL := cfg.LoginTransactionTTL
	if transactionTTL <= 0 {
		transactionTTL = 10 * time.Minute
	}

	defaultDestination := strings.TrimSpace(
		cfg.DefaultLoginDestination,
	)

	if defaultDestination == "" {
		defaultDestination = "/dashboard"
	}

	if !isSafeLocalPath(defaultDestination) {
		return nil, fmt.Errorf(
			"default login destination must be a local path",
		)
	}

	return &Handler{
		oidcClient:              oidcClient,
		transactionStore:        transactionStore,
		identitySynchronizer:    identitySynchronizer,
		frontendURL:             frontendURL,
		loginTransactionTTL:     transactionTTL,
		defaultLoginDestination: defaultDestination,
		now:                     time.Now,
	}, nil
}

func isSafeLocalPath(value string) bool {
	if value == "" {
		return false
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}

	return parsed.IsAbs() == false &&
		parsed.Host == "" &&
		strings.HasPrefix(parsed.Path, "/") &&
		!strings.HasPrefix(parsed.Path, "//")
}
