package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"rinofinance-api/internal/domain/notification"
)

type PushSubscriptionRepository struct {
	collection *mongo.Collection
}

func NewPushSubscriptionRepository(db *mongo.Database) *PushSubscriptionRepository {
	return &PushSubscriptionRepository{collection: db.Collection(pushSubscriptionsCollection)}
}

type pushSubscriptionDoc struct {
	ID        string    `bson:"_id"`
	UserID    string    `bson:"user_id"`
	Endpoint  string    `bson:"endpoint"`
	P256DH    string    `bson:"p256dh"`
	Auth      string    `bson:"auth"`
	CreatedAt time.Time `bson:"created_at"`
}

func (d pushSubscriptionDoc) toDomain() (*notification.PushSubscription, error) {
	id, err := uuid.Parse(d.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de inscrição: %w", err)
	}
	userID, err := uuid.Parse(d.UserID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de usuário da inscrição: %w", err)
	}
	return &notification.PushSubscription{
		ID:        id,
		UserID:    userID,
		Endpoint:  d.Endpoint,
		P256DH:    d.P256DH,
		Auth:      d.Auth,
		CreatedAt: d.CreatedAt,
	}, nil
}

func (r *PushSubscriptionRepository) Save(ctx context.Context, sub *notification.PushSubscription) error {
	doc := pushSubscriptionDoc{
		ID:        sub.ID.String(),
		UserID:    sub.UserID.String(),
		Endpoint:  sub.Endpoint,
		P256DH:    sub.P256DH,
		Auth:      sub.Auth,
		CreatedAt: sub.CreatedAt,
	}
	opts := options.Replace().SetUpsert(true)
	if _, err := r.collection.ReplaceOne(ctx, bson.M{"endpoint": sub.Endpoint}, doc, opts); err != nil {
		return fmt.Errorf("erro ao salvar inscrição de notificação: %w", err)
	}
	return nil
}

func (r *PushSubscriptionRepository) DeleteByEndpoint(ctx context.Context, userID uuid.UUID, endpoint string) error {
	if _, err := r.collection.DeleteOne(ctx, bson.M{"user_id": userID.String(), "endpoint": endpoint}); err != nil {
		return fmt.Errorf("erro ao remover inscrição de notificação: %w", err)
	}
	return nil
}

func (r *PushSubscriptionRepository) ListAll(ctx context.Context) ([]*notification.PushSubscription, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("erro ao listar inscrições: %w", err)
	}
	var docs []pushSubscriptionDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("erro ao decodificar inscrições: %w", err)
	}
	subs := make([]*notification.PushSubscription, 0, len(docs))
	for _, doc := range docs {
		sub, err := doc.toDomain()
		if err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

func (r *PushSubscriptionRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	if _, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.String()}); err != nil {
		return fmt.Errorf("erro ao remover inscrição: %w", err)
	}
	return nil
}
