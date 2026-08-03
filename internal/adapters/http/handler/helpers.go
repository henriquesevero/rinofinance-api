package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"rinofinance-api/internal/adapters/http/middleware"
)

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

var errNoAuthenticatedUser = errors.New("nenhum usuário autenticado no contexto da requisição")

func requireUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, errNoAuthenticatedUser)
		return uuid.Nil, false
	}
	return userID, true
}

func requireAuthUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	userID, ok := middleware.AuthUserIDFromContext(r.Context())
	if !ok {
		writeError(w, errNoAuthenticatedUser)
		return uuid.Nil, false
	}
	return userID, true
}

func parseUUIDPathValue(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue(name))
	if err != nil {
		writeError(w, errBadRequest)
		return uuid.Nil, false
	}
	return id, true
}

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

const referenceMonthLayout = "2006-01"

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
