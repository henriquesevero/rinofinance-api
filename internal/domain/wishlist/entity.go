// Package wishlist models a user's "itens para comprar" — a personal
// storefront of products they want to buy, each with a link, price and
// optional image, organized into user-defined sections (Eletrônicos,
// Roupas, ...).
package wishlist

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"rinofinance-api/internal/domain/shared"
)

// Section is a user-defined grouping of wishlist items (e.g. "Eletrônicos").
type Section struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Name      string
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewSection builds a new section for a user.
func NewSection(userID uuid.UUID, name string) (*Section, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, shared.ErrEmptyName
	}
	now := time.Now().UTC()
	return &Section{ID: uuid.New(), UserID: userID, Name: name, CreatedAt: now, UpdatedAt: now}, nil
}

// Rename updates the section's display name.
func (s *Section) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return shared.ErrEmptyName
	}
	s.Name = name
	s.UpdatedAt = time.Now().UTC()
	return nil
}

// SetPosition sets the manual sort order of the section in its list.
func (s *Section) SetPosition(position int) {
	s.Position = position
	s.UpdatedAt = time.Now().UTC()
}

// Item is a single product the user wants to buy.
type Item struct {
	ID     uuid.UUID
	UserID uuid.UUID
	// SectionID optionally links the item to a section. Nil means the item
	// is ungrouped ("Sem seção").
	SectionID *uuid.UUID
	Name      string
	// URL is the store link opened when the item is clicked.
	URL string
	// Price is the item's value, summed into the wishlist total.
	Price shared.Money
	// ImageURL optionally holds a product image (a data: URL or a remote
	// URL) shown as the item's thumbnail.
	ImageURL  string
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewItem builds a new wishlist item for a user.
func NewItem(userID uuid.UUID, name, url string, price shared.Money) (*Item, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, shared.ErrEmptyName
	}
	if price.IsNegative() {
		return nil, shared.ErrNegativeAmount
	}
	now := time.Now().UTC()
	return &Item{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      name,
		URL:       strings.TrimSpace(url),
		Price:     price,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Update replaces the item's editable fields.
func (i *Item) Update(name, url string, price shared.Money, imageURL string, sectionID *uuid.UUID) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return shared.ErrEmptyName
	}
	if price.IsNegative() {
		return shared.ErrNegativeAmount
	}
	i.Name = name
	i.URL = strings.TrimSpace(url)
	i.Price = price
	i.ImageURL = imageURL
	i.SectionID = sectionID
	i.UpdatedAt = time.Now().UTC()
	return nil
}

// SetImage replaces the item's image (data: or remote URL). Empty clears it.
func (i *Item) SetImage(imageURL string) {
	i.ImageURL = imageURL
	i.UpdatedAt = time.Now().UTC()
}

// SetSection links (or clears, when nil) the item's section.
func (i *Item) SetSection(sectionID *uuid.UUID) {
	i.SectionID = sectionID
	i.UpdatedAt = time.Now().UTC()
}

// SetPosition sets the manual sort order within the item list.
func (i *Item) SetPosition(position int) {
	i.Position = position
	i.UpdatedAt = time.Now().UTC()
}
