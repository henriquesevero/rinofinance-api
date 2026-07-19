package auth

import "errors"

// ErrPasswordTooShort indicates a registration attempt with a password
// below the minimum acceptable length.
var ErrPasswordTooShort = errors.New("senha deve ter ao menos 8 caracteres")

// ErrInvalidRegistrationCode indicates a sign-up attempt with a missing or
// wrong invite code, when one is required.
var ErrInvalidRegistrationCode = errors.New("código de convite inválido")
