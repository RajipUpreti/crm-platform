package session

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

const sessionTokenBytes = 32

func generateToken() (string, error) {
	randomBytes := make(
		[]byte,
		sessionTokenBytes,
	)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf(
			"generate session token: %w",
			err,
		)
	}

	return base64.RawURLEncoding.EncodeToString(
		randomBytes,
	), nil
}

func tokenDigest(token string) string {
	digest := sha256.Sum256([]byte(token))

	return hex.EncodeToString(digest[:])
}
