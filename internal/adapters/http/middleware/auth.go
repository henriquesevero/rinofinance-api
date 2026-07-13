// Package middleware holds cross-cutting HTTP concerns: JWT
// authentication and CORS. Neither depends on any specific resource
// handler, so they can wrap any route in router.go.
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"rinofinance-api/internal/pkg/auth"
)

type contextKey int

const userIDContextKey contextKey = iota

// Auth returns middleware that requires a valid "Authorization: Bearer
// <token>" header, parses the user ID out of it via tokens, and stores it
// in the request context for handlers to read with UserIDFromContext.
func Auth(tokens *auth.TokenIssuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			const prefix = "Bearer "
			if !strings.HasPrefix(header, prefix) {
				writeUnauthorized(w, "token de autenticação ausente")
				return
			}

			userID, err := tokens.Parse(strings.TrimPrefix(header, prefix))
			if err != nil {
				writeUnauthorized(w, "token inválido ou expirado")
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext extracts the authenticated user's ID, set by Auth.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDContextKey).(uuid.UUID)
	return id, ok
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
