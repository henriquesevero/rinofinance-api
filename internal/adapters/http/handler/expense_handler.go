package handler

import (
	"net/http"

	appexpense "rinofinance-api/internal/application/expense"

	"rinofinance-api/internal/adapters/http/dto"
)

// ExpenseHandler exposes CRUD, toggle and card-linkage endpoints for
// Aba 1's "Saídas".
type ExpenseHandler struct {
	create            *appexpense.CreateExpenseUseCase
	createCardLinked  *appexpense.CreateCardLinkedExpenseUseCase
	createAccountLink *appexpense.CreateAccountLinkedExpenseUseCase
	update            *appexpense.UpdateExpenseUseCase
	toggle            *appexpense.ToggleExpenseUseCase
	togglePaid        *appexpense.TogglePaidUseCase
	delete            *appexpense.DeleteExpenseUseCase
	list              *appexpense.ListExpensesUseCase
	reorder           *appexpense.ReorderExpensesUseCase
}

// NewExpenseHandler wires the dependencies for ExpenseHandler.
func NewExpenseHandler(
	create *appexpense.CreateExpenseUseCase,
	createCardLinked *appexpense.CreateCardLinkedExpenseUseCase,
	createAccountLink *appexpense.CreateAccountLinkedExpenseUseCase,
	update *appexpense.UpdateExpenseUseCase,
	toggle *appexpense.ToggleExpenseUseCase,
	togglePaid *appexpense.TogglePaidUseCase,
	delete *appexpense.DeleteExpenseUseCase,
	list *appexpense.ListExpensesUseCase,
	reorder *appexpense.ReorderExpensesUseCase,
) *ExpenseHandler {
	return &ExpenseHandler{
		create:            create,
		createCardLinked:  createCardLinked,
		createAccountLink: createAccountLink,
		update:            update,
		toggle:            toggle,
		togglePaid:        togglePaid,
		delete:            delete,
		list:              list,
		reorder:           reorder,
	}
}

// CreateAccountLinked handles POST /api/expenses/account-linked.
func (h *ExpenseHandler) CreateAccountLinked(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	var req dto.AccountLinkedExpenseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	categoryID, err := parseOptionalUUID(req.CategoryID)
	if err != nil {
		writeError(w, errBadRequest)
		return
	}
	e, err := h.createAccountLink.Execute(r.Context(), userID, req.AccountID, req.Name, categoryID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto.NewExpenseResponse(e))
}

// Reorder handles PUT /api/expenses/order.
func (h *ExpenseHandler) Reorder(w http.ResponseWriter, r *http.Request) {
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

// List handles GET /api/expenses?month=YYYY-MM.
func (h *ExpenseHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	reference, err := parseReferenceMonth(r)
	if err != nil {
		writeError(w, err)
		return
	}

	expenses, err := h.list.Execute(r.Context(), userID, reference)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewExpenseResponseList(expenses))
}

// Create handles POST /api/expenses (standalone expense).
func (h *ExpenseHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	var req dto.ExpenseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	categoryID, err := parseOptionalUUID(req.CategoryID)
	if err != nil {
		writeError(w, errBadRequest)
		return
	}

	e, err := h.create.Execute(r.Context(), userID, req.Name, req.Amount, categoryID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto.NewExpenseResponse(e))
}

// CreateCardLinked handles POST /api/expenses/card-linked.
func (h *ExpenseHandler) CreateCardLinked(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	var req dto.CardLinkedExpenseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	categoryID, err := parseOptionalUUID(req.CategoryID)
	if err != nil {
		writeError(w, errBadRequest)
		return
	}

	e, err := h.createCardLinked.Execute(r.Context(), userID, req.CardID, req.Name, categoryID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto.NewExpenseResponse(e))
}

// Update handles PUT /api/expenses/{id}.
func (h *ExpenseHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	expenseID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}

	var req dto.ExpenseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	categoryID, err := parseOptionalUUID(req.CategoryID)
	if err != nil {
		writeError(w, errBadRequest)
		return
	}

	e, err := h.update.Execute(r.Context(), userID, expenseID, req.Name, req.Amount, categoryID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewExpenseResponse(e))
}

// Toggle handles PATCH /api/expenses/{id}/toggle.
func (h *ExpenseHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	expenseID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}

	e, err := h.toggle.Execute(r.Context(), userID, expenseID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewExpenseResponse(e))
}

// TogglePaid handles PATCH /api/expenses/{id}/paid.
func (h *ExpenseHandler) TogglePaid(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	expenseID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}

	e, err := h.togglePaid.Execute(r.Context(), userID, expenseID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewExpenseResponse(e))
}

// Delete handles DELETE /api/expenses/{id}.
func (h *ExpenseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	expenseID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}

	if err := h.delete.Execute(r.Context(), userID, expenseID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
