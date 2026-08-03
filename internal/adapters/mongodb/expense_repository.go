package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	domainexpense "rinofinance-api/internal/domain/expense"
	"rinofinance-api/internal/domain/shared"
)

type ExpenseRepository struct {
	collection *mongo.Collection
}

func NewExpenseRepository(db *mongo.Database) *ExpenseRepository {
	return &ExpenseRepository{collection: db.Collection(expensesCollection)}
}

type expenseDoc struct {
	ID         string          `bson:"_id"`
	UserID     string          `bson:"user_id"`
	Name       string          `bson:"name"`
	Amount     bson.Decimal128 `bson:"amount"`
	Active     bool            `bson:"active"`
	Paid       bool            `bson:"paid,omitempty"`
	CardID     *string         `bson:"card_id,omitempty"`
	CategoryID *string         `bson:"category_id,omitempty"`
	AccountID  *string         `bson:"account_id,omitempty"`
	Position   int             `bson:"position"`
	CreatedAt  time.Time       `bson:"created_at"`
	UpdatedAt  time.Time       `bson:"updated_at"`
}

func newExpenseDoc(e *domainexpense.Expense) (expenseDoc, error) {
	amount, err := toDecimal128(e.Amount)
	if err != nil {
		return expenseDoc{}, err
	}
	var cardID *string
	if e.CardID != nil {
		s := e.CardID.String()
		cardID = &s
	}
	return expenseDoc{
		ID:         e.ID.String(),
		UserID:     e.UserID.String(),
		Name:       e.Name,
		Amount:     amount,
		Active:     e.Active,
		Paid:       e.Paid,
		CardID:     cardID,
		CategoryID: uuidPtrToString(e.CategoryID),
		AccountID:  uuidPtrToString(e.AccountID),
		Position:   e.Position,
		CreatedAt:  e.CreatedAt,
		UpdatedAt:  e.UpdatedAt,
	}, nil
}

func (d expenseDoc) toDomain() (*domainexpense.Expense, error) {
	id, err := uuid.Parse(d.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de saída: %w", err)
	}
	userID, err := uuid.Parse(d.UserID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de usuário da saída: %w", err)
	}
	amount, err := fromDecimal128(d.Amount)
	if err != nil {
		return nil, err
	}

	var cardID *uuid.UUID
	if d.CardID != nil {
		parsed, err := uuid.Parse(*d.CardID)
		if err != nil {
			return nil, fmt.Errorf("erro ao converter ID de cartão vinculado: %w", err)
		}
		cardID = &parsed
	}
	categoryID, err := stringPtrToUUID(d.CategoryID, "ID de categoria da saída")
	if err != nil {
		return nil, err
	}
	accountID, err := stringPtrToUUID(d.AccountID, "ID de conta da saída")
	if err != nil {
		return nil, err
	}

	return &domainexpense.Expense{
		ID:         id,
		UserID:     userID,
		Name:       d.Name,
		Amount:     amount,
		Active:     d.Active,
		Paid:       d.Paid,
		CardID:     cardID,
		CategoryID: categoryID,
		AccountID:  accountID,
		Position:   d.Position,
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
	}, nil
}

func (r *ExpenseRepository) Create(ctx context.Context, e *domainexpense.Expense) error {
	doc, err := newExpenseDoc(e)
	if err != nil {
		return err
	}
	if _, err := r.collection.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("erro ao inserir saída: %w", err)
	}
	return nil
}

func (r *ExpenseRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainexpense.Expense, error) {
	var doc expenseDoc
	if err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("erro ao buscar saída: %w", err)
	}
	return doc.toDomain()
}

func (r *ExpenseRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domainexpense.Expense, error) {
	return r.list(ctx, bson.M{"user_id": userID.String()})
}

func (r *ExpenseRepository) FindByCardID(ctx context.Context, cardID uuid.UUID) ([]*domainexpense.Expense, error) {
	return r.list(ctx, bson.M{"card_id": cardID.String()})
}

func (r *ExpenseRepository) list(ctx context.Context, filter bson.M) ([]*domainexpense.Expense, error) {
	opts := options.Find().SetSort(bson.D{{Key: "position", Value: 1}, {Key: "created_at", Value: 1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar saídas: %w", err)
	}

	var docs []expenseDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("erro ao decodificar saídas: %w", err)
	}

	expenses := make([]*domainexpense.Expense, 0, len(docs))
	for _, doc := range docs {
		e, err := doc.toDomain()
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, e)
	}
	return expenses, nil
}

func (r *ExpenseRepository) Update(ctx context.Context, e *domainexpense.Expense) error {
	doc, err := newExpenseDoc(e)
	if err != nil {
		return err
	}
	res, err := r.collection.ReplaceOne(ctx, bson.M{"_id": e.ID.String()}, doc)
	if err != nil {
		return fmt.Errorf("erro ao atualizar saída: %w", err)
	}
	return checkMatchedCount(res)
}

func (r *ExpenseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return fmt.Errorf("erro ao remover saída: %w", err)
	}
	return checkDeletedCount(res)
}
