// Package category orchestrates CRUD use cases for user-defined spending
// categories.
package category

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincategory "rinofinance-api/internal/domain/category"
	"rinofinance-api/internal/domain/shared"
)

// CreateCategoryUseCase creates a new category for a user.
type CreateCategoryUseCase struct {
	repo domaincategory.Repository
}

func NewCreateCategoryUseCase(repo domaincategory.Repository) *CreateCategoryUseCase {
	return &CreateCategoryUseCase{repo: repo}
}

func (uc *CreateCategoryUseCase) Execute(ctx context.Context, userID uuid.UUID, name, color, icon string) (*domaincategory.Category, error) {
	c, err := domaincategory.NewCategory(userID, name, color, icon)
	if err != nil {
		return nil, err
	}

	existing, err := uc.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao contar categorias: %w", err)
	}
	c.SetPosition(len(existing))

	if err := uc.repo.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("erro ao criar categoria: %w", err)
	}
	return c, nil
}

// ReorderCategoriesUseCase persists a new manual ordering of a user's
// categories.
type ReorderCategoriesUseCase struct {
	repo domaincategory.Repository
}

func NewReorderCategoriesUseCase(repo domaincategory.Repository) *ReorderCategoriesUseCase {
	return &ReorderCategoriesUseCase{repo: repo}
}

func (uc *ReorderCategoriesUseCase) Execute(ctx context.Context, userID uuid.UUID, orderedIDs []uuid.UUID) error {
	owned, err := uc.repo.ListByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("erro ao listar categorias: %w", err)
	}
	byID := make(map[uuid.UUID]*domaincategory.Category, len(owned))
	for _, c := range owned {
		byID[c.ID] = c
	}

	position := 0
	for _, id := range orderedIDs {
		c, ok := byID[id]
		if !ok {
			continue
		}
		if c.Position != position {
			c.SetPosition(position)
			if err := uc.repo.Update(ctx, c); err != nil {
				return fmt.Errorf("erro ao reordenar categoria: %w", err)
			}
		}
		position++
	}
	return nil
}

// UpdateCategoryUseCase renames/recolors an existing category.
type UpdateCategoryUseCase struct {
	repo domaincategory.Repository
}

func NewUpdateCategoryUseCase(repo domaincategory.Repository) *UpdateCategoryUseCase {
	return &UpdateCategoryUseCase{repo: repo}
}

func (uc *UpdateCategoryUseCase) Execute(ctx context.Context, userID, categoryID uuid.UUID, name, color, icon string) (*domaincategory.Category, error) {
	c, err := uc.repo.FindByID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar categoria: %w", err)
	}
	if c.UserID != userID {
		return nil, shared.ErrNotFound
	}
	if err := c.Rename(name); err != nil {
		return nil, err
	}
	c.SetColor(color)
	c.SetIcon(icon)
	if err := uc.repo.Update(ctx, c); err != nil {
		return nil, fmt.Errorf("erro ao atualizar categoria: %w", err)
	}
	return c, nil
}

// DeleteCategoryUseCase removes a category. Items referencing a deleted
// category simply fall back to "sem categoria" in the UI, so there's no
// cascade to the item repositories.
type DeleteCategoryUseCase struct {
	repo domaincategory.Repository
}

func NewDeleteCategoryUseCase(repo domaincategory.Repository) *DeleteCategoryUseCase {
	return &DeleteCategoryUseCase{repo: repo}
}

func (uc *DeleteCategoryUseCase) Execute(ctx context.Context, userID, categoryID uuid.UUID) error {
	c, err := uc.repo.FindByID(ctx, categoryID)
	if err != nil {
		return fmt.Errorf("erro ao buscar categoria: %w", err)
	}
	if c.UserID != userID {
		return shared.ErrNotFound
	}
	if err := uc.repo.Delete(ctx, categoryID); err != nil {
		return fmt.Errorf("erro ao remover categoria: %w", err)
	}
	return nil
}

// ListCategoriesUseCase lists every category belonging to a user.
type ListCategoriesUseCase struct {
	repo domaincategory.Repository
}

func NewListCategoriesUseCase(repo domaincategory.Repository) *ListCategoriesUseCase {
	return &ListCategoriesUseCase{repo: repo}
}

func (uc *ListCategoriesUseCase) Execute(ctx context.Context, userID uuid.UUID) ([]*domaincategory.Category, error) {
	categories, err := uc.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar categorias: %w", err)
	}
	return categories, nil
}
