package utils

import (
	"net/http"
	"strings"
)

// AuthenticationMiddleware checks the Authorization header for a valid token.
func AuthenticationMiddleware(next http.Handler, token string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token != "" {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") || strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer ")) != token {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
