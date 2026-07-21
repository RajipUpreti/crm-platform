package middleware

import (
	"net/http"
	"strings"
)

type CORS struct {
	allowedOrigins map[string]struct{}
}

func NewCORS(
	origins []string,
) *CORS {
	allowedOrigins := make(
		map[string]struct{},
		len(origins),
	)

	for _, origin := range origins {
		origin = strings.TrimSpace(origin)

		if origin != "" {
			allowedOrigins[origin] = struct{}{}
		}
	}

	return &CORS{
		allowedOrigins: allowedOrigins,
	}
}

func (m *CORS) Wrap(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			origin := r.Header.Get("Origin")

			if _, allowed := m.allowedOrigins[origin]; allowed {
				w.Header().Set(
					"Access-Control-Allow-Origin",
					origin,
				)

				w.Header().Set(
					"Access-Control-Allow-Credentials",
					"true",
				)

				w.Header().Set(
					"Access-Control-Allow-Headers",
					"Content-Type, Accept",
				)

				w.Header().Set(
					"Access-Control-Allow-Methods",
					"GET, POST, PATCH, DELETE, OPTIONS",
				)

				w.Header().Add(
					"Vary",
					"Origin",
				)
			}

			if r.Method ==
				http.MethodOptions {
				w.WriteHeader(
					http.StatusNoContent,
				)
				return
			}

			next.ServeHTTP(w, r)
		},
	)
}
