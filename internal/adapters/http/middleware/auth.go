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

const (
	userIDContextKey contextKey = iota
	authUserIDContextKey
)

func Auth(tokens *auth.TokenIssuer, resolveOwner func(uuid.UUID) uuid.UUID) func(http.Handler) http.Handler {
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

			owner := userID
			if resolveOwner != nil {
				owner = resolveOwner(userID)
			}
			ctx := context.WithValue(r.Context(), authUserIDContextKey, userID)
			ctx = context.WithValue(ctx, userIDContextKey, owner)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDContextKey).(uuid.UUID)
	return id, ok
}

func AuthUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(authUserIDContextKey).(uuid.UUID)
	return id, ok
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
