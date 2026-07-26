// Package handler implements the primary (driving) HTTP adapter: one
// handler per resource, translating requests into application-layer use
// case calls and domain errors into HTTP status codes.
package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	appauth "rinofinance-api/internal/application/auth"
	domaincard "rinofinance-api/internal/domain/card"
	domainexpense "rinofinance-api/internal/domain/expense"
	"rinofinance-api/internal/domain/shared"
	domainuser "rinofinance-api/internal/domain/user"
	"rinofinance-api/internal/pkg/auth"
)

// writeJSON encodes payload as the JSON response body with the given
// status code.
func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

// decodeJSON parses the request body into dst, returning a 400-mappable
// error on malformed JSON.
func decodeJSON(r *http.Request, dst interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return errBadRequest
	}
	return nil
}

// errBadRequest marks a request body that failed to parse as JSON.
var errBadRequest = errors.New("corpo da requisição inválido")

// writeError maps a domain/application error to the appropriate HTTP
// status and writes it as {"error": "..."}. Unexpected (unmapped) errors
// are logged server-side and returned as a generic 500 so internal detail
// is never leaked to the client.
func writeError(w http.ResponseWriter, err error) {
	status := statusForError(err)
	if status == http.StatusInternalServerError {
		log.Printf("internal error: %v", err)
		writeJSON(w, status, map[string]string{"error": "erro interno do servidor"})
		return
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func statusForError(err error) int {
	switch {
	case errors.Is(err, errBadRequest):
		return http.StatusBadRequest

	case errors.Is(err, shared.ErrNotFound):
		return http.StatusNotFound

	case errors.Is(err, domainuser.ErrInvalidCredentials):
		return http.StatusUnauthorized

	case errors.Is(err, shared.ErrUnauthorized),
		errors.Is(err, appauth.ErrInvalidRegistrationCode):
		return http.StatusForbidden

	case errors.Is(err, shared.ErrEmptyName),
		errors.Is(err, shared.ErrInvalidAmount),
		errors.Is(err, shared.ErrNegativeAmount),
		errors.Is(err, domaincard.ErrInvalidInstallmentCount),
		errors.Is(err, domaincard.ErrInvalidFirstInstallmentDate),
		errors.Is(err, domainexpense.ErrAmountManagedByCard),
		errors.Is(err, domainexpense.ErrNotCardLinked),
		errors.Is(err, domainuser.ErrInvalidEmail),
		errors.Is(err, domainuser.ErrEmailAlreadyInUse),
		errors.Is(err, appauth.ErrPasswordTooShort),
		errors.Is(err, auth.ErrInvalidToken):
		return http.StatusUnprocessableEntity

	default:
		return http.StatusInternalServerError
	}
}
