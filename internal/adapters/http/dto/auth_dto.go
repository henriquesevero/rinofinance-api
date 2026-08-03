package dto

import (
	"github.com/google/uuid"

	domainuser "rinofinance-api/internal/domain/user"
)

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatarUrl,omitempty"`

	Shared bool `json:"shared"`
}

type ShareDataRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

func NewUserResponse(u *domainuser.User) UserResponse {
	return UserResponse{ID: u.ID, Name: u.Name, Email: u.Email, AvatarURL: u.AvatarURL, Shared: u.DataOwnerID != nil}
}
