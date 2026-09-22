// Package middleware holds the application HTTP middlewares.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"miyano/internal/ports"
)

type contextKey int

const userIDKey contextKey = iota

// UserID returns the authenticated user id injected by RequireAuth,
// or an empty string when the request is unauthenticated.
func UserID(r *http.Request) string {
	id, _ := r.Context().Value(userIDKey).(string)
	return id
}

// RequireAuth rejects requests without a valid Bearer access token with a
// generic 401 (identical body for missing, malformed, invalid and expired
// tokens — no information leaks) and injects the user id into the context.
func RequireAuth(verifier ports.TokenVerifier) func(http.Handler) http.Handler {
	const prefix = "Bearer "

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, prefix) {
				unauthorized(w)
				return
			}

			userID, err := verifier.Verify(strings.TrimPrefix(header, prefix), ports.TokenTypeAccess)
			if err != nil {
				unauthorized(w)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"missing or invalid token"}` + "\n"))
}
