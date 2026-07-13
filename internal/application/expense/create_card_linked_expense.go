package expense

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
	domainexpense "rinofinance-api/internal/domain/expense"
	"rinofinance-api/internal/domain/shared"
)

// CreateCardLinkedExpenseUseCase creates an expense whose amount is
// derived from a credit card's monthly total, implementing the "Vínculo
// com Cartões" rule.
type CreateCardLinkedExpenseUseCase struct {
	expenses domainexpense.Repository
	cards    domaincard.CardRepository
}

// NewCreateCardLinkedExpenseUseCase wires the dependencies for
// CreateCardLinkedExpenseUseCase.
func NewCreateCardLinkedExpenseUseCase(expenses domainexpense.Repository, cards domaincard.CardRepository) *CreateCardLinkedExpenseUseCase {
	return &CreateCardLinkedExpenseUseCase{expenses: expenses, cards: cards}
}

// Execute verifies the target card belongs to userID before linking, so a
// user can never attach an expense to another account's card.
func (uc *CreateCardLinkedExpenseUseCase) Execute(ctx context.Context, userID, cardID uuid.UUID, name string, categoryID *uuid.UUID) (*domainexpense.Expense, error) {
	c, err := uc.cards.FindByID(ctx, cardID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar cartão: %w", err)
	}
	if c.UserID != userID {
		return nil, shared.ErrNotFound
	}

	e, err := domainexpense.NewCardLinkedExpense(userID, name, cardID)
	if err != nil {
		return nil, err
	}
	e.SetCategory(categoryID)

	existing, err := uc.expenses.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao contar saídas: %w", err)
	}
	e.SetPosition(len(existing))

	if err := uc.expenses.Create(ctx, e); err != nil {
		return nil, fmt.Errorf("erro ao criar saída vinculada ao cartão: %w", err)
	}
	return e, nil
}
