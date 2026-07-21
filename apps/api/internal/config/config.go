package config

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnvironment     string
	HTTPAddress        string
	FrontendURL        string
	CORSAllowedOrigins []string

	DatabaseURL string

	RedisAddress   string
	RedisPassword  string
	RedisDatabase  int
	RedisKeyPrefix string

	OIDCIssuerURL             string
	OIDCClientID              string
	OIDCClientSecret          string
	OIDCRedirectURL           string
	OIDCDockerKeycloakAddress string

	Session SessionConfig
}

type SessionConfig struct {
	TTL            time.Duration
	CookieName     string
	CookieDomain   string
	CookieSecure   bool
	CookieSameSite http.SameSite
	RedisKeyPrefix string
}

func Load() (Config, error) {
	redisDatabase, err := readInt(
		"REDIS_DATABASE",
		0,
	)
	if err != nil {
		return Config{}, err
	}

	sessionTTL, err := readDuration(
		"SESSION_TTL",
		24*time.Hour,
	)
	if err != nil {
		return Config{}, err
	}

	sessionCookieSecure, err := readBool(
		"SESSION_COOKIE_SECURE",
		false,
	)
	if err != nil {
		return Config{}, err
	}

	sessionCookieSameSite, err := parseSameSite(
		envOrDefault(
			"SESSION_COOKIE_SAME_SITE",
			"lax",
		),
	)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		AppEnvironment: strings.TrimSpace(
			envOrDefault(
				"APP_ENV",
				"development",
			),
		),
		HTTPAddress: strings.TrimSpace(
			envOrDefault(
				"HTTP_ADDRESS",
				":8080",
			),
		),
		FrontendURL: strings.TrimSpace(
			envOrDefault(
				"FRONTEND_URL",
				"http://localhost:3000",
			),
		),
		CORSAllowedOrigins: strings.Split(
			envOrDefault(
				"CORS_ALLOWED_ORIGINS",
				"http://localhost:3000",
			),
			",",
		),

		DatabaseURL: strings.TrimSpace(
			os.Getenv("DATABASE_URL"),
		),

		RedisAddress: strings.TrimSpace(
			envOrDefault(
				"REDIS_ADDRESS",
				"redis:6379",
			),
		),
		RedisPassword: os.Getenv(
			"REDIS_PASSWORD",
		),
		RedisDatabase: redisDatabase,
		RedisKeyPrefix: strings.TrimSpace(
			envOrDefault(
				"REDIS_KEY_PREFIX",
				"crm:development",
			),
		),

		OIDCIssuerURL: strings.TrimSpace(
			os.Getenv("OIDC_ISSUER_URL"),
		),
		OIDCClientID: strings.TrimSpace(
			os.Getenv("OIDC_CLIENT_ID"),
		),
		OIDCClientSecret: os.Getenv(
			"OIDC_CLIENT_SECRET",
		),
		OIDCRedirectURL: strings.TrimSpace(
			os.Getenv("OIDC_REDIRECT_URL"),
		),
		OIDCDockerKeycloakAddress: strings.TrimSpace(
			os.Getenv(
				"OIDC_DOCKER_KEYCLOAK_ADDRESS",
			),
		),

		Session: SessionConfig{
			TTL: sessionTTL,
			CookieName: strings.TrimSpace(
				envOrDefault(
					"SESSION_COOKIE_NAME",
					"crm_session",
				),
			),
			CookieDomain: strings.TrimSpace(
				os.Getenv(
					"SESSION_COOKIE_DOMAIN",
				),
			),
			CookieSecure:   sessionCookieSecure,
			CookieSameSite: sessionCookieSameSite,
			RedisKeyPrefix: strings.TrimSpace(
				envOrDefault(
					"REDIS_SESSION_KEY_PREFIX",
					"crm:development:session",
				),
			),
		},
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	requiredValues := map[string]string{
		"DATABASE_URL":       c.DatabaseURL,
		"REDIS_ADDRESS":      c.RedisAddress,
		"REDIS_KEY_PREFIX":   c.RedisKeyPrefix,
		"OIDC_ISSUER_URL":    c.OIDCIssuerURL,
		"OIDC_CLIENT_ID":     c.OIDCClientID,
		"OIDC_CLIENT_SECRET": c.OIDCClientSecret,
		"OIDC_REDIRECT_URL":  c.OIDCRedirectURL,
		"FRONTEND_URL":       c.FrontendURL,
	}

	for name, value := range requiredValues {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf(
				"%s is required",
				name,
			)
		}
	}

	if strings.TrimSpace(c.HTTPAddress) == "" {
		return fmt.Errorf(
			"HTTP_ADDRESS is required",
		)
	}

	if c.RedisDatabase < 0 {
		return fmt.Errorf(
			"REDIS_DATABASE cannot be negative",
		)
	}

	if err := validateAbsoluteHTTPURL(
		"FRONTEND_URL",
		c.FrontendURL,
	); err != nil {
		return err
	}

	if err := validateAbsoluteHTTPURL(
		"OIDC_ISSUER_URL",
		c.OIDCIssuerURL,
	); err != nil {
		return err
	}

	if err := validateAbsoluteHTTPURL(
		"OIDC_REDIRECT_URL",
		c.OIDCRedirectURL,
	); err != nil {
		return err
	}

	if c.OIDCDockerKeycloakAddress != "" {
		if err := validateHostPort(
			"OIDC_DOCKER_KEYCLOAK_ADDRESS",
			c.OIDCDockerKeycloakAddress,
		); err != nil {
			return err
		}
	}

	if c.Session.TTL <= 0 {
		return fmt.Errorf(
			"SESSION_TTL must be positive",
		)
	}

	if c.Session.CookieName == "" {
		return fmt.Errorf(
			"SESSION_COOKIE_NAME is required",
		)
	}

	if strings.ContainsAny(
		c.Session.CookieName,
		" \t\r\n;=,",
	) {
		return fmt.Errorf(
			"SESSION_COOKIE_NAME contains invalid characters",
		)
	}

	if c.Session.RedisKeyPrefix == "" {
		return fmt.Errorf(
			"REDIS_SESSION_KEY_PREFIX is required",
		)
	}

	if c.Session.CookieSameSite ==
		http.SameSiteNoneMode &&
		!c.Session.CookieSecure {
		return fmt.Errorf(
			"SESSION_COOKIE_SAME_SITE=none requires SESSION_COOKIE_SECURE=true",
		)
	}

	if strings.HasPrefix(
		c.Session.CookieName,
		"__Host-",
	) {
		if !c.Session.CookieSecure {
			return fmt.Errorf(
				"__Host- session cookies require SESSION_COOKIE_SECURE=true",
			)
		}

		if c.Session.CookieDomain != "" {
			return fmt.Errorf(
				"__Host- session cookies must not define SESSION_COOKIE_DOMAIN",
			)
		}
	}

	if c.AppEnvironment == "production" &&
		!c.Session.CookieSecure {
		return fmt.Errorf(
			"SESSION_COOKIE_SECURE must be true in production",
		)
	}

	return nil
}

func parseSameSite(
	value string,
) (http.SameSite, error) {
	switch strings.ToLower(
		strings.TrimSpace(value),
	) {
	case "", "lax":
		return http.SameSiteLaxMode, nil

	case "strict":
		return http.SameSiteStrictMode, nil

	case "none":
		return http.SameSiteNoneMode, nil

	default:
		return http.SameSiteDefaultMode, fmt.Errorf(
			"SESSION_COOKIE_SAME_SITE must be lax, strict, or none",
		)
	}
}

func validateAbsoluteHTTPURL(
	name string,
	value string,
) error {
	parsed, err := url.Parse(
		strings.TrimSpace(value),
	)
	if err != nil {
		return fmt.Errorf(
			"%s must be a valid URL: %w",
			name,
			err,
		)
	}

	if parsed.Scheme == "" ||
		parsed.Host == "" {
		return fmt.Errorf(
			"%s must be an absolute URL",
			name,
		)
	}

	if parsed.Scheme != "http" &&
		parsed.Scheme != "https" {
		return fmt.Errorf(
			"%s must use http or https",
			name,
		)
	}

	return nil
}

func validateHostPort(
	name string,
	value string,
) error {
	value = strings.TrimSpace(value)
	if strings.Contains(value, "://") {
		return fmt.Errorf(
			"%s must use host:port format, not a URL",
			name,
		)
	}

	host, port, err := net.SplitHostPort(
		value,
	)
	if err != nil || host == "" || port == "" {
		return fmt.Errorf(
			"%s must use host:port format",
			name,
		)
	}

	return nil
}

func envOrDefault(
	name string,
	fallback string,
) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}

	return value
}

func readBool(
	name string,
	fallback bool,
) (bool, error) {
	value := strings.TrimSpace(
		os.Getenv(name),
	)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf(
			"%s must be a boolean: %w",
			name,
			err,
		)
	}

	return parsed, nil
}

func readInt(
	name string,
	fallback int,
) (int, error) {
	value := strings.TrimSpace(
		os.Getenv(name),
	)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf(
			"%s must be an integer: %w",
			name,
			err,
		)
	}

	return parsed, nil
}

func readDuration(
	name string,
	fallback time.Duration,
) (time.Duration, error) {
	value := strings.TrimSpace(
		os.Getenv(name),
	)
	if value == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf(
			"%s must be a valid duration such as 30m, 24h, or 168h: %w",
			name,
			err,
		)
	}

	return parsed, nil
}
