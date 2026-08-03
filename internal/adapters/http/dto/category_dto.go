package dto

import (
	"github.com/google/uuid"

	domaincategory "rinofinance-api/internal/domain/category"
)

type CategoryRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}

type CategoryResponse struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color string    `json:"color"`
	Icon  string    `json:"icon,omitempty"`
}

func NewCategoryResponse(c *domaincategory.Category) CategoryResponse {
	return CategoryResponse{ID: c.ID, Name: c.Name, Color: c.Color, Icon: c.Icon}
}

func NewCategoriesResponse(categories []*domaincategory.Category) []CategoryResponse {
	out := make([]CategoryResponse, 0, len(categories))
	for _, c := range categories {
		out = append(out, NewCategoryResponse(c))
	}
	return out
}
