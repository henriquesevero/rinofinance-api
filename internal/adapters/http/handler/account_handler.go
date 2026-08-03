package handler

import (
	"net/http"

	appprofile "rinofinance-api/internal/application/profile"

	"rinofinance-api/internal/adapters/http/dto"
)

type AccountHandler struct {
	updateProfile  *appprofile.UpdateProfileUseCase
	changeEmail    *appprofile.ChangeEmailUseCase
	changePassword *appprofile.ChangePasswordUseCase
	deleteAccount  *appprofile.DeleteAccountUseCase
	shareData      *appprofile.ShareDataUseCase
	stopSharing    *appprofile.StopSharingUseCase
}

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
