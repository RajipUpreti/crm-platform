package auth

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/rajipupreti/crm-platform/apps/api/internal/session"
)

type HandlerConfig struct {
	FrontendURL             string
	LoginTransactionTTL     time.Duration
	DefaultLoginDestination string
}

type Handler struct {
	oidcClient           *OIDCClient
	transactionStore     LoginTransactionStore
	identitySynchronizer IdentitySynchronizer
	tenantOnboarder      TenantOnboarder

	sessionCreator       SessionCreator
	sessionDestroyer     SessionDestroyer
	sessionCookieManager *session.CookieManager

	frontendURL             *url.URL
	loginTransactionTTL     time.Duration
	defaultLoginDestination string
	now                     func() time.Time
}

func NewHandler(
	oidcClient *OIDCClient,
	transactionStore LoginTransactionStore,
	identitySynchronizer IdentitySynchronizer,
	tenantOnboarder TenantOnboarder,
	sessionCreator SessionCreator,
	sessionDestroyer SessionDestroyer,
	sessionCookieManager *session.CookieManager,
	cfg HandlerConfig,
) (*Handler, error) {
	if oidcClient == nil {
		return nil, fmt.Errorf(
			"OIDC client is required",
		)
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

	if sessionCreator == nil {
		return nil, fmt.Errorf(
			"session creator is required",
		)
	}

	if sessionCookieManager == nil {
		return nil, fmt.Errorf(
			"session cookie manager is required",
		)
	}

	if sessionDestroyer == nil {
		return nil, fmt.Errorf(
			"session destroyer is required",
		)
	}

	frontendURL, err := url.Parse(
		strings.TrimSpace(cfg.FrontendURL),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"parse frontend URL: %w",
			err,
		)
	}

	if frontendURL.Scheme == "" ||
		frontendURL.Host == "" {
		return nil, fmt.Errorf(
			"frontend URL must be absolute",
		)
	}

	if frontendURL.Scheme != "http" &&
		frontendURL.Scheme != "https" {
		return nil, fmt.Errorf(
			"frontend URL scheme must be http or https",
		)
	}
	if tenantOnboarder == nil {
		return nil, fmt.Errorf(
			"tenant onboarder is required",
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
			"default login destination must be a safe local path",
		)
	}

	return &Handler{
		oidcClient:              oidcClient,
		transactionStore:        transactionStore,
		identitySynchronizer:    identitySynchronizer,
		sessionCreator:          sessionCreator,
		sessionDestroyer:        sessionDestroyer,
		sessionCookieManager:    sessionCookieManager,
		frontendURL:             frontendURL,
		loginTransactionTTL:     transactionTTL,
		defaultLoginDestination: defaultDestination,
		tenantOnboarder:         tenantOnboarder,
		now:                     time.Now,
	}, nil
}

func isSafeLocalPath(
	value string,
) bool {
	value = strings.TrimSpace(value)

	if value == "" {
		return false
	}

	if strings.Contains(
		value,
		"\\",
	) {
		return false
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}

	return !parsed.IsAbs() &&
		parsed.Host == "" &&
		parsed.User == nil &&
		strings.HasPrefix(
			parsed.Path,
			"/",
		) &&
		!strings.HasPrefix(
			parsed.Path,
			"//",
		)
}

func (h *Handler) frontendDestination(
	returnTo string,
) (string, error) {
	returnTo = strings.TrimSpace(returnTo)
	if !isSafeLocalPath(returnTo) {
		return "", fmt.Errorf(
			"frontend destination must be a safe local path",
		)
	}

	localURL, err := url.Parse(returnTo)
	if err != nil {
		return "", fmt.Errorf(
			"parse frontend destination: %w",
			err,
		)
	}

	return h.frontendURL.ResolveReference(
		localURL,
	).String(), nil
}
