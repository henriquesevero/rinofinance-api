package expense

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
	domainexpense "rinofinance-api/internal/domain/expense"
	"rinofinance-api/internal/domain/shared"
)

type CreateCardLinkedExpenseUseCase struct {
	expenses domainexpense.Repository
	cards    domaincard.CardRepository
}

func NewCreateCardLinkedExpenseUseCase(expenses domainexpense.Repository, cards domaincard.CardRepository) *CreateCardLinkedExpenseUseCase {
	return &CreateCardLinkedExpenseUseCase{expenses: expenses, cards: cards}
}

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
