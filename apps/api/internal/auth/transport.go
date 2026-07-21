package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

type TransportConfig struct {
	PublicAddress string
	DockerAddress string
}

func NewDockerAwareHTTPClient(
	cfg TransportConfig,
) (*http.Client, error) {
	publicHost, err := normalizeAddress(cfg.PublicAddress)
	if err != nil {
		return nil, fmt.Errorf(
			"normalize public OIDC address: %w",
			err,
		)
	}

	dockerHost, err := normalizeAddress(cfg.DockerAddress)
	if err != nil {
		return nil, fmt.Errorf(
			"normalize Docker OIDC address: %w",
			err,
		)
	}

	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()

	transport.DialContext = func(
		ctx context.Context,
		network string,
		address string,
	) (net.Conn, error) {
		if address == publicHost {
			address = dockerHost
		}

		return dialer.DialContext(ctx, network, address)
	}

	return &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}, nil
}

func normalizeAddress(value string) (string, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return "", fmt.Errorf("address is empty")
	}
	if strings.Contains(value, "://") {
		return "", fmt.Errorf(
			"address %q must not be a URL",
			value,
		)
	}

	host, port, err := net.SplitHostPort(value)
	if err != nil || host == "" || port == "" {
		return "", fmt.Errorf(
			"address %q must use host:port format",
			value,
		)
	}

	return value, nil
}
