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

	domaincard "rinofinance-api/internal/domain/card"
	"rinofinance-api/internal/domain/shared"
)

// InstallmentPurchaseRepository implements
// domain/card.InstallmentPurchaseRepository against MongoDB.
type InstallmentPurchaseRepository struct {
	collection *mongo.Collection
}

// NewInstallmentPurchaseRepository wires an InstallmentPurchaseRepository
// to a database handle.
func NewInstallmentPurchaseRepository(db *mongo.Database) *InstallmentPurchaseRepository {
	return &InstallmentPurchaseRepository{collection: db.Collection(installmentPurchasesCollection)}
}

type installmentPurchaseDoc struct {
	ID                   string          `bson:"_id"`
	CardID               string          `bson:"card_id"`
	Name                 string          `bson:"name"`
	InstallmentAmount    bson.Decimal128 `bson:"installment_amount"`
	TotalInstallments    int             `bson:"total_installments"`
	FirstInstallmentDate time.Time       `bson:"first_installment_date"`
	Domain               string          `bson:"domain,omitempty"`
	Flagged              bool            `bson:"flagged,omitempty"`
	CategoryID           *string         `bson:"category_id,omitempty"`
	Position             int             `bson:"position"`
	CreatedAt            time.Time       `bson:"created_at"`
	UpdatedAt            time.Time       `bson:"updated_at"`
}

func newInstallmentPurchaseDoc(p *domaincard.InstallmentPurchase) (installmentPurchaseDoc, error) {
	amount, err := toDecimal128(p.InstallmentAmount)
	if err != nil {
		return installmentPurchaseDoc{}, err
	}
	return installmentPurchaseDoc{
		ID:                   p.ID.String(),
		CardID:               p.CardID.String(),
		Name:                 p.Name,
		InstallmentAmount:    amount,
		TotalInstallments:    p.TotalInstallments,
		FirstInstallmentDate: p.FirstInstallmentDate,
		Domain:               p.Domain,
		Flagged:              p.Flagged,
		CategoryID:           uuidPtrToString(p.CategoryID),
		Position:             p.Position,
		CreatedAt:            p.CreatedAt,
		UpdatedAt:            p.UpdatedAt,
	}, nil
}

func (d installmentPurchaseDoc) toDomain() (*domaincard.InstallmentPurchase, error) {
	id, err := uuid.Parse(d.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de compra parcelada: %w", err)
	}
	cardID, err := uuid.Parse(d.CardID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de cartão da compra parcelada: %w", err)
	}
	amount, err := fromDecimal128(d.InstallmentAmount)
	if err != nil {
		return nil, err
	}
	categoryID, err := stringPtrToUUID(d.CategoryID, "ID de categoria da compra")
	if err != nil {
		return nil, err
	}
	return &domaincard.InstallmentPurchase{
		ID:                   id,
		CardID:               cardID,
		Name:                 d.Name,
		InstallmentAmount:    amount,
		TotalInstallments:    d.TotalInstallments,
		FirstInstallmentDate: d.FirstInstallmentDate,
		Domain:               d.Domain,
		Flagged:              d.Flagged,
		CategoryID:           categoryID,
		Position:             d.Position,
		CreatedAt:            d.CreatedAt,
		UpdatedAt:            d.UpdatedAt,
	}, nil
}

// Create inserts a new installment purchase document.
func (r *InstallmentPurchaseRepository) Create(ctx context.Context, p *domaincard.InstallmentPurchase) error {
	doc, err := newInstallmentPurchaseDoc(p)
	if err != nil {
		return err
	}
	if _, err := r.collection.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("erro ao inserir compra parcelada: %w", err)
	}
	return nil
}

// FindByID fetches an installment purchase by ID.
func (r *InstallmentPurchaseRepository) FindByID(ctx context.Context, id uuid.UUID) (*domaincard.InstallmentPurchase, error) {
	var doc installmentPurchaseDoc
	if err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("erro ao buscar compra parcelada: %w", err)
	}
	return doc.toDomain()
}

// ListByCard fetches every installment purchase belonging to a card.
func (r *InstallmentPurchaseRepository) ListByCard(ctx context.Context, cardID uuid.UUID) ([]*domaincard.InstallmentPurchase, error) {
	opts := options.Find().SetSort(bson.D{{Key: "position", Value: 1}, {Key: "created_at", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{"card_id": cardID.String()}, opts)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar compras parceladas: %w", err)
	}

	var docs []installmentPurchaseDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("erro ao decodificar compras parceladas: %w", err)
	}

	purchases := make([]*domaincard.InstallmentPurchase, 0, len(docs))
	for _, doc := range docs {
		p, err := doc.toDomain()
		if err != nil {
			return nil, err
		}
		purchases = append(purchases, p)
	}
	return purchases, nil
}

// Update persists changes to an installment purchase's fields.
func (r *InstallmentPurchaseRepository) Update(ctx context.Context, p *domaincard.InstallmentPurchase) error {
	doc, err := newInstallmentPurchaseDoc(p)
	if err != nil {
		return err
	}
	res, err := r.collection.ReplaceOne(ctx, bson.M{"_id": p.ID.String()}, doc)
	if err != nil {
		return fmt.Errorf("erro ao atualizar compra parcelada: %w", err)
	}
	return checkMatchedCount(res)
}

// Delete permanently removes an installment purchase document.
func (r *InstallmentPurchaseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return fmt.Errorf("erro ao remover compra parcelada: %w", err)
	}
	return checkDeletedCount(res)
}
