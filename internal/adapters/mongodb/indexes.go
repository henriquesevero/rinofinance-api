package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

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
		investmentProventosCollection:  {userScoped("user_id"), userScoped("asset_id")},
		categoriesCollection:           {userScoped("user_id")},
		accountsCollection:             {userScoped("user_id")},
		accountPurchasesCollection:     {userScoped("account_id")},
		wishlistSectionsCollection:     {userScoped("user_id")},
		wishlistItemsCollection:        {userScoped("user_id")},
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

	for _, col := range []string{wishlistSectionsCollection, wishlistItemsCollection} {
		if _, err := db.Collection(col).UpdateMany(
			ctx,
			bson.M{"kind": bson.M{"$exists": false}},
			bson.M{"$set": bson.M{"kind": "wishlist"}},
		); err != nil {
			return fmt.Errorf("erro ao preencher kind em %s: %w", col, err)
		}
	}

	return nil
}

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
