package auth

import "errors"

// ErrPasswordTooShort indicates a registration attempt with a password
// below the minimum acceptable length.
var ErrPasswordTooShort = errors.New("senha deve ter ao menos 8 caracteres")
