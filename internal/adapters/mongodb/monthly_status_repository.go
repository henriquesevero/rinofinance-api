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

	"rinofinance-api/internal/domain/monthlystatus"
)

// MonthlyStatusRepository implements domain/monthlystatus.Repository against
// MongoDB. Each (itemType, itemID, month) has exactly one document, keyed by
// a deterministic _id so Set is a simple idempotent upsert.
type MonthlyStatusRepository struct {
	collection *mongo.Collection
}

// NewMonthlyStatusRepository wires the repository to a database handle.
func NewMonthlyStatusRepository(db *mongo.Database) *MonthlyStatusRepository {
	return &MonthlyStatusRepository{collection: db.Collection(monthlyStatusCollection)}
}

type monthlyStatusDoc struct {
	ID        string    `bson:"_id"`
	UserID    string    `bson:"user_id"`
	ItemType  string    `bson:"item_type"`
	ItemID    string    `bson:"item_id"`
	Month     string    `bson:"month"`
	Done      bool      `bson:"done"`
	UpdatedAt time.Time `bson:"updated_at"`
}

func statusDocID(itemType monthlystatus.ItemType, itemID uuid.UUID, month string) string {
	return fmt.Sprintf("%s:%s:%s", itemType, itemID.String(), month)
}

// ByMonth returns itemID -> done for the user's items of a type in a month.
func (r *MonthlyStatusRepository) ByMonth(ctx context.Context, userID uuid.UUID, itemType monthlystatus.ItemType, month string) (map[uuid.UUID]bool, error) {
	filter := bson.M{"user_id": userID.String(), "item_type": string(itemType), "month": month}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar status mensais: %w", err)
	}
	var docs []monthlyStatusDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("erro ao decodificar status mensais: %w", err)
	}
	out := make(map[uuid.UUID]bool, len(docs))
	for _, d := range docs {
		id, err := uuid.Parse(d.ItemID)
		if err != nil {
			continue
		}
		out[id] = d.Done
	}
	return out, nil
}

// Get returns the done value for one item in a month (false when unset).
func (r *MonthlyStatusRepository) Get(ctx context.Context, userID uuid.UUID, itemType monthlystatus.ItemType, itemID uuid.UUID, month string) (bool, error) {
	var doc monthlyStatusDoc
	if err := r.collection.FindOne(ctx, bson.M{"_id": statusDocID(itemType, itemID, month)}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, fmt.Errorf("erro ao buscar status mensal: %w", err)
	}
	return doc.Done, nil
}

// Set upserts one item's done value for a month.
func (r *MonthlyStatusRepository) Set(ctx context.Context, userID uuid.UUID, itemType monthlystatus.ItemType, itemID uuid.UUID, month string, done bool) error {
	doc := monthlyStatusDoc{
		ID:        statusDocID(itemType, itemID, month),
		UserID:    userID.String(),
		ItemType:  string(itemType),
		ItemID:    itemID.String(),
		Month:     month,
		Done:      done,
		UpdatedAt: time.Now().UTC(),
	}
	if _, err := r.collection.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc, options.Replace().SetUpsert(true)); err != nil {
		return fmt.Errorf("erro ao salvar status mensal: %w", err)
	}
	return nil
}
