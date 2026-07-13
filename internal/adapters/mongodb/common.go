package mongodb

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"rinofinance-api/internal/domain/shared"
)

// uuidPtrToString renders an optional UUID as an optional string for BSON
// storage (nil stays nil, so omitempty drops the field entirely).
func uuidPtrToString(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

// stringPtrToUUID parses an optional stored string back into an optional
// UUID (nil stays nil).
func stringPtrToUUID(s *string, field string) (*uuid.UUID, error) {
	if s == nil {
		return nil, nil
	}
	id, err := uuid.Parse(*s)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter %s: %w", field, err)
	}
	return &id, nil
}

const (
	usersCollection                = "users"
	incomesCollection              = "incomes"
	expensesCollection             = "expenses"
	creditCardsCollection          = "credit_cards"
	installmentPurchasesCollection = "installment_purchases"
	subscriptionsCollection        = "subscriptions"
	investmentAssetsCollection     = "investment_assets"
	categoriesCollection           = "categories"
	accountsCollection             = "accounts"
	accountPurchasesCollection     = "account_purchases"
)

// toDecimal128 converts a domain Money value into the BSON Decimal128 used
// to store every monetary field, going through a decimal string so no
// binary floating point rounding is ever introduced.
func toDecimal128(m shared.Money) (bson.Decimal128, error) {
	d, err := bson.ParseDecimal128(m.Decimal().String())
	if err != nil {
		return bson.Decimal128{}, fmt.Errorf("erro ao converter valor monetário para decimal128: %w", err)
	}
	return d, nil
}

// fromDecimal128 converts a stored Decimal128 back into a domain Money.
func fromDecimal128(d bson.Decimal128) (shared.Money, error) {
	dec, err := decimal.NewFromString(d.String())
	if err != nil {
		return shared.Money{}, fmt.Errorf("erro ao converter decimal128 para valor monetário: %w", err)
	}
	return shared.NewMoneyFromDecimal(dec), nil
}

// checkMatchedCount returns shared.ErrNotFound if an update matched no
// document, so use cases can distinguish "nothing to update" from a
// driver error — the Mongo equivalent of checking SQL rows affected.
func checkMatchedCount(res *mongo.UpdateResult) error {
	if res.MatchedCount == 0 {
		return shared.ErrNotFound
	}
	return nil
}

// checkDeletedCount returns shared.ErrNotFound if a delete matched no
// document.
func checkDeletedCount(res *mongo.DeleteResult) error {
	if res.DeletedCount == 0 {
		return shared.ErrNotFound
	}
	return nil
}
