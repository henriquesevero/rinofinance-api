// Package wishlist orchestrates the "itens para comprar" use cases: managing
// sections and the items (products) inside them, plus the overview total.
package wishlist

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainwishlist "rinofinance-api/internal/domain/wishlist"
	"rinofinance-api/internal/domain/shared"
)

// Overview is the full wishlist payload: sections, items and their combined
// value.
type Overview struct {
	Sections []*domainwishlist.Section
	Items    []*domainwishlist.Item
	Total    shared.Money
}

// Service groups the wishlist use cases behind a single struct, since they
// share the same two repositories.
type Service struct {
	sections domainwishlist.SectionRepository
	items    domainwishlist.ItemRepository
}

// NewService wires the wishlist service to its repositories.
func NewService(sections domainwishlist.SectionRepository, items domainwishlist.ItemRepository) *Service {
	return &Service{sections: sections, items: items}
}

// GetOverview returns the user's sections, items and total value.
func (s *Service) GetOverview(ctx context.Context, userID uuid.UUID) (Overview, error) {
	sections, err := s.sections.ListByUser(ctx, userID)
	if err != nil {
		return Overview{}, fmt.Errorf("erro ao listar seções: %w", err)
	}
	items, err := s.items.ListByUser(ctx, userID)
	if err != nil {
		return Overview{}, fmt.Errorf("erro ao listar itens: %w", err)
	}
	total := shared.Zero
	for _, i := range items {
		total = total.Add(i.Price)
	}
	return Overview{Sections: sections, Items: items, Total: total}, nil
}

// CreateSection adds a section, appended to the end of the user's ordering.
func (s *Service) CreateSection(ctx context.Context, userID uuid.UUID, name string) (*domainwishlist.Section, error) {
	sec, err := domainwishlist.NewSection(userID, name)
	if err != nil {
		return nil, err
	}
	existing, err := s.sections.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao contar seções: %w", err)
	}
	sec.SetPosition(len(existing))
	if err := s.sections.Create(ctx, sec); err != nil {
		return nil, fmt.Errorf("erro ao criar seção: %w", err)
	}
	return sec, nil
}

// UpdateSection renames a section.
func (s *Service) UpdateSection(ctx context.Context, userID, sectionID uuid.UUID, name string) (*domainwishlist.Section, error) {
	sec, err := s.sections.FindByID(ctx, sectionID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar seção: %w", err)
	}
	if sec.UserID != userID {
		return nil, shared.ErrNotFound
	}
	if err := sec.Rename(name); err != nil {
		return nil, err
	}
	if err := s.sections.Update(ctx, sec); err != nil {
		return nil, fmt.Errorf("erro ao atualizar seção: %w", err)
	}
	return sec, nil
}

// DeleteSection removes a section and moves its items to "Sem seção" so the
// products the user wanted are never lost.
func (s *Service) DeleteSection(ctx context.Context, userID, sectionID uuid.UUID) error {
	sec, err := s.sections.FindByID(ctx, sectionID)
	if err != nil {
		return fmt.Errorf("erro ao buscar seção: %w", err)
	}
	if sec.UserID != userID {
		return shared.ErrNotFound
	}

	items, err := s.items.ListByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("erro ao listar itens: %w", err)
	}
	for _, i := range items {
		if i.SectionID != nil && *i.SectionID == sectionID {
			i.SetSection(nil)
			if err := s.items.Update(ctx, i); err != nil {
				return fmt.Errorf("erro ao desvincular item: %w", err)
			}
		}
	}

	if err := s.sections.Delete(ctx, sectionID); err != nil {
		return fmt.Errorf("erro ao remover seção: %w", err)
	}
	return nil
}

// ItemInput carries the editable fields of a wishlist item.
type ItemInput struct {
	Name      string
	URL       string
	Price     shared.Money
	ImageURL  string
	SectionID *uuid.UUID
}

// CreateItem adds a product to the wishlist, validating the section belongs
// to the user when provided.
func (s *Service) CreateItem(ctx context.Context, userID uuid.UUID, in ItemInput) (*domainwishlist.Item, error) {
	if err := s.verifySection(ctx, userID, in.SectionID); err != nil {
		return nil, err
	}
	item, err := domainwishlist.NewItem(userID, in.Name, in.URL, in.Price)
	if err != nil {
		return nil, err
	}
	item.SetImage(in.ImageURL)
	item.SetSection(in.SectionID)

	existing, err := s.items.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao contar itens: %w", err)
	}
	item.SetPosition(len(existing))
	if err := s.items.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("erro ao criar item: %w", err)
	}
	return item, nil
}

// UpdateItem replaces an item's editable fields.
func (s *Service) UpdateItem(ctx context.Context, userID, itemID uuid.UUID, in ItemInput) (*domainwishlist.Item, error) {
	item, err := s.items.FindByID(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar item: %w", err)
	}
	if item.UserID != userID {
		return nil, shared.ErrNotFound
	}
	if err := s.verifySection(ctx, userID, in.SectionID); err != nil {
		return nil, err
	}
	if err := item.Update(in.Name, in.URL, in.Price, in.ImageURL, in.SectionID); err != nil {
		return nil, err
	}
	if err := s.items.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("erro ao atualizar item: %w", err)
	}
	return item, nil
}

// DeleteItem removes a wishlist item.
func (s *Service) DeleteItem(ctx context.Context, userID, itemID uuid.UUID) error {
	item, err := s.items.FindByID(ctx, itemID)
	if err != nil {
		return fmt.Errorf("erro ao buscar item: %w", err)
	}
	if item.UserID != userID {
		return shared.ErrNotFound
	}
	if err := s.items.Delete(ctx, itemID); err != nil {
		return fmt.Errorf("erro ao remover item: %w", err)
	}
	return nil
}

// verifySection ensures a referenced section exists and belongs to the user.
func (s *Service) verifySection(ctx context.Context, userID uuid.UUID, sectionID *uuid.UUID) error {
	if sectionID == nil {
		return nil
	}
	sec, err := s.sections.FindByID(ctx, *sectionID)
	if err != nil {
		return fmt.Errorf("erro ao buscar seção: %w", err)
	}
	if sec.UserID != userID {
		return shared.ErrNotFound
	}
	return nil
}
