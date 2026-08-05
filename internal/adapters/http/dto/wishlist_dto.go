package dto

import (
	"github.com/google/uuid"

	appwishlist "rinofinance-api/internal/application/wishlist"
	"rinofinance-api/internal/domain/shared"
	domainwishlist "rinofinance-api/internal/domain/wishlist"
)

type WishlistSectionRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type WishlistItemRequest struct {
	Name      string       `json:"name"`
	URL       string       `json:"url"`
	Price     shared.Money `json:"price"`
	ImageURL  string       `json:"imageUrl"`
	LogoURL   string       `json:"logoUrl"`
	SectionID string       `json:"sectionId"`
}

type WishlistSectionResponse struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color string    `json:"color"`
}

func NewWishlistSectionResponse(s *domainwishlist.Section) WishlistSectionResponse {
	return WishlistSectionResponse{ID: s.ID, Name: s.Name, Color: s.Color}
}

type WishlistItemResponse struct {
	ID        uuid.UUID    `json:"id"`
	SectionID *uuid.UUID   `json:"sectionId,omitempty"`
	Name      string       `json:"name"`
	URL       string       `json:"url,omitempty"`
	Price     shared.Money `json:"price"`
	ImageURL  string       `json:"imageUrl,omitempty"`
	LogoURL   string       `json:"logoUrl,omitempty"`
}

func NewWishlistItemResponse(i *domainwishlist.Item) WishlistItemResponse {
	return WishlistItemResponse{
		ID:        i.ID,
		SectionID: i.SectionID,
		Name:      i.Name,
		URL:       i.URL,
		Price:     i.Price,
		ImageURL:  i.ImageURL,
		LogoURL:   i.LogoURL,
	}
}

type WishlistOverviewResponse struct {
	Sections []WishlistSectionResponse `json:"sections"`
	Items    []WishlistItemResponse    `json:"items"`
	Total    shared.Money              `json:"total"`
}

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
