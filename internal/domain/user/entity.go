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

type User struct {
	ID           uuid.UUID
	Name         string
	Email        string
	PasswordHash string

	AvatarURL string

	DataOwnerID *uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (u *User) ShareDataWith(ownerID uuid.UUID) {
	u.DataOwnerID = &ownerID
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) StopSharing() {
	u.DataOwnerID = nil
	u.UpdatedAt = time.Now().UTC()
}

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

func (u *User) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return shared.ErrEmptyName
	}
	u.Name = name
	u.UpdatedAt = time.Now().UTC()
	return nil
}

func (u *User) ChangePasswordHash(newHash string) error {
	if newHash == "" {
		return ErrEmptyPasswordHash
	}
	u.PasswordHash = newHash
	u.UpdatedAt = time.Now().UTC()
	return nil
}

func (u *User) ChangeEmail(email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if !emailPattern.MatchString(email) {
		return fmt.Errorf("%w: %q", ErrInvalidEmail, email)
	}
	u.Email = email
	u.UpdatedAt = time.Now().UTC()
	return nil
}

func (u *User) UpdateAvatar(avatarURL string) {
	u.AvatarURL = avatarURL
	u.UpdatedAt = time.Now().UTC()
}
