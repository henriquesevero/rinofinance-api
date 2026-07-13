package shared

import "errors"

// Sentinel errors shared by every bounded context. Use case and adapter
// layers can map these to HTTP status codes (e.g. ErrNotFound -> 404,
// the validation errors -> 422) with errors.Is, without the domain layer
// knowing anything about HTTP.
var (
	// ErrNotFound indicates the requested aggregate does not exist or does
	// not belong to the requesting user.
	ErrNotFound = errors.New("registro não encontrado")

	// ErrEmptyName indicates a required name/label field was blank.
	ErrEmptyName = errors.New("nome não pode ser vazio")

	// ErrInvalidAmount indicates a monetary value could not be parsed or
	// violates a domain invariant (e.g. negative where not allowed).
	ErrInvalidAmount = errors.New("valor monetário inválido")

	// ErrNegativeAmount indicates a monetary value must be zero or positive.
	ErrNegativeAmount = errors.New("valor não pode ser negativo")

	// ErrUnauthorized indicates the acting user does not own the aggregate
	// being mutated.
	ErrUnauthorized = errors.New("usuário não autorizado a acessar este recurso")
)
