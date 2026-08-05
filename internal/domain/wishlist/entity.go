package wishlist

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"rinofinance-api/internal/domain/shared"
)

const fallbackSectionColor = "#6B7280"

type Section struct {
	ID     uuid.UUID
	UserID uuid.UUID

	Kind      string
	Name      string
	Color     string
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewSection(userID uuid.UUID, kind, name string) (*Section, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, shared.ErrEmptyName
	}
	now := time.Now().UTC()
	return &Section{
		ID:        uuid.New(),
		UserID:    userID,
		Kind:      kind,
		Name:      name,
		Color:     fallbackSectionColor,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (s *Section) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return shared.ErrEmptyName
	}
	s.Name = name
	s.UpdatedAt = time.Now().UTC()
	return nil
}

func (s *Section) SetColor(color string) {
	color = strings.TrimSpace(color)
	if color == "" {
		color = fallbackSectionColor
	}
	s.Color = color
	s.UpdatedAt = time.Now().UTC()
}

func (s *Section) SetPosition(position int) {
	s.Position = position
	s.UpdatedAt = time.Now().UTC()
}

type Item struct {
	ID     uuid.UUID
	UserID uuid.UUID

	Kind string

	SectionID *uuid.UUID
	Name      string

	URL string

	Price shared.Money

	ImageURL  string
	LogoURL   string
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewItem(userID uuid.UUID, kind, name, url string, price shared.Money) (*Item, error) {
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
		Kind:      kind,
		Name:      name,
		URL:       strings.TrimSpace(url),
		Price:     price,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (i *Item) Update(name, url string, price shared.Money, imageURL, logoURL string, sectionID *uuid.UUID) error {
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
	i.LogoURL = strings.TrimSpace(logoURL)
	i.SectionID = sectionID
	i.UpdatedAt = time.Now().UTC()
	return nil
}

func (i *Item) SetImage(imageURL string) {
	i.ImageURL = imageURL
	i.UpdatedAt = time.Now().UTC()
}

func (i *Item) SetLogo(logoURL string) {
	i.LogoURL = strings.TrimSpace(logoURL)
	i.UpdatedAt = time.Now().UTC()
}

func (i *Item) SetSection(sectionID *uuid.UUID) {
	i.SectionID = sectionID
	i.UpdatedAt = time.Now().UTC()
}

func (i *Item) SetPosition(position int) {
	i.Position = position
	i.UpdatedAt = time.Now().UTC()
}
