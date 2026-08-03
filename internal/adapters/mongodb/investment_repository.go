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

	domaininvestment "rinofinance-api/internal/domain/investment"
	"rinofinance-api/internal/domain/shared"
)

type InvestmentRepository struct {
	collection *mongo.Collection
}

func NewInvestmentRepository(db *mongo.Database) *InvestmentRepository {
	return &InvestmentRepository{collection: db.Collection(investmentAssetsCollection)}
}

type investmentAssetDoc struct {
	ID     string `bson:"_id"`
	UserID string `bson:"user_id"`
	Name   string `bson:"name"`

	Ticker         *string          `bson:"ticker,omitempty"`
	Class          *string          `bson:"class,omitempty"`
	Quantity       *float64         `bson:"quantity,omitempty"`
	AvgPrice       *bson.Decimal128 `bson:"avg_price,omitempty"`
	CurrentPrice   *bson.Decimal128 `bson:"current_price,omitempty"`
	InvestedAmount *bson.Decimal128 `bson:"invested_amount,omitempty"`
	CurrentBalance bson.Decimal128  `bson:"current_balance"`

	Active    *bool     `bson:"active,omitempty"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

func newInvestmentAssetDoc(a *domaininvestment.Asset) (investmentAssetDoc, error) {
	balance, err := toDecimal128(a.CurrentBalance)
	if err != nil {
		return investmentAssetDoc{}, err
	}
	avg, err := toDecimal128(a.AvgPrice)
	if err != nil {
		return investmentAssetDoc{}, err
	}
	current, err := toDecimal128(a.CurrentPrice)
	if err != nil {
		return investmentAssetDoc{}, err
	}
	invested, err := toDecimal128(a.InvestedAmount)
	if err != nil {
		return investmentAssetDoc{}, err
	}
	active := a.Active
	ticker := a.Ticker
	class := a.Class
	quantity := a.Quantity
	return investmentAssetDoc{
		ID:             a.ID.String(),
		UserID:         a.UserID.String(),
		Name:           a.Name,
		Ticker:         &ticker,
		Class:          &class,
		Quantity:       &quantity,
		AvgPrice:       &avg,
		CurrentPrice:   &current,
		InvestedAmount: &invested,
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
	avg, err := decimalPtrToMoney(d.AvgPrice)
	if err != nil {
		return nil, err
	}
	current, err := decimalPtrToMoney(d.CurrentPrice)
	if err != nil {
		return nil, err
	}
	invested, err := decimalPtrToMoney(d.InvestedAmount)
	if err != nil {
		return nil, err
	}

	if d.InvestedAmount == nil {
		invested = balance
	}

	active := d.Active == nil || *d.Active
	ticker := ""
	if d.Ticker != nil {
		ticker = *d.Ticker
	}
	class := "outro"
	if d.Class != nil && *d.Class != "" {
		class = *d.Class
	}
	var quantity float64
	if d.Quantity != nil {
		quantity = *d.Quantity
	}
	return &domaininvestment.Asset{
		ID:             id,
		UserID:         userID,
		Name:           d.Name,
		Ticker:         ticker,
		Class:          class,
		Quantity:       quantity,
		AvgPrice:       avg,
		CurrentPrice:   current,
		InvestedAmount: invested,
		CurrentBalance: balance,
		Active:         active,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}, nil
}

func decimalPtrToMoney(d *bson.Decimal128) (shared.Money, error) {
	if d == nil {
		return shared.Zero, nil
	}
	return fromDecimal128(*d)
}

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

func (r *InvestmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return fmt.Errorf("erro ao remover ativo: %w", err)
	}
	return checkDeletedCount(res)
}

type ProventoRepository struct {
	collection *mongo.Collection
}

func NewProventoRepository(db *mongo.Database) *ProventoRepository {
	return &ProventoRepository{collection: db.Collection(investmentProventosCollection)}
}

type proventoDoc struct {
	ID        string          `bson:"_id"`
	UserID    string          `bson:"user_id"`
	AssetID   string          `bson:"asset_id"`
	Amount    bson.Decimal128 `bson:"amount"`
	Date      time.Time       `bson:"date"`
	CreatedAt time.Time       `bson:"created_at"`
}

func newProventoDoc(p *domaininvestment.Provento) (proventoDoc, error) {
	amount, err := toDecimal128(p.Amount)
	if err != nil {
		return proventoDoc{}, err
	}
	return proventoDoc{
		ID:        p.ID.String(),
		UserID:    p.UserID.String(),
		AssetID:   p.AssetID.String(),
		Amount:    amount,
		Date:      p.Date,
		CreatedAt: p.CreatedAt,
	}, nil
}

func (d proventoDoc) toDomain() (*domaininvestment.Provento, error) {
	id, err := uuid.Parse(d.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de provento: %w", err)
	}
	userID, err := uuid.Parse(d.UserID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de usuário do provento: %w", err)
	}
	assetID, err := uuid.Parse(d.AssetID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de ativo do provento: %w", err)
	}
	amount, err := fromDecimal128(d.Amount)
	if err != nil {
		return nil, err
	}
	return &domaininvestment.Provento{
		ID:        id,
		UserID:    userID,
		AssetID:   assetID,
		Amount:    amount,
		Date:      d.Date,
		CreatedAt: d.CreatedAt,
	}, nil
}

func (r *ProventoRepository) Create(ctx context.Context, p *domaininvestment.Provento) error {
	doc, err := newProventoDoc(p)
	if err != nil {
		return err
	}
	if _, err := r.collection.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("erro ao inserir provento: %w", err)
	}
	return nil
}

func (r *ProventoRepository) FindByID(ctx context.Context, id uuid.UUID) (*domaininvestment.Provento, error) {
	var doc proventoDoc
	if err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("erro ao buscar provento: %w", err)
	}
	return doc.toDomain()
}

func (r *ProventoRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domaininvestment.Provento, error) {
	opts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID.String()}, opts)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar proventos: %w", err)
	}

	var docs []proventoDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("erro ao decodificar proventos: %w", err)
	}

	proventos := make([]*domaininvestment.Provento, 0, len(docs))
	for _, doc := range docs {
		p, err := doc.toDomain()
		if err != nil {
			return nil, err
		}
		proventos = append(proventos, p)
	}
	return proventos, nil
}

func (r *ProventoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return fmt.Errorf("erro ao remover provento: %w", err)
	}
	return checkDeletedCount(res)
}

func (r *ProventoRepository) DeleteByAsset(ctx context.Context, assetID uuid.UUID) error {
	if _, err := r.collection.DeleteMany(ctx, bson.M{"asset_id": assetID.String()}); err != nil {
		return fmt.Errorf("erro ao remover proventos do ativo: %w", err)
	}
	return nil
}
