package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type OIDCConfig struct {
	IssuerURL             string
	ClientID              string
	ClientSecret          string
	RedirectURL           string
	DockerKeycloakAddress string
}

type OIDCClient struct {
	Provider     *oidc.Provider
	Verifier     *oidc.IDTokenVerifier
	OAuth2Config oauth2.Config
	HTTPClient   *http.Client

	IssuerURL     string
	EndSessionURL string
}

type providerMetadata struct {
	EndSessionEndpoint string `json:"end_session_endpoint"`
}

func NewOIDCClient(
	ctx context.Context,
	cfg OIDCConfig,
) (*OIDCClient, error) {
	if err := validateOIDCConfig(cfg); err != nil {
		return nil, err
	}

	issuerURL, err := url.Parse(cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("parse OIDC issuer URL: %w", err)
	}

	if issuerURL.Scheme == "" || issuerURL.Host == "" {
		return nil, fmt.Errorf(
			"OIDC issuer URL must be absolute",
		)
	}

	httpClient, err := NewDockerAwareHTTPClient(
		TransportConfig{
			PublicAddress: issuerURL.Host,
			DockerAddress: cfg.DockerKeycloakAddress,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create OIDC HTTP client: %w",
			err,
		)
	}

	discoveryContext := oidc.ClientContext(ctx, httpClient)

	provider, err := oidc.NewProvider(
		discoveryContext,
		strings.TrimRight(cfg.IssuerURL, "/"),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"discover OIDC provider: %w",
			err,
		)
	}

	var metadata providerMetadata

	if err := provider.Claims(&metadata); err != nil {
		return nil, fmt.Errorf(
			"read OIDC provider metadata: %w",
			err,
		)
	}

	oauthConfig := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes: []string{
			oidc.ScopeOpenID,
			oidc.ScopeProfile,
			oidc.ScopeEmail,
		},
	}

	verifier := provider.VerifierContext(
		discoveryContext,
		&oidc.Config{
			ClientID: cfg.ClientID,
		},
	)

	return &OIDCClient{
		Provider:     provider,
		Verifier:     verifier,
		OAuth2Config: oauthConfig,
		HTTPClient:   httpClient,

		IssuerURL:     strings.TrimRight(cfg.IssuerURL, "/"),
		EndSessionURL: metadata.EndSessionEndpoint,
	}, nil
}

func validateOIDCConfig(cfg OIDCConfig) error {
	required := map[string]string{
		"IssuerURL":             cfg.IssuerURL,
		"ClientID":              cfg.ClientID,
		"ClientSecret":          cfg.ClientSecret,
		"RedirectURL":           cfg.RedirectURL,
		"DockerKeycloakAddress": cfg.DockerKeycloakAddress,
	}

	for name, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf(
				"OIDC %s is required",
				name,
			)
		}
	}

	return nil
}
