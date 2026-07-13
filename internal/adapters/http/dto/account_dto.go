package dto

// UpdateProfileRequest is the payload for PUT /api/account/profile.
type UpdateProfileRequest struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
}

// ChangeEmailRequest is the payload for PUT /api/account/email.
type ChangeEmailRequest struct {
	NewEmail        string `json:"newEmail"`
	CurrentPassword string `json:"currentPassword"`
}

// ChangePasswordRequest is the payload for PUT /api/account/password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// DeleteAccountRequest is the payload for DELETE /api/account.
type DeleteAccountRequest struct {
	CurrentPassword string `json:"currentPassword"`
}
