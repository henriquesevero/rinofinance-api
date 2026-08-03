package handler

import (
	"net/http"

	appincome "rinofinance-api/internal/application/income"

	"rinofinance-api/internal/adapters/http/dto"
)

type IncomeHandler struct {
	create            *appincome.CreateIncomeUseCase
	createAccountLink *appincome.CreateAccountLinkedIncomeUseCase
	update            *appincome.UpdateIncomeUseCase
	toggle            *appincome.ToggleIncomeUseCase
	toggleReceived    *appincome.ToggleReceivedUseCase
	delete            *appincome.DeleteIncomeUseCase
	list              *appincome.ListIncomesUseCase
	reorder           *appincome.ReorderIncomesUseCase
}

func NewIncomeHandler(
	create *appincome.CreateIncomeUseCase,
	createAccountLink *appincome.CreateAccountLinkedIncomeUseCase,
	update *appincome.UpdateIncomeUseCase,
	toggle *appincome.ToggleIncomeUseCase,
	toggleReceived *appincome.ToggleReceivedUseCase,
	delete *appincome.DeleteIncomeUseCase,
	list *appincome.ListIncomesUseCase,
	reorder *appincome.ReorderIncomesUseCase,
) *IncomeHandler {
	return &IncomeHandler{
		create:            create,
		createAccountLink: createAccountLink,
		update:            update,
		toggle:            toggle,
		toggleReceived:    toggleReceived,
		delete:            delete,
		list:              list,
		reorder:           reorder,
	}
}

func (h *IncomeHandler) CreateAccountLinked(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	var req dto.AccountLinkedIncomeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	categoryID, err := parseOptionalUUID(req.CategoryID)
	if err != nil {
		writeError(w, errBadRequest)
		return
	}
	inc, err := h.createAccountLink.Execute(r.Context(), userID, req.AccountID, req.Name, categoryID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto.NewIncomeResponse(inc))
}

func (h *IncomeHandler) Reorder(w http.ResponseWriter, r *http.Request) {
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

func (h *IncomeHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	incomes, err := h.list.Execute(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewIncomeResponseList(incomes))
}

func (h *IncomeHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	var req dto.IncomeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	categoryID, err := parseOptionalUUID(req.CategoryID)
	if err != nil {
		writeError(w, errBadRequest)
		return
	}

	inc, err := h.create.Execute(r.Context(), userID, req.Name, req.Amount, categoryID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto.NewIncomeResponse(inc))
}

func (h *IncomeHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	incomeID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}

	var req dto.IncomeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	categoryID, err := parseOptionalUUID(req.CategoryID)
	if err != nil {
		writeError(w, errBadRequest)
		return
	}

	inc, err := h.update.Execute(r.Context(), userID, incomeID, req.Name, req.Amount, categoryID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewIncomeResponse(inc))
}

func (h *IncomeHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	incomeID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}

	inc, err := h.toggle.Execute(r.Context(), userID, incomeID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewIncomeResponse(inc))
}

func (h *IncomeHandler) ToggleReceived(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	incomeID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}
	reference, err := parseReferenceMonth(r)
	if err != nil {
		writeError(w, err)
		return
	}

	inc, err := h.toggleReceived.Execute(r.Context(), userID, incomeID, reference.Format("2006-01"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewIncomeResponse(inc))
}

func (h *IncomeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	incomeID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}

	if err := h.delete.Execute(r.Context(), userID, incomeID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
