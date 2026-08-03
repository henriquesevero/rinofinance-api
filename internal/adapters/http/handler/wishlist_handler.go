package handler

import (
	"errors"
	"net/http"

	appwishlist "rinofinance-api/internal/application/wishlist"

	"rinofinance-api/internal/adapters/http/dto"
	"rinofinance-api/internal/pkg/unfurl"
)

type WishlistHandler struct {
	svc *appwishlist.Service
}

func NewWishlistHandler(svc *appwishlist.Service) *WishlistHandler {
	return &WishlistHandler{svc: svc}
}

func (h *WishlistHandler) Overview(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	overview, err := h.svc.GetOverview(r.Context(), userID, wishlistKind(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewWishlistOverviewResponse(overview))
}

func (h *WishlistHandler) Unfurl(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireUserID(w, r); !ok {
		return
	}
	meta, err := unfurl.Fetch(r.Context(), r.URL.Query().Get("url"))
	if errors.Is(err, unfurl.ErrInvalidURL) {
		writeError(w, errBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, meta)
}

func (h *WishlistHandler) CreateSection(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	var req dto.WishlistSectionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	s, err := h.svc.CreateSection(r.Context(), userID, wishlistKind(r), req.Name)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto.NewWishlistSectionResponse(s))
}

func (h *WishlistHandler) UpdateSection(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	sectionID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}
	var req dto.WishlistSectionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	s, err := h.svc.UpdateSection(r.Context(), userID, sectionID, req.Name)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewWishlistSectionResponse(s))
}

func (h *WishlistHandler) DeleteSection(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	sectionID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteSection(r.Context(), userID, sectionID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WishlistHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	var req dto.WishlistItemRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	in, err := itemInput(req)
	if err != nil {
		writeError(w, err)
		return
	}
	item, err := h.svc.CreateItem(r.Context(), userID, wishlistKind(r), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto.NewWishlistItemResponse(item))
}

func (h *WishlistHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	itemID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}
	var req dto.WishlistItemRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	in, err := itemInput(req)
	if err != nil {
		writeError(w, err)
		return
	}
	item, err := h.svc.UpdateItem(r.Context(), userID, itemID, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewWishlistItemResponse(item))
}

func (h *WishlistHandler) ReorderItems(w http.ResponseWriter, r *http.Request) {
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
	if err := h.svc.ReorderItems(r.Context(), userID, wishlistKind(r), ids); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WishlistHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	itemID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteItem(r.Context(), userID, itemID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func wishlistKind(r *http.Request) string {
	if r.URL.Query().Get("kind") == "owned" {
		return "owned"
	}
	return "wishlist"
}

func itemInput(req dto.WishlistItemRequest) (appwishlist.ItemInput, error) {
	sectionID, err := parseOptionalUUID(req.SectionID)
	if err != nil {
		return appwishlist.ItemInput{}, err
	}
	return appwishlist.ItemInput{
		Name:      req.Name,
		URL:       req.URL,
		Price:     req.Price,
		ImageURL:  req.ImageURL,
		SectionID: sectionID,
	}, nil
}
