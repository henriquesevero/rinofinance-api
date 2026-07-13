// Package mongodb implements every domain repository port against
// MongoDB (hosted on MongoDB Atlas in production). Each aggregate maps to
// one collection; IDs are stored as the string form of the domain's
// uuid.UUID (not Mongo's ObjectID) so the domain layer never needs to
// know about the persistence technology, and monetary fields are stored
// as BSON Decimal128 to avoid floating point rounding.
package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Connect opens a connection to MongoDB Atlas (or any MongoDB deployment)
// at uri, verifies it with a ping, and returns both the client (needed
// only so main.go can defer its Disconnect) and a database handle scoped
// to databaseName.
func Connect(uri, databaseName string) (*mongo.Client, *mongo.Database, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao conectar ao MongoDB: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		return nil, nil, fmt.Errorf("erro ao verificar conexão com o MongoDB: %w", err)
	}

	return client, client.Database(databaseName), nil
}
