package handler

import (
	"net/http"

	appinvestment "rinofinance-api/internal/application/investment"

	"rinofinance-api/internal/adapters/http/dto"
)

// InvestmentHandler exposes CRUD endpoints for Aba 3's investment/
// patrimony assets.
type InvestmentHandler struct {
	create *appinvestment.CreateAssetUseCase
	update *appinvestment.UpdateAssetUseCase
	toggle *appinvestment.ToggleAssetUseCase
	delete *appinvestment.DeleteAssetUseCase
	list   *appinvestment.ListAssetsUseCase
}

// NewInvestmentHandler wires the dependencies for InvestmentHandler.
func NewInvestmentHandler(
	create *appinvestment.CreateAssetUseCase,
	update *appinvestment.UpdateAssetUseCase,
	toggle *appinvestment.ToggleAssetUseCase,
	delete *appinvestment.DeleteAssetUseCase,
	list *appinvestment.ListAssetsUseCase,
) *InvestmentHandler {
	return &InvestmentHandler{create: create, update: update, toggle: toggle, delete: delete, list: list}
}

// List handles GET /api/investments.
func (h *InvestmentHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	assets, total, err := h.list.Execute(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewAssetsOverviewResponse(assets, total))
}

// Create handles POST /api/investments.
func (h *InvestmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	var req dto.AssetRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	a, err := h.create.Execute(r.Context(), userID, req.Name, req.CurrentBalance)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto.NewAssetResponse(a))
}

// Update handles PUT /api/investments/{id}.
func (h *InvestmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	assetID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}

	var req dto.AssetRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	a, err := h.update.Execute(r.Context(), userID, assetID, req.Name, req.CurrentBalance)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewAssetResponse(a))
}

// Toggle handles PATCH /api/investments/{id}/toggle.
func (h *InvestmentHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	assetID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}

	a, err := h.toggle.Execute(r.Context(), userID, assetID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewAssetResponse(a))
}

// Delete handles DELETE /api/investments/{id}.
func (h *InvestmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	assetID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}

	if err := h.delete.Execute(r.Context(), userID, assetID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
