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

type SubscriptionRepository struct {
	collection *mongo.Collection
}

func NewSubscriptionRepository(db *mongo.Database) *SubscriptionRepository {
	return &SubscriptionRepository{collection: db.Collection(subscriptionsCollection)}
}

type subscriptionDoc struct {
	ID            string          `bson:"_id"`
	CardID        string          `bson:"card_id"`
	Name          string          `bson:"name"`
	MonthlyAmount bson.Decimal128 `bson:"monthly_amount"`
	Domain        string          `bson:"domain,omitempty"`
	CategoryID    *string         `bson:"category_id,omitempty"`
	CanceledFrom  *time.Time      `bson:"canceled_from,omitempty"`
	EffectiveFrom *time.Time      `bson:"effective_from,omitempty"`
	Position      int             `bson:"position"`
	CreatedAt     time.Time       `bson:"created_at"`
	UpdatedAt     time.Time       `bson:"updated_at"`
}

func newSubscriptionDoc(s *domaincard.Subscription) (subscriptionDoc, error) {
	amount, err := toDecimal128(s.MonthlyAmount)
	if err != nil {
		return subscriptionDoc{}, err
	}
	return subscriptionDoc{
		ID:            s.ID.String(),
		CardID:        s.CardID.String(),
		Name:          s.Name,
		MonthlyAmount: amount,
		Domain:        s.Domain,
		CategoryID:    uuidPtrToString(s.CategoryID),
		CanceledFrom:  s.CanceledFrom,
		EffectiveFrom: s.EffectiveFrom,
		Position:      s.Position,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
	}, nil
}

func (d subscriptionDoc) toDomain() (*domaincard.Subscription, error) {
	id, err := uuid.Parse(d.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de assinatura: %w", err)
	}
	cardID, err := uuid.Parse(d.CardID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de cartão da assinatura: %w", err)
	}
	amount, err := fromDecimal128(d.MonthlyAmount)
	if err != nil {
		return nil, err
	}
	categoryID, err := stringPtrToUUID(d.CategoryID, "ID de categoria da assinatura")
	if err != nil {
		return nil, err
	}
	return &domaincard.Subscription{
		ID:            id,
		CardID:        cardID,
		Name:          d.Name,
		MonthlyAmount: amount,
		Domain:        d.Domain,
		CategoryID:    categoryID,
		CanceledFrom:  d.CanceledFrom,
		EffectiveFrom: d.EffectiveFrom,
		Position:      d.Position,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}, nil
}

func (r *SubscriptionRepository) Create(ctx context.Context, s *domaincard.Subscription) error {
	doc, err := newSubscriptionDoc(s)
	if err != nil {
		return err
	}
	if _, err := r.collection.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("erro ao inserir assinatura: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domaincard.Subscription, error) {
	var doc subscriptionDoc
	if err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("erro ao buscar assinatura: %w", err)
	}
	return doc.toDomain()
}

func (r *SubscriptionRepository) ListByCard(ctx context.Context, cardID uuid.UUID) ([]*domaincard.Subscription, error) {
	return r.list(ctx, bson.M{"card_id": cardID.String()})
}

func (r *SubscriptionRepository) ListByCards(ctx context.Context, cardIDs []uuid.UUID) ([]*domaincard.Subscription, error) {
	if len(cardIDs) == 0 {
		return nil, nil
	}
	return r.list(ctx, bson.M{"card_id": bson.M{"$in": stringIDs(cardIDs)}})
}

func (r *SubscriptionRepository) list(ctx context.Context, filter bson.M) ([]*domaincard.Subscription, error) {
	opts := options.Find().SetSort(bson.D{{Key: "position", Value: 1}, {Key: "created_at", Value: 1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar assinaturas: %w", err)
	}

	var docs []subscriptionDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("erro ao decodificar assinaturas: %w", err)
	}

	subscriptions := make([]*domaincard.Subscription, 0, len(docs))
	for _, doc := range docs {
		s, err := doc.toDomain()
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, s)
	}
	return subscriptions, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, s *domaincard.Subscription) error {
	doc, err := newSubscriptionDoc(s)
	if err != nil {
		return err
	}
	res, err := r.collection.ReplaceOne(ctx, bson.M{"_id": s.ID.String()}, doc)
	if err != nil {
		return fmt.Errorf("erro ao atualizar assinatura: %w", err)
	}
	return checkMatchedCount(res)
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return fmt.Errorf("erro ao remover assinatura: %w", err)
	}
	return checkDeletedCount(res)
}
