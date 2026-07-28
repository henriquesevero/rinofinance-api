package dto

import (
	"github.com/google/uuid"

	appwishlist "rinofinance-api/internal/application/wishlist"
	domainwishlist "rinofinance-api/internal/domain/wishlist"
	"rinofinance-api/internal/domain/shared"
)

// WishlistSectionRequest is the payload for creating/updating a section.
type WishlistSectionRequest struct {
	Name string `json:"name"`
}

// WishlistItemRequest is the payload for creating/updating a wishlist item.
type WishlistItemRequest struct {
	Name      string       `json:"name"`
	URL       string       `json:"url"`
	Price     shared.Money `json:"price"`
	ImageURL  string       `json:"imageUrl"`
	SectionID string       `json:"sectionId"`
}

// WishlistSectionResponse is the public representation of a section.
type WishlistSectionResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// NewWishlistSectionResponse builds a section response from the domain.
func NewWishlistSectionResponse(s *domainwishlist.Section) WishlistSectionResponse {
	return WishlistSectionResponse{ID: s.ID, Name: s.Name}
}

// WishlistItemResponse is the public representation of a wishlist item.
type WishlistItemResponse struct {
	ID        uuid.UUID    `json:"id"`
	SectionID *uuid.UUID   `json:"sectionId,omitempty"`
	Name      string       `json:"name"`
	URL       string       `json:"url,omitempty"`
	Price     shared.Money `json:"price"`
	ImageURL  string       `json:"imageUrl,omitempty"`
}

// NewWishlistItemResponse builds an item response from the domain.
func NewWishlistItemResponse(i *domainwishlist.Item) WishlistItemResponse {
	return WishlistItemResponse{
		ID:        i.ID,
		SectionID: i.SectionID,
		Name:      i.Name,
		URL:       i.URL,
		Price:     i.Price,
		ImageURL:  i.ImageURL,
	}
}

// WishlistOverviewResponse is the full payload for GET /api/wishlist.
type WishlistOverviewResponse struct {
	Sections []WishlistSectionResponse `json:"sections"`
	Items    []WishlistItemResponse    `json:"items"`
	Total    shared.Money              `json:"total"`
}

// NewWishlistOverviewResponse builds the overview payload.
func NewWishlistOverviewResponse(o appwishlist.Overview) WishlistOverviewResponse {
	sections := make([]WishlistSectionResponse, 0, len(o.Sections))
	for _, s := range o.Sections {
		sections = append(sections, NewWishlistSectionResponse(s))
	}
	items := make([]WishlistItemResponse, 0, len(o.Items))
	for _, i := range o.Items {
		items = append(items, NewWishlistItemResponse(i))
	}
	return WishlistOverviewResponse{Sections: sections, Items: items, Total: o.Total}
}
