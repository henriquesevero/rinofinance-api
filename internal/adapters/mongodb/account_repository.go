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

type AccountRepository struct {
	collection *mongo.Collection
}

func NewAccountRepository(db *mongo.Database) *AccountRepository {
	return &AccountRepository{collection: db.Collection(accountsCollection)}
}

type accountDoc struct {
	ID            string          `bson:"_id"`
	UserID        string          `bson:"user_id"`
	Name          string          `bson:"name"`
	Color         string          `bson:"color,omitempty"`
	ImageURL      string          `bson:"image_url,omitempty"`
	Balance       bson.Decimal128 `bson:"balance"`
	Agency        string          `bson:"agency,omitempty"`
	AccountNumber string          `bson:"account_number,omitempty"`
	AccountType   string          `bson:"account_type,omitempty"`
	Position      int             `bson:"position"`
	CreatedAt     time.Time       `bson:"created_at"`
	UpdatedAt     time.Time       `bson:"updated_at"`
}

func newAccountDoc(a *domainaccount.Account) (accountDoc, error) {
	balance, err := toDecimal128(a.Balance)
	if err != nil {
		return accountDoc{}, err
	}
	return accountDoc{
		ID:            a.ID.String(),
		UserID:        a.UserID.String(),
		Name:          a.Name,
		Color:         a.Color,
		ImageURL:      a.ImageURL,
		Balance:       balance,
		Agency:        a.Agency,
		AccountNumber: a.AccountNumber,
		AccountType:   a.AccountType,
		Position:      a.Position,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
	}, nil
}

func (d accountDoc) toDomain() (*domainaccount.Account, error) {
	id, err := uuid.Parse(d.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de conta: %w", err)
	}
	userID, err := uuid.Parse(d.UserID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de usuário da conta: %w", err)
	}
	balance, err := fromDecimal128(d.Balance)
	if err != nil {
		return nil, err
	}
	return &domainaccount.Account{
		ID:            id,
		UserID:        userID,
		Name:          d.Name,
		Color:         d.Color,
		ImageURL:      d.ImageURL,
		Balance:       balance,
		Agency:        d.Agency,
		AccountNumber: d.AccountNumber,
		AccountType:   d.AccountType,
		Position:      d.Position,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}, nil
}

func (r *AccountRepository) Create(ctx context.Context, a *domainaccount.Account) error {
	doc, err := newAccountDoc(a)
	if err != nil {
		return err
	}
	if _, err := r.collection.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("erro ao inserir conta: %w", err)
	}
	return nil
}

func (r *AccountRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainaccount.Account, error) {
	var doc accountDoc
	if err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("erro ao buscar conta: %w", err)
	}
	return doc.toDomain()
}

func (r *AccountRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domainaccount.Account, error) {
	opts := options.Find().SetSort(bson.D{{Key: "position", Value: 1}, {Key: "name", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID.String()}, opts)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar contas: %w", err)
	}

	var docs []accountDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("erro ao decodificar contas: %w", err)
	}

	accounts := make([]*domainaccount.Account, 0, len(docs))
	for _, doc := range docs {
		a, err := doc.toDomain()
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

func (r *AccountRepository) Update(ctx context.Context, a *domainaccount.Account) error {
	doc, err := newAccountDoc(a)
	if err != nil {
		return err
	}
	res, err := r.collection.ReplaceOne(ctx, bson.M{"_id": a.ID.String()}, doc)
	if err != nil {
		return fmt.Errorf("erro ao atualizar conta: %w", err)
	}
	return checkMatchedCount(res)
}

func (r *AccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return fmt.Errorf("erro ao remover conta: %w", err)
	}
	return checkDeletedCount(res)
}
