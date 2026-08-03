package mongodb

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"rinofinance-api/internal/domain/shared"
)

func uuidPtrToString(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

func stringIDs(ids []uuid.UUID) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = id.String()
	}
	return out
}

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
	investmentProventosCollection  = "investment_proventos"
	categoriesCollection           = "categories"
	accountsCollection             = "accounts"
	accountPurchasesCollection     = "account_purchases"
	monthlyStatusCollection        = "monthly_item_status"
	wishlistSectionsCollection     = "wishlist_sections"
	wishlistItemsCollection        = "wishlist_items"
)

func toDecimal128(m shared.Money) (bson.Decimal128, error) {
	d, err := bson.ParseDecimal128(m.Decimal().String())
	if err != nil {
		return bson.Decimal128{}, fmt.Errorf("erro ao converter valor monetário para decimal128: %w", err)
	}
	return d, nil
}

func fromDecimal128(d bson.Decimal128) (shared.Money, error) {
	dec, err := decimal.NewFromString(d.String())
	if err != nil {
		return shared.Money{}, fmt.Errorf("erro ao converter decimal128 para valor monetário: %w", err)
	}
	return shared.NewMoneyFromDecimal(dec), nil
}

func checkMatchedCount(res *mongo.UpdateResult) error {
	if res.MatchedCount == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func checkDeletedCount(res *mongo.DeleteResult) error {
	if res.DeletedCount == 0 {
		return shared.ErrNotFound
	}
	return nil
}
