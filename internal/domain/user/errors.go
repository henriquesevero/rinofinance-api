package user

import "errors"

var (
	// ErrInvalidEmail indicates the provided email does not match a basic
	// well-formed email pattern.
	ErrInvalidEmail = errors.New("email inválido")

	// ErrEmptyPasswordHash indicates a user was constructed without a
	// password hash. The domain never accepts plaintext passwords, so this
	// signals a bug in the calling use case, not bad user input.
	ErrEmptyPasswordHash = errors.New("hash de senha não pode ser vazio")

	// ErrEmailAlreadyInUse indicates the Postgres adapter found a unique
	// constraint violation on email during Create.
	ErrEmailAlreadyInUse = errors.New("email já cadastrado")

	// ErrInvalidCredentials indicates a login attempt with an unknown email
	// or a password that does not match the stored hash.
	ErrInvalidCredentials = errors.New("credenciais inválidas")
)
