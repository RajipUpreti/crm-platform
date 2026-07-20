package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func GenerateRandomValue(byteLength int) (string, error) {
	if byteLength < 16 {
		return "", fmt.Errorf(
			"random value length must be at least 16 bytes",
		)
	}

	value := make([]byte, byteLength)

	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf(
			"read cryptographically secure random bytes: %w",
			err,
		)
	}

	return base64.RawURLEncoding.EncodeToString(value), nil
}
