package handler

import (
	"net/http"

	appcategory "rinofinance-api/internal/application/category"

	"rinofinance-api/internal/adapters/http/dto"
)

// CategoryHandler exposes CRUD endpoints for user-defined spending
// categories.
type CategoryHandler struct {
	create  *appcategory.CreateCategoryUseCase
	update  *appcategory.UpdateCategoryUseCase
	delete  *appcategory.DeleteCategoryUseCase
	list    *appcategory.ListCategoriesUseCase
	reorder *appcategory.ReorderCategoriesUseCase
}

// NewCategoryHandler wires the dependencies for CategoryHandler.
func NewCategoryHandler(
	create *appcategory.CreateCategoryUseCase,
	update *appcategory.UpdateCategoryUseCase,
	delete *appcategory.DeleteCategoryUseCase,
	list *appcategory.ListCategoriesUseCase,
	reorder *appcategory.ReorderCategoriesUseCase,
) *CategoryHandler {
	return &CategoryHandler{create: create, update: update, delete: delete, list: list, reorder: reorder}
}

// Reorder handles PUT /api/categories/order.
func (h *CategoryHandler) Reorder(w http.ResponseWriter, r *http.Request) {
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

// List handles GET /api/categories.
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	categories, err := h.list.Execute(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewCategoriesResponse(categories))
}

// Create handles POST /api/categories.
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	var req dto.CategoryRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	c, err := h.create.Execute(r.Context(), userID, req.Name, req.Color, req.Icon)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto.NewCategoryResponse(c))
}

// Update handles PUT /api/categories/{id}.
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	categoryID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}
	var req dto.CategoryRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	c, err := h.update.Execute(r.Context(), userID, categoryID, req.Name, req.Color, req.Icon)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewCategoryResponse(c))
}

// Delete handles DELETE /api/categories/{id}.
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	categoryID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}
	if err := h.delete.Execute(r.Context(), userID, categoryID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
