package category

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"rinofinance-api/internal/domain/shared"
)

const fallbackColor = "#6B7280"

type Category struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Name   string

	Color string

	Icon string

	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (c *Category) SetPosition(position int) {
	c.Position = position
	c.UpdatedAt = time.Now().UTC()
}

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

func (c *Category) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return shared.ErrEmptyName
	}
	c.Name = name
	c.UpdatedAt = time.Now().UTC()
	return nil
}

func (c *Category) SetColor(color string) {
	color = strings.TrimSpace(color)
	if color == "" {
		color = fallbackColor
	}
	c.Color = color
	c.UpdatedAt = time.Now().UTC()
}

func (c *Category) SetIcon(icon string) {
	c.Icon = strings.TrimSpace(icon)
	c.UpdatedAt = time.Now().UTC()
}
