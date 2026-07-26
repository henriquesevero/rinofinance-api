package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// EnsureIndexes creates every index the application relies on. MongoDB is
// schemaless, so this replaces what a relational schema would express as
// migrations: the unique index on users.email enforces the same
// constraint a SQL UNIQUE column would, and the user_id/card_id indexes
// keep the per-user and per-card list queries fast. It is idempotent —
// safe to call on every boot.
func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	if _, err := db.Collection(usersCollection).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return fmt.Errorf("erro ao criar índice único de email: %w", err)
	}

	userScoped := func(field string) mongo.IndexModel {
		return mongo.IndexModel{Keys: bson.D{{Key: field, Value: 1}}}
	}

	indexesByCollection := map[string][]mongo.IndexModel{
		incomesCollection:              {userScoped("user_id")},
		expensesCollection:             {userScoped("user_id"), userScoped("card_id")},
		creditCardsCollection:          {userScoped("user_id")},
		installmentPurchasesCollection: {userScoped("card_id")},
		subscriptionsCollection:        {userScoped("card_id")},
		investmentAssetsCollection:     {userScoped("user_id")},
		categoriesCollection:           {userScoped("user_id")},
		accountsCollection:             {userScoped("user_id")},
		accountPurchasesCollection:     {userScoped("account_id")},
		monthlyStatusCollection: {
			{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "item_type", Value: 1}, {Key: "month", Value: 1}}},
		},
	}

	for collection, models := range indexesByCollection {
		if _, err := db.Collection(collection).Indexes().CreateMany(ctx, models); err != nil {
			return fmt.Errorf("erro ao criar índices da coleção %s: %w", collection, err)
		}
	}

	if err := backfillPositions(ctx, db); err != nil {
		return err
	}

	return nil
}

// backfillPositions gives every document created before manual ordering
// existed a real position of 0. Without this, legacy docs have no position
// field (which MongoDB sorts as null, i.e. before 0), so editing one would
// write position 0 and jump it past the un-migrated ones. Idempotent:
// once a doc has the field it is skipped.
func backfillPositions(ctx context.Context, db *mongo.Database) error {
	orderedCollections := []string{
		incomesCollection,
		expensesCollection,
		creditCardsCollection,
		installmentPurchasesCollection,
		subscriptionsCollection,
		categoriesCollection,
		accountsCollection,
		accountPurchasesCollection,
	}
	for _, collection := range orderedCollections {
		_, err := db.Collection(collection).UpdateMany(
			ctx,
			bson.M{"position": bson.M{"$exists": false}},
			bson.M{"$set": bson.M{"position": 0}},
		)
		if err != nil {
			return fmt.Errorf("erro ao preencher position em %s: %w", collection, err)
		}
	}
	return nil
}
