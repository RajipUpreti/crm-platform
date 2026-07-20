package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppEnvironment string
	HTTPAddress    string
	FrontendURL    string

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

	SessionCookieName string
	SessionSecure     bool
}

func readInt(
	name string,
	fallback int,
) (int, error) {
	value := os.Getenv(name)
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

func Load() (Config, error) {
	sessionSecure, err := readBool(
		"SESSION_SECURE",
		false,
	)
	if err != nil {
		return Config{}, err
	}

	redisDatabase, err := readInt(
		"REDIS_DATABASE",
		0,
	)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		AppEnvironment: envOrDefault(
			"APP_ENV",
			"development",
		),
		HTTPAddress: envOrDefault(
			"HTTP_ADDRESS",
			":8080",
		),
		FrontendURL: envOrDefault(
			"FRONTEND_URL",
			"http://localhost:3000",
		),

		DatabaseURL: os.Getenv("DATABASE_URL"),

		RedisAddress: envOrDefault(
			"REDIS_ADDRESS",
			"redis:6379",
		),
		RedisPassword: os.Getenv(
			"REDIS_PASSWORD",
		),
		RedisDatabase: redisDatabase,
		RedisKeyPrefix: envOrDefault(
			"REDIS_KEY_PREFIX",
			"crm:development",
		),

		OIDCIssuerURL: os.Getenv(
			"OIDC_ISSUER_URL",
		),
		OIDCClientID: os.Getenv(
			"OIDC_CLIENT_ID",
		),
		OIDCClientSecret: os.Getenv(
			"OIDC_CLIENT_SECRET",
		),
		OIDCRedirectURL: os.Getenv(
			"OIDC_REDIRECT_URL",
		),
		OIDCDockerKeycloakAddress: os.Getenv(
			"OIDC_DOCKER_KEYCLOAK_ADDRESS",
		),

		SessionCookieName: envOrDefault(
			"SESSION_COOKIE_NAME",
			"crm_session",
		),
		SessionSecure: sessionSecure,
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
		if value == "" {
			return fmt.Errorf("%s is required", name)
		}
	}

	if c.RedisDatabase < 0 {
		return fmt.Errorf(
			"REDIS_DATABASE cannot be negative",
		)
	}

	return nil
}

func envOrDefault(name string, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}

	return value
}

func readBool(name string, fallback bool) (bool, error) {
	value := os.Getenv(name)
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
