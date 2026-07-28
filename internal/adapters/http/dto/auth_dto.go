// Package dto holds the HTTP request/response shapes for every resource.
// Amount fields use shared.Money directly (it already marshals/unmarshals
// as a plain JSON number backed by decimal.Decimal), so the frontend never
// has to think about precision loss.
package dto

import (
	"github.com/google/uuid"

	domainuser "rinofinance-api/internal/domain/user"
)

// RegisterRequest is the payload for POST /api/auth/register.
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

// LoginRequest is the payload for POST /api/auth/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserResponse is the public-facing representation of a User (never
// includes PasswordHash).
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatarUrl,omitempty"`
	// Shared is true when this account is viewing another's data (household).
	Shared bool `json:"shared"`
}

// ShareDataRequest is the payload for POST /api/account/share.
type ShareDataRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse is returned by both register and login.
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// NewUserResponse builds a UserResponse from the domain User.
func NewUserResponse(u *domainuser.User) UserResponse {
	return UserResponse{ID: u.ID, Name: u.Name, Email: u.Email, AvatarURL: u.AvatarURL, Shared: u.DataOwnerID != nil}
}
