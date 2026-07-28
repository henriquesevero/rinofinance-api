// Package user is the bounded context responsible for account identity and
// authentication. Every other bounded context references a User only by ID
// (UserID), never by importing this package's entity directly, keeping
// aggregates decoupled as required by DDD.
package user

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"rinofinance-api/internal/domain/shared"
)

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// User is the aggregate root for authentication and account ownership.
// PasswordHash always stores a bcrypt (or equivalent) hash — the domain
// layer never sees or handles plaintext passwords.
type User struct {
	ID           uuid.UUID
	Name         string
	Email        string
	PasswordHash string
	// AvatarURL holds the user's profile picture as a data: URL (the
	// frontend downsizes the image client-side before sending it), so no
	// separate object storage is needed for this version. It is empty
	// until the user uploads one, in which case the UI falls back to
	// initials.
	AvatarURL string
	// DataOwnerID, when set, points at another user whose data this account
	// sees and edits (shared household). Nil means the account uses its own
	// data. Auth/profile always act on the real user, never the owner.
	DataOwnerID *uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ShareDataWith makes this user see/edit ownerID's data.
func (u *User) ShareDataWith(ownerID uuid.UUID) {
	u.DataOwnerID = &ownerID
	u.UpdatedAt = time.Now().UTC()
}

// StopSharing reverts the account to using its own data.
func (u *User) StopSharing() {
	u.DataOwnerID = nil
	u.UpdatedAt = time.Now().UTC()
}

// NewUser builds a new User aggregate, validating the invariants that must
// hold for any account regardless of which adapter creates it.
func NewUser(name, email, passwordHash string) (*User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, shared.ErrEmptyName
	}

	email = strings.TrimSpace(strings.ToLower(email))
	if !emailPattern.MatchString(email) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidEmail, email)
	}

	if passwordHash == "" {
		return nil, ErrEmptyPasswordHash
	}

	now := time.Now().UTC()
	return &User{
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// Rename updates the display name of the user.
func (u *User) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return shared.ErrEmptyName
	}
	u.Name = name
	u.UpdatedAt = time.Now().UTC()
	return nil
}

// ChangePasswordHash replaces the stored password hash after the caller
// (an application-layer use case) has already hashed the new plaintext
// password.
func (u *User) ChangePasswordHash(newHash string) error {
	if newHash == "" {
		return ErrEmptyPasswordHash
	}
	u.PasswordHash = newHash
	u.UpdatedAt = time.Now().UTC()
	return nil
}

// ChangeEmail updates the account's login email.
func (u *User) ChangeEmail(email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if !emailPattern.MatchString(email) {
		return fmt.Errorf("%w: %q", ErrInvalidEmail, email)
	}
	u.Email = email
	u.UpdatedAt = time.Now().UTC()
	return nil
}

// UpdateAvatar replaces the profile picture. Passing an empty string
// clears it, falling back to initials in the UI.
func (u *User) UpdateAvatar(avatarURL string) {
	u.AvatarURL = avatarURL
	u.UpdatedAt = time.Now().UTC()
}
