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

	domaincard "rinofinance-api/internal/domain/card"
	"rinofinance-api/internal/domain/shared"
)

// CardRepository implements domain/card.CardRepository against MongoDB.
type CardRepository struct {
	collection *mongo.Collection
}

// NewCardRepository wires a CardRepository to a database handle.
func NewCardRepository(db *mongo.Database) *CardRepository {
	return &CardRepository{collection: db.Collection(creditCardsCollection)}
}

type creditCardDoc struct {
	ID          string          `bson:"_id"`
	UserID      string          `bson:"user_id"`
	Name        string          `bson:"name"`
	Color       string          `bson:"color,omitempty"`
	LogoURL     string          `bson:"logo_url,omitempty"`
	ImageURL    string          `bson:"image_url,omitempty"`
	CreditLimit bson.Decimal128 `bson:"credit_limit"`
	DueDay      int             `bson:"due_day,omitempty"`
	Position    int             `bson:"position"`
	CreatedAt   time.Time       `bson:"created_at"`
	UpdatedAt   time.Time       `bson:"updated_at"`
}

func newCreditCardDoc(c *domaincard.CreditCard) (creditCardDoc, error) {
	creditLimit, err := toDecimal128(c.CreditLimit)
	if err != nil {
		return creditCardDoc{}, err
	}
	return creditCardDoc{
		ID:          c.ID.String(),
		UserID:      c.UserID.String(),
		Name:        c.Name,
		Color:       c.Color,
		LogoURL:     c.LogoURL,
		ImageURL:    c.ImageURL,
		CreditLimit: creditLimit,
		DueDay:      c.DueDay,
		Position:    c.Position,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}, nil
}

func (d creditCardDoc) toDomain() (*domaincard.CreditCard, error) {
	id, err := uuid.Parse(d.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de cartão: %w", err)
	}
	userID, err := uuid.Parse(d.UserID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de usuário do cartão: %w", err)
	}
	creditLimit, err := fromDecimal128(d.CreditLimit)
	if err != nil {
		return nil, err
	}
	return &domaincard.CreditCard{
		ID:          id,
		UserID:      userID,
		Name:        d.Name,
		Color:       d.Color,
		LogoURL:     d.LogoURL,
		ImageURL:    d.ImageURL,
		CreditLimit: creditLimit,
		DueDay:      d.DueDay,
		Position:    d.Position,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}, nil
}

// Create inserts a new credit card document.
func (r *CardRepository) Create(ctx context.Context, c *domaincard.CreditCard) error {
	doc, err := newCreditCardDoc(c)
	if err != nil {
		return err
	}
	if _, err := r.collection.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("erro ao inserir cartão: %w", err)
	}
	return nil
}

// FindByID fetches a credit card by ID.
func (r *CardRepository) FindByID(ctx context.Context, id uuid.UUID) (*domaincard.CreditCard, error) {
	var doc creditCardDoc
	if err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("erro ao buscar cartão: %w", err)
	}
	return doc.toDomain()
}

// ListByUser fetches every credit card belonging to userID, ordered by
// the user's manual position (with created_at as a stable tiebreaker).
func (r *CardRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domaincard.CreditCard, error) {
	opts := options.Find().SetSort(bson.D{{Key: "position", Value: 1}, {Key: "created_at", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID.String()}, opts)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar cartões: %w", err)
	}

	var docs []creditCardDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("erro ao decodificar cartões: %w", err)
	}

	cards := make([]*domaincard.CreditCard, 0, len(docs))
	for _, doc := range docs {
		c, err := doc.toDomain()
		if err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, nil
}

// Update persists changes to a credit card's name.
func (r *CardRepository) Update(ctx context.Context, c *domaincard.CreditCard) error {
	doc, err := newCreditCardDoc(c)
	if err != nil {
		return err
	}
	res, err := r.collection.ReplaceOne(ctx, bson.M{"_id": c.ID.String()}, doc)
	if err != nil {
		return fmt.Errorf("erro ao atualizar cartão: %w", err)
	}
	return checkMatchedCount(res)
}

// Delete permanently removes a credit card document. MongoDB has no
// foreign key cascade, so application/card.DeleteCardUseCase is
// responsible for removing the card's installment purchases and
// subscriptions and unlinking its expenses before calling this method.
func (r *CardRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return fmt.Errorf("erro ao remover cartão: %w", err)
	}
	return checkDeletedCount(res)
}
