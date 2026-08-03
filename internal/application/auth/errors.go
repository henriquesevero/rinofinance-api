package auth

import "errors"

var ErrPasswordTooShort = errors.New("senha deve ter ao menos 8 caracteres")

var ErrInvalidRegistrationCode = errors.New("código de convite inválido")
