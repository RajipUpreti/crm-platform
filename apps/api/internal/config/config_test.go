package config

import "testing"

func TestValidateHostPort(t *testing.T) {
	t.Parallel()

	if err := validateHostPort(
		"OIDC_DOCKER_KEYCLOAK_ADDRESS",
		"keycloak:8080",
	); err != nil {
		t.Fatalf("validateHostPort() error = %v", err)
	}
}

func TestValidateHostPortRejectsURL(t *testing.T) {
	t.Parallel()

	if err := validateHostPort(
		"OIDC_DOCKER_KEYCLOAK_ADDRESS",
		"http://keycloak:8080",
	); err == nil {
		t.Fatal("validateHostPort() accepted a URL")
	}
}
