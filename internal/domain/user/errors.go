package user

import "errors"

var (
	ErrInvalidEmail = errors.New("email inválido")

	ErrEmptyPasswordHash = errors.New("hash de senha não pode ser vazio")

	ErrEmailAlreadyInUse = errors.New("email já cadastrado")

	ErrInvalidCredentials = errors.New("credenciais inválidas")
)
