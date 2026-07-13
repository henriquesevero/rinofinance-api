package handler

import (
	"net/http"

	appaccount "rinofinance-api/internal/application/account"

	"rinofinance-api/internal/adapters/http/dto"
)

// WalletHandler exposes endpoints for the user's bank/wallet accounts and
// their debit purchases (named distinctly to avoid clashing with the
// user-profile AccountHandler).
type WalletHandler struct {
	create         *appaccount.CreateAccountUseCase
	update         *appaccount.UpdateAccountUseCase
	delete         *appaccount.DeleteAccountUseCase
	list           *appaccount.ListAccountsUseCase
	reorder        *appaccount.ReorderAccountsUseCase
	createPurchase *appaccount.CreateAccountPurchaseUseCase
	updatePurchase *appaccount.UpdateAccountPurchaseUseCase
	deletePurchase *appaccount.DeleteAccountPurchaseUseCase
}

// NewWalletHandler wires the dependencies for WalletHandler.
func NewWalletHandler(
	create *appaccount.CreateAccountUseCase,
	update *appaccount.UpdateAccountUseCase,
	delete *appaccount.DeleteAccountUseCase,
	list *appaccount.ListAccountsUseCase,
	reorder *appaccount.ReorderAccountsUseCase,
	createPurchase *appaccount.CreateAccountPurchaseUseCase,
	updatePurchase *appaccount.UpdateAccountPurchaseUseCase,
	deletePurchase *appaccount.DeleteAccountPurchaseUseCase,
) *WalletHandler {
	return &WalletHandler{
		create:         create,
		update:         update,
		delete:         delete,
		list:           list,
		reorder:        reorder,
		createPurchase: createPurchase,
		updatePurchase: updatePurchase,
		deletePurchase: deletePurchase,
	}
}

// Reorder handles PUT /api/accounts/order.
func (h *WalletHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	var req dto.ReorderRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	ids, err := parseUUIDList(req.IDs)
	if err != nil {
		writeError(w, errBadRequest)
		return
	}
	if err := h.reorder.Execute(r.Context(), userID, ids); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// List handles GET /api/accounts?month=YYYY-MM.
func (h *WalletHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	reference, err := parseReferenceMonth(r)
	if err != nil {
		writeError(w, err)
		return
	}
	overviews, total, err := h.list.Execute(r.Context(), userID, reference)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewAccountsOverviewResponse(overviews, total))
}

// Create handles POST /api/accounts.
func (h *WalletHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	var req dto.AccountRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	a, err := h.create.Execute(r.Context(), userID, req.Name, req.Color, req.ImageURL, req.Balance)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto.NewCreatedAccountResponse(a))
}

// Update handles PUT /api/accounts/{id}.
func (h *WalletHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	accountID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}
	var req dto.AccountRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	a, err := h.update.Execute(r.Context(), userID, accountID, req.Name, req.Color, req.ImageURL, req.Balance)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewCreatedAccountResponse(a))
}

// Delete handles DELETE /api/accounts/{id}.
func (h *WalletHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	accountID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}
	if err := h.delete.Execute(r.Context(), userID, accountID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// CreatePurchase handles POST /api/accounts/{accountId}/purchases.
func (h *WalletHandler) CreatePurchase(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	accountID, ok := parseUUIDPathValue(w, r, "accountId")
	if !ok {
		return
	}
	var req dto.AccountPurchaseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	date, err := req.ParseDate()
	if err != nil {
		writeError(w, errBadRequest)
		return
	}
	categoryID, err := parseOptionalUUID(req.CategoryID)
	if err != nil {
		writeError(w, errBadRequest)
		return
	}
	p, err := h.createPurchase.Execute(r.Context(), userID, accountID, req.Name, req.Amount, date, categoryID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto.NewAccountPurchaseResponse(p))
}

// UpdatePurchase handles PUT /api/account-purchases/{id}.
func (h *WalletHandler) UpdatePurchase(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	purchaseID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}
	var req dto.AccountPurchaseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	date, err := req.ParseDate()
	if err != nil {
		writeError(w, errBadRequest)
		return
	}
	categoryID, err := parseOptionalUUID(req.CategoryID)
	if err != nil {
		writeError(w, errBadRequest)
		return
	}
	p, err := h.updatePurchase.Execute(r.Context(), userID, purchaseID, req.Name, req.Amount, date, categoryID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewAccountPurchaseResponse(p))
}

// DeletePurchase handles DELETE /api/account-purchases/{id}.
func (h *WalletHandler) DeletePurchase(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	purchaseID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}
	if err := h.deletePurchase.Execute(r.Context(), userID, purchaseID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
