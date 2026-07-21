package session

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type CookieConfig struct {
	Name     string
	Path     string
	Domain   string
	Secure   bool
	SameSite http.SameSite
}

type CookieManager struct {
	name     string
	path     string
	domain   string
	secure   bool
	sameSite http.SameSite
}

func NewCookieManager(
	cfg CookieConfig,
) (*CookieManager, error) {
	name := strings.TrimSpace(cfg.Name)
	if name == "" {
		return nil, fmt.Errorf(
			"session cookie name is required",
		)
	}

	path := strings.TrimSpace(cfg.Path)
	if path == "" {
		path = "/"
	}

	return &CookieManager{
		name:     name,
		path:     path,
		domain:   strings.TrimSpace(cfg.Domain),
		secure:   cfg.Secure,
		sameSite: cfg.SameSite,
	}, nil
}

func (m *CookieManager) Set(
	w http.ResponseWriter,
	token string,
	expiresAt time.Time,
) {
	maxAge := int(
		time.Until(expiresAt).Seconds(),
	)

	if maxAge < 1 {
		maxAge = 1
	}

	http.SetCookie(
		w,
		&http.Cookie{
			Name:     m.name,
			Value:    token,
			Path:     m.path,
			Domain:   m.domain,
			Expires:  expiresAt,
			MaxAge:   maxAge,
			Secure:   m.secure,
			HttpOnly: true,
			SameSite: m.sameSite,
		},
	)
}

func (m *CookieManager) Clear(
	w http.ResponseWriter,
) {
	http.SetCookie(
		w,
		&http.Cookie{
			Name:     m.name,
			Value:    "",
			Path:     m.path,
			Domain:   m.domain,
			Expires:  time.Unix(1, 0),
			MaxAge:   -1,
			Secure:   m.secure,
			HttpOnly: true,
			SameSite: m.sameSite,
		},
	)
}

func (m *CookieManager) Read(
	r *http.Request,
) (string, error) {
	cookie, err := r.Cookie(m.name)
	if err != nil {
		return "", ErrNotFound
	}

	token := strings.TrimSpace(cookie.Value)
	if token == "" {
		return "", ErrInvalid
	}

	return token, nil
}
