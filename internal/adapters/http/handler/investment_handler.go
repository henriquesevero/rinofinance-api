package handler

import (
	"net/http"

	appinvestment "rinofinance-api/internal/application/investment"

	"rinofinance-api/internal/adapters/http/dto"
)

type InvestmentHandler struct {
	create         *appinvestment.CreateAssetUseCase
	update         *appinvestment.UpdateAssetUseCase
	toggle         *appinvestment.ToggleAssetUseCase
	delete         *appinvestment.DeleteAssetUseCase
	list           *appinvestment.ListAssetsUseCase
	createProvento *appinvestment.CreateProventoUseCase
	deleteProvento *appinvestment.DeleteProventoUseCase
}

func NewInvestmentHandler(
	create *appinvestment.CreateAssetUseCase,
	update *appinvestment.UpdateAssetUseCase,
	toggle *appinvestment.ToggleAssetUseCase,
	delete *appinvestment.DeleteAssetUseCase,
	list *appinvestment.ListAssetsUseCase,
	createProvento *appinvestment.CreateProventoUseCase,
	deleteProvento *appinvestment.DeleteProventoUseCase,
) *InvestmentHandler {
	return &InvestmentHandler{
		create:         create,
		update:         update,
		toggle:         toggle,
		delete:         delete,
		list:           list,
		createProvento: createProvento,
		deleteProvento: deleteProvento,
	}
}

func (h *InvestmentHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	overview, err := h.list.Execute(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewAssetsOverviewResponse(overview))
}

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

	a, err := h.create.Execute(r.Context(), userID, req.ToInput())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto.NewAssetResponse(a))
}

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

	a, err := h.update.Execute(r.Context(), userID, assetID, req.ToInput())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewAssetResponse(a))
}

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

func (h *InvestmentHandler) CreateProvento(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	var req dto.ProventoRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	p, err := h.createProvento.Execute(r.Context(), userID, req.AssetID, req.Amount, req.Date)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto.NewProventoResponse(p))
}

func (h *InvestmentHandler) DeleteProvento(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	proventoID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}

	if err := h.deleteProvento.Execute(r.Context(), userID, proventoID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
