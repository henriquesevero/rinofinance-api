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

	domainincome "rinofinance-api/internal/domain/income"
	"rinofinance-api/internal/domain/shared"
)

type IncomeRepository struct {
	collection *mongo.Collection
}

func NewIncomeRepository(db *mongo.Database) *IncomeRepository {
	return &IncomeRepository{collection: db.Collection(incomesCollection)}
}

type incomeDoc struct {
	ID         string          `bson:"_id"`
	UserID     string          `bson:"user_id"`
	Name       string          `bson:"name"`
	Amount     bson.Decimal128 `bson:"amount"`
	Active     bool            `bson:"active"`
	Received   bool            `bson:"received,omitempty"`
	CategoryID *string         `bson:"category_id,omitempty"`
	AccountID  *string         `bson:"account_id,omitempty"`
	Position   int             `bson:"position"`
	CreatedAt  time.Time       `bson:"created_at"`
	UpdatedAt  time.Time       `bson:"updated_at"`
}

func newIncomeDoc(i *domainincome.Income) (incomeDoc, error) {
	amount, err := toDecimal128(i.Amount)
	if err != nil {
		return incomeDoc{}, err
	}
	return incomeDoc{
		ID:         i.ID.String(),
		UserID:     i.UserID.String(),
		Name:       i.Name,
		Amount:     amount,
		Active:     i.Active,
		Received:   i.Received,
		CategoryID: uuidPtrToString(i.CategoryID),
		AccountID:  uuidPtrToString(i.AccountID),
		Position:   i.Position,
		CreatedAt:  i.CreatedAt,
		UpdatedAt:  i.UpdatedAt,
	}, nil
}

func (d incomeDoc) toDomain() (*domainincome.Income, error) {
	id, err := uuid.Parse(d.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de entrada: %w", err)
	}
	userID, err := uuid.Parse(d.UserID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de usuário da entrada: %w", err)
	}
	amount, err := fromDecimal128(d.Amount)
	if err != nil {
		return nil, err
	}
	categoryID, err := stringPtrToUUID(d.CategoryID, "ID de categoria da entrada")
	if err != nil {
		return nil, err
	}
	accountID, err := stringPtrToUUID(d.AccountID, "ID de conta da entrada")
	if err != nil {
		return nil, err
	}
	return &domainincome.Income{
		ID:         id,
		UserID:     userID,
		Name:       d.Name,
		Amount:     amount,
		Active:     d.Active,
		Received:   d.Received,
		CategoryID: categoryID,
		AccountID:  accountID,
		Position:   d.Position,
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
	}, nil
}

func (r *IncomeRepository) Create(ctx context.Context, i *domainincome.Income) error {
	doc, err := newIncomeDoc(i)
	if err != nil {
		return err
	}
	if _, err := r.collection.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("erro ao inserir entrada: %w", err)
	}
	return nil
}

func (r *IncomeRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainincome.Income, error) {
	var doc incomeDoc
	if err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("erro ao buscar entrada: %w", err)
	}
	return doc.toDomain()
}

func (r *IncomeRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domainincome.Income, error) {
	opts := options.Find().SetSort(bson.D{{Key: "position", Value: 1}, {Key: "created_at", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID.String()}, opts)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar entradas: %w", err)
	}

	var docs []incomeDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("erro ao decodificar entradas: %w", err)
	}

	incomes := make([]*domainincome.Income, 0, len(docs))
	for _, doc := range docs {
		inc, err := doc.toDomain()
		if err != nil {
			return nil, err
		}
		incomes = append(incomes, inc)
	}
	return incomes, nil
}

func (r *IncomeRepository) Update(ctx context.Context, i *domainincome.Income) error {
	doc, err := newIncomeDoc(i)
	if err != nil {
		return err
	}
	res, err := r.collection.ReplaceOne(ctx, bson.M{"_id": i.ID.String()}, doc)
	if err != nil {
		return fmt.Errorf("erro ao atualizar entrada: %w", err)
	}
	return checkMatchedCount(res)
}

func (r *IncomeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return fmt.Errorf("erro ao remover entrada: %w", err)
	}
	return checkDeletedCount(res)
}
