package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"rinofinance-api/internal/adapters/http/middleware"
)

// parseOptionalUUID parses an optional UUID string from a request payload:
// an empty string means "not set" (nil), any other value must be a valid
// UUID. Used for nullable references like an item's category.
func parseOptionalUUID(raw string) (*uuid.UUID, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// errNoAuthenticatedUser indicates UserIDFromContext found nothing, which
// should never happen for a route wrapped in middleware.Auth — its
// presence would signal a routing/wiring bug, not bad client input.
var errNoAuthenticatedUser = errors.New("nenhum usuário autenticado no contexto da requisição")

// requireUserID extracts the authenticated user's ID set by
// middleware.Auth, writing a 500 (via writeError as an internal error) if
// somehow missing.
func requireUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, errNoAuthenticatedUser)
		return uuid.Nil, false
	}
	return userID, true
}

// requireAuthUserID returns the real authenticated user's id (not the shared
// data owner) — used by profile/account actions that must target the real
// account.
func requireAuthUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	userID, ok := middleware.AuthUserIDFromContext(r.Context())
	if !ok {
		writeError(w, errNoAuthenticatedUser)
		return uuid.Nil, false
	}
	return userID, true
}

// parseUUIDPathValue reads a {name} path value (Go 1.22+ ServeMux
// pattern) as a uuid.UUID, writing a 404-mappable error on failure so a
// malformed ID behaves the same as a missing resource.
func parseUUIDPathValue(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue(name))
	if err != nil {
		writeError(w, errBadRequest)
		return uuid.Nil, false
	}
	return id, true
}

// parseUUIDList parses a slice of string IDs into uuid.UUIDs, returning an
// error on the first malformed entry.
func parseUUIDList(raw []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(raw))
	for _, s := range raw {
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// referenceMonthLayout is the wire format for the ?month= query param
// (e.g. "2026-07").
const referenceMonthLayout = "2006-01"

// parseReferenceMonth reads the ?month=YYYY-MM query parameter, defaulting
// to the current month (UTC) when absent, for every endpoint whose totals
// depend on "which month" is being viewed.
func parseReferenceMonth(r *http.Request) (time.Time, error) {
	raw := r.URL.Query().Get("month")
	if raw == "" {
		now := time.Now().UTC()
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC), nil
	}
	parsed, err := time.Parse(referenceMonthLayout, raw)
	if err != nil {
		return time.Time{}, errBadRequest
	}
	return parsed, nil
}
