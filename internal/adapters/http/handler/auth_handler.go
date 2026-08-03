package handler

import (
	"net/http"

	appauth "rinofinance-api/internal/application/auth"

	"rinofinance-api/internal/adapters/http/dto"
)

type AuthHandler struct {
	register *appauth.RegisterUserUseCase
	login    *appauth.LoginUserUseCase
}

func NewAuthHandler(register *appauth.RegisterUserUseCase, login *appauth.LoginUserUseCase) *AuthHandler {
	return &AuthHandler{register: register, login: login}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	u, err := h.register.Execute(r.Context(), req.Name, req.Email, req.Password, req.Code)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, dto.NewUserResponse(u))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	token, u, err := h.login.Execute(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.AuthResponse{Token: token, User: dto.NewUserResponse(u)})
}
