package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	domaininvestment "rinofinance-api/internal/domain/investment"
	"rinofinance-api/internal/domain/shared"
)

// InvestmentRepository implements domain/investment.Repository against
// MongoDB.
type InvestmentRepository struct {
	collection *mongo.Collection
}

// NewInvestmentRepository wires an InvestmentRepository to a database
// handle.
func NewInvestmentRepository(db *mongo.Database) *InvestmentRepository {
	return &InvestmentRepository{collection: db.Collection(investmentAssetsCollection)}
}

type investmentAssetDoc struct {
	ID             string          `bson:"_id"`
	UserID         string          `bson:"user_id"`
	Name           string          `bson:"name"`
	CurrentBalance bson.Decimal128 `bson:"current_balance"`
	// Active is a pointer so documents created before this field existed
	// (which have no "active" key) decode as nil and are treated as
	// active, rather than defaulting to false and silently vanishing from
	// the patrimony total.
	Active    *bool     `bson:"active,omitempty"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

func newInvestmentAssetDoc(a *domaininvestment.Asset) (investmentAssetDoc, error) {
	balance, err := toDecimal128(a.CurrentBalance)
	if err != nil {
		return investmentAssetDoc{}, err
	}
	active := a.Active
	return investmentAssetDoc{
		ID:             a.ID.String(),
		UserID:         a.UserID.String(),
		Name:           a.Name,
		CurrentBalance: balance,
		Active:         &active,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}, nil
}

func (d investmentAssetDoc) toDomain() (*domaininvestment.Asset, error) {
	id, err := uuid.Parse(d.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de ativo: %w", err)
	}
	userID, err := uuid.Parse(d.UserID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de usuário do ativo: %w", err)
	}
	balance, err := fromDecimal128(d.CurrentBalance)
	if err != nil {
		return nil, err
	}
	// Legacy documents without the field are considered active.
	active := d.Active == nil || *d.Active
	return &domaininvestment.Asset{
		ID:             id,
		UserID:         userID,
		Name:           d.Name,
		CurrentBalance: balance,
		Active:         active,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}, nil
}

// Create inserts a new investment asset document.
func (r *InvestmentRepository) Create(ctx context.Context, a *domaininvestment.Asset) error {
	doc, err := newInvestmentAssetDoc(a)
	if err != nil {
		return err
	}
	if _, err := r.collection.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("erro ao inserir ativo: %w", err)
	}
	return nil
}

// FindByID fetches an investment asset by ID.
func (r *InvestmentRepository) FindByID(ctx context.Context, id uuid.UUID) (*domaininvestment.Asset, error) {
	var doc investmentAssetDoc
	if err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("erro ao buscar ativo: %w", err)
	}
	return doc.toDomain()
}

// ListByUser fetches every investment asset belonging to userID.
func (r *InvestmentRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domaininvestment.Asset, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID.String()})
	if err != nil {
		return nil, fmt.Errorf("erro ao listar ativos: %w", err)
	}

	var docs []investmentAssetDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("erro ao decodificar ativos: %w", err)
	}

	assets := make([]*domaininvestment.Asset, 0, len(docs))
	for _, doc := range docs {
		a, err := doc.toDomain()
		if err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}
	return assets, nil
}

// Update persists changes to an asset's name and current balance.
func (r *InvestmentRepository) Update(ctx context.Context, a *domaininvestment.Asset) error {
	doc, err := newInvestmentAssetDoc(a)
	if err != nil {
		return err
	}
	res, err := r.collection.ReplaceOne(ctx, bson.M{"_id": a.ID.String()}, doc)
	if err != nil {
		return fmt.Errorf("erro ao atualizar ativo: %w", err)
	}
	return checkMatchedCount(res)
}

// Delete permanently removes an investment asset document.
func (r *InvestmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return fmt.Errorf("erro ao remover ativo: %w", err)
	}
	return checkDeletedCount(res)
}
