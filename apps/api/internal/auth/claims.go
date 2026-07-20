package auth

import (
	"fmt"
	"strings"

	"github.com/rajipupreti/crm-platform/apps/api/internal/user"
)

type IdentityClaims struct {
	Subject           string `json:"sub"`
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	Name              string `json:"name"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	PreferredUsername string `json:"preferred_username"`
}
type AuthenticatedIdentityResponse struct {
	User     user.User             `json:"user"`
	Identity AuthenticatedIdentity `json:"identity"`
	ReturnTo string                `json:"returnTo"`
}

type AuthenticatedIdentity struct {
	Subject           string `json:"subject"`
	Email             string `json:"email"`
	EmailVerified     bool   `json:"emailVerified"`
	Name              string `json:"name,omitempty"`
	GivenName         string `json:"givenName,omitempty"`
	FamilyName        string `json:"familyName,omitempty"`
	PreferredUsername string `json:"preferredUsername,omitempty"`
}

func (c IdentityClaims) Validate() error {
	if strings.TrimSpace(c.Subject) == "" {
		return fmt.Errorf(
			"%w: subject is required",
			ErrInvalidIdentityClaims,
		)
	}

	if strings.TrimSpace(c.Email) == "" {
		return fmt.Errorf(
			"%w: email is required",
			ErrInvalidIdentityClaims,
		)
	}

	return nil
}
func (c IdentityClaims) Response() AuthenticatedIdentity {
	return AuthenticatedIdentity{
		Subject:           c.Subject,
		Email:             c.Email,
		EmailVerified:     c.EmailVerified,
		Name:              c.Name,
		GivenName:         c.GivenName,
		FamilyName:        c.FamilyName,
		PreferredUsername: c.PreferredUsername,
	}
}
