// Package category models user-defined spending categories (e.g.
// "Alimentação", "Transporte") that can be attached to incomes, expenses,
// installment purchases and subscriptions to power the "gastos por
// categoria" breakdown.
package category

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"rinofinance-api/internal/domain/shared"
)

// Category is a named, colored label belonging to a user.
type Category struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Name   string
	// Color is a user-chosen hex color (e.g. "#8A05BE") used for the
	// category's dot and its slice in the "por categoria" chart.
	Color string
	// Icon optionally holds a single emoji shown next to the name.
	Icon string
	// Position is the manual sort order in the list (lower comes first).
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SetPosition sets the manual sort order of the category in its list.
func (c *Category) SetPosition(position int) {
	c.Position = position
	c.UpdatedAt = time.Now().UTC()
}

// NewCategory builds a new category for a user.
func NewCategory(userID uuid.UUID, name, color, icon string) (*Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, shared.ErrEmptyName
	}

	now := time.Now().UTC()
	c := &Category{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	c.SetColor(color)
	c.SetIcon(icon)
	c.UpdatedAt = now
	return c, nil
}

// Rename updates the category's display name.
func (c *Category) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return shared.ErrEmptyName
	}
	c.Name = name
	c.UpdatedAt = time.Now().UTC()
	return nil
}

// SetColor replaces the category's color. Empty falls back to a neutral
// default so the UI always has something to render.
func (c *Category) SetColor(color string) {
	color = strings.TrimSpace(color)
	if color == "" {
		color = "#6B7280"
	}
	c.Color = color
	c.UpdatedAt = time.Now().UTC()
}

// SetIcon replaces the category's emoji icon (may be empty).
func (c *Category) SetIcon(icon string) {
	c.Icon = strings.TrimSpace(icon)
	c.UpdatedAt = time.Now().UTC()
}
