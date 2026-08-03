package shared

import "errors"

var (
	ErrNotFound = errors.New("registro não encontrado")

	ErrEmptyName = errors.New("nome não pode ser vazio")

	ErrInvalidAmount = errors.New("valor monetário inválido")

	ErrNegativeAmount = errors.New("valor não pode ser negativo")

	ErrUnauthorized = errors.New("usuário não autorizado a acessar este recurso")
)
