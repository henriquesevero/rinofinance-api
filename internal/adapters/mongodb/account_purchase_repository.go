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

	domainaccount "rinofinance-api/internal/domain/account"
	"rinofinance-api/internal/domain/shared"
)

// AccountPurchaseRepository implements domain/account.PurchaseRepository
// against MongoDB.
type AccountPurchaseRepository struct {
	collection *mongo.Collection
}

// NewAccountPurchaseRepository wires the repository to a database handle.
func NewAccountPurchaseRepository(db *mongo.Database) *AccountPurchaseRepository {
	return &AccountPurchaseRepository{collection: db.Collection(accountPurchasesCollection)}
}

type accountPurchaseDoc struct {
	ID         string          `bson:"_id"`
	AccountID  string          `bson:"account_id"`
	Name       string          `bson:"name"`
	Amount     bson.Decimal128 `bson:"amount"`
	Date       time.Time       `bson:"date"`
	CategoryID *string         `bson:"category_id,omitempty"`
	Direction  string          `bson:"direction,omitempty"`
	ExternalID string          `bson:"external_id,omitempty"`
	Position   int             `bson:"position"`
	CreatedAt  time.Time       `bson:"created_at"`
	UpdatedAt  time.Time       `bson:"updated_at"`
}

func newAccountPurchaseDoc(p *domainaccount.Purchase) (accountPurchaseDoc, error) {
	amount, err := toDecimal128(p.Amount)
	if err != nil {
		return accountPurchaseDoc{}, err
	}
	return accountPurchaseDoc{
		ID:         p.ID.String(),
		AccountID:  p.AccountID.String(),
		Name:       p.Name,
		Amount:     amount,
		Date:       p.Date,
		CategoryID: uuidPtrToString(p.CategoryID),
		Direction:  p.Direction,
		ExternalID: p.ExternalID,
		Position:   p.Position,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}, nil
}

func (d accountPurchaseDoc) toDomain() (*domainaccount.Purchase, error) {
	id, err := uuid.Parse(d.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de compra da conta: %w", err)
	}
	accountID, err := uuid.Parse(d.AccountID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de conta da compra: %w", err)
	}
	amount, err := fromDecimal128(d.Amount)
	if err != nil {
		return nil, err
	}
	categoryID, err := stringPtrToUUID(d.CategoryID, "ID de categoria da compra da conta")
	if err != nil {
		return nil, err
	}
	return &domainaccount.Purchase{
		ID:         id,
		AccountID:  accountID,
		Name:       d.Name,
		Amount:     amount,
		Date:       d.Date,
		CategoryID: categoryID,
		Direction:  d.Direction,
		ExternalID: d.ExternalID,
		Position:   d.Position,
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
	}, nil
}

// Create inserts a new account purchase document.
func (r *AccountPurchaseRepository) Create(ctx context.Context, p *domainaccount.Purchase) error {
	doc, err := newAccountPurchaseDoc(p)
	if err != nil {
		return err
	}
	if _, err := r.collection.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("erro ao inserir compra da conta: %w", err)
	}
	return nil
}

// FindByID fetches an account purchase by ID.
func (r *AccountPurchaseRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainaccount.Purchase, error) {
	var doc accountPurchaseDoc
	if err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("erro ao buscar compra da conta: %w", err)
	}
	return doc.toDomain()
}

// ListByAccount fetches every purchase for an account, ordered by the
// user's manual position (with date as a stable tiebreaker).
func (r *AccountPurchaseRepository) ListByAccount(ctx context.Context, accountID uuid.UUID) ([]*domainaccount.Purchase, error) {
	opts := options.Find().SetSort(bson.D{{Key: "position", Value: 1}, {Key: "date", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{"account_id": accountID.String()}, opts)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar compras da conta: %w", err)
	}

	var docs []accountPurchaseDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("erro ao decodificar compras da conta: %w", err)
	}

	purchases := make([]*domainaccount.Purchase, 0, len(docs))
	for _, doc := range docs {
		p, err := doc.toDomain()
		if err != nil {
			return nil, err
		}
		purchases = append(purchases, p)
	}
	return purchases, nil
}

// Update persists changes to an account purchase.
func (r *AccountPurchaseRepository) Update(ctx context.Context, p *domainaccount.Purchase) error {
	doc, err := newAccountPurchaseDoc(p)
	if err != nil {
		return err
	}
	res, err := r.collection.ReplaceOne(ctx, bson.M{"_id": p.ID.String()}, doc)
	if err != nil {
		return fmt.Errorf("erro ao atualizar compra da conta: %w", err)
	}
	return checkMatchedCount(res)
}

// Delete permanently removes an account purchase document.
func (r *AccountPurchaseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return fmt.Errorf("erro ao remover compra da conta: %w", err)
	}
	return checkDeletedCount(res)
}

// DeleteByAccount removes every purchase belonging to an account, used when
// the account itself is deleted.
func (r *AccountPurchaseRepository) DeleteByAccount(ctx context.Context, accountID uuid.UUID) error {
	if _, err := r.collection.DeleteMany(ctx, bson.M{"account_id": accountID.String()}); err != nil {
		return fmt.Errorf("erro ao remover compras da conta: %w", err)
	}
	return nil
}
