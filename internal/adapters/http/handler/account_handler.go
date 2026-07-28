package handler

import (
	"net/http"

	appprofile "rinofinance-api/internal/application/profile"

	"rinofinance-api/internal/adapters/http/dto"
)

// AccountHandler exposes account-settings endpoints: profile (name +
// avatar), email, password and account deletion.
type AccountHandler struct {
	updateProfile  *appprofile.UpdateProfileUseCase
	changeEmail    *appprofile.ChangeEmailUseCase
	changePassword *appprofile.ChangePasswordUseCase
	deleteAccount  *appprofile.DeleteAccountUseCase
	shareData      *appprofile.ShareDataUseCase
	stopSharing    *appprofile.StopSharingUseCase
}

// NewAccountHandler wires the dependencies for AccountHandler.
func NewAccountHandler(
	updateProfile *appprofile.UpdateProfileUseCase,
	changeEmail *appprofile.ChangeEmailUseCase,
	changePassword *appprofile.ChangePasswordUseCase,
	deleteAccount *appprofile.DeleteAccountUseCase,
	shareData *appprofile.ShareDataUseCase,
	stopSharing *appprofile.StopSharingUseCase,
) *AccountHandler {
	return &AccountHandler{
		updateProfile:  updateProfile,
		changeEmail:    changeEmail,
		changePassword: changePassword,
		deleteAccount:  deleteAccount,
		shareData:      shareData,
		stopSharing:    stopSharing,
	}
}

// ShareData handles POST /api/account/share — link this account to another
// (household sharing) using its email + password.
func (h *AccountHandler) ShareData(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}
	var req dto.ShareDataRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	u, err := h.shareData.Execute(r.Context(), userID, req.Email, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewUserResponse(u))
}

// StopSharing handles POST /api/account/unshare — revert to own data.
func (h *AccountHandler) StopSharing(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}
	if err := h.stopSharing.Execute(r.Context(), userID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UpdateProfile handles PUT /api/account/profile.
func (h *AccountHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	var req dto.UpdateProfileRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	u, err := h.updateProfile.Execute(r.Context(), userID, req.Name, req.AvatarURL)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewUserResponse(u))
}

// ChangeEmail handles PUT /api/account/email.
func (h *AccountHandler) ChangeEmail(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	var req dto.ChangeEmailRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	u, err := h.changeEmail.Execute(r.Context(), userID, req.NewEmail, req.CurrentPassword)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewUserResponse(u))
}

// ChangePassword handles PUT /api/account/password.
func (h *AccountHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	var req dto.ChangePasswordRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	if err := h.changePassword.Execute(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteAccount handles DELETE /api/account.
func (h *AccountHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	var req dto.DeleteAccountRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	if err := h.deleteAccount.Execute(r.Context(), userID, req.CurrentPassword); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
