package dto

import (
	"github.com/google/uuid"

	domaincategory "rinofinance-api/internal/domain/category"
)

// CategoryRequest is the payload for creating/updating a category.
type CategoryRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}

// CategoryResponse is the public representation of a Category.
type CategoryResponse struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color string    `json:"color"`
	Icon  string    `json:"icon,omitempty"`
}

// NewCategoryResponse builds a CategoryResponse from the domain Category.
func NewCategoryResponse(c *domaincategory.Category) CategoryResponse {
	return CategoryResponse{ID: c.ID, Name: c.Name, Color: c.Color, Icon: c.Icon}
}

// NewCategoriesResponse builds the list payload for GET /api/categories.
func NewCategoriesResponse(categories []*domaincategory.Category) []CategoryResponse {
	out := make([]CategoryResponse, 0, len(categories))
	for _, c := range categories {
		out = append(out, NewCategoryResponse(c))
	}
	return out
}
