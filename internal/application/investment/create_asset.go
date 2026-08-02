// Package investment orchestrates CRUD use cases for Aba 3's investment
// portfolio (positions and their proventos).
package investment

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaininvestment "rinofinance-api/internal/domain/investment"
)

// CreateAssetUseCase creates a new portfolio asset for a user.
type CreateAssetUseCase struct {
	repo domaininvestment.Repository
}

// NewCreateAssetUseCase wires the dependencies for CreateAssetUseCase.
func NewCreateAssetUseCase(repo domaininvestment.Repository) *CreateAssetUseCase {
	return &CreateAssetUseCase{repo: repo}
}

// Execute builds and persists a new Asset from the given input.
func (uc *CreateAssetUseCase) Execute(ctx context.Context, userID uuid.UUID, in domaininvestment.AssetInput) (*domaininvestment.Asset, error) {
	a, err := domaininvestment.NewAsset(userID, in)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Create(ctx, a); err != nil {
		return nil, fmt.Errorf("erro ao criar ativo: %w", err)
	}
	return a, nil
}
