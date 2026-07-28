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

	domainwishlist "rinofinance-api/internal/domain/wishlist"
	"rinofinance-api/internal/domain/shared"
)

// ---- Sections ----

// WishlistSectionRepository implements domain/wishlist.SectionRepository.
type WishlistSectionRepository struct {
	collection *mongo.Collection
}

// NewWishlistSectionRepository wires the repository to a database handle.
func NewWishlistSectionRepository(db *mongo.Database) *WishlistSectionRepository {
	return &WishlistSectionRepository{collection: db.Collection(wishlistSectionsCollection)}
}

type wishlistSectionDoc struct {
	ID        string    `bson:"_id"`
	UserID    string    `bson:"user_id"`
	Name      string    `bson:"name"`
	Position  int       `bson:"position"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

func newWishlistSectionDoc(s *domainwishlist.Section) wishlistSectionDoc {
	return wishlistSectionDoc{
		ID:        s.ID.String(),
		UserID:    s.UserID.String(),
		Name:      s.Name,
		Position:  s.Position,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func (d wishlistSectionDoc) toDomain() (*domainwishlist.Section, error) {
	id, err := uuid.Parse(d.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de seção: %w", err)
	}
	userID, err := uuid.Parse(d.UserID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de usuário da seção: %w", err)
	}
	return &domainwishlist.Section{
		ID:        id,
		UserID:    userID,
		Name:      d.Name,
		Position:  d.Position,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}, nil
}

func (r *WishlistSectionRepository) Create(ctx context.Context, s *domainwishlist.Section) error {
	if _, err := r.collection.InsertOne(ctx, newWishlistSectionDoc(s)); err != nil {
		return fmt.Errorf("erro ao inserir seção: %w", err)
	}
	return nil
}

func (r *WishlistSectionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainwishlist.Section, error) {
	var doc wishlistSectionDoc
	if err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("erro ao buscar seção: %w", err)
	}
	return doc.toDomain()
}

func (r *WishlistSectionRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domainwishlist.Section, error) {
	opts := options.Find().SetSort(bson.D{{Key: "position", Value: 1}, {Key: "created_at", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID.String()}, opts)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar seções: %w", err)
	}
	var docs []wishlistSectionDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("erro ao decodificar seções: %w", err)
	}
	sections := make([]*domainwishlist.Section, 0, len(docs))
	for _, doc := range docs {
		s, err := doc.toDomain()
		if err != nil {
			return nil, err
		}
		sections = append(sections, s)
	}
	return sections, nil
}

func (r *WishlistSectionRepository) Update(ctx context.Context, s *domainwishlist.Section) error {
	res, err := r.collection.ReplaceOne(ctx, bson.M{"_id": s.ID.String()}, newWishlistSectionDoc(s))
	if err != nil {
		return fmt.Errorf("erro ao atualizar seção: %w", err)
	}
	return checkMatchedCount(res)
}

func (r *WishlistSectionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return fmt.Errorf("erro ao remover seção: %w", err)
	}
	return checkDeletedCount(res)
}

// ---- Items ----

// WishlistItemRepository implements domain/wishlist.ItemRepository.
type WishlistItemRepository struct {
	collection *mongo.Collection
}

// NewWishlistItemRepository wires the repository to a database handle.
func NewWishlistItemRepository(db *mongo.Database) *WishlistItemRepository {
	return &WishlistItemRepository{collection: db.Collection(wishlistItemsCollection)}
}

type wishlistItemDoc struct {
	ID        string          `bson:"_id"`
	UserID    string          `bson:"user_id"`
	SectionID *string         `bson:"section_id,omitempty"`
	Name      string          `bson:"name"`
	URL       string          `bson:"url,omitempty"`
	Price     bson.Decimal128 `bson:"price"`
	ImageURL  string          `bson:"image_url,omitempty"`
	Position  int             `bson:"position"`
	CreatedAt time.Time       `bson:"created_at"`
	UpdatedAt time.Time       `bson:"updated_at"`
}

func newWishlistItemDoc(i *domainwishlist.Item) (wishlistItemDoc, error) {
	price, err := toDecimal128(i.Price)
	if err != nil {
		return wishlistItemDoc{}, err
	}
	return wishlistItemDoc{
		ID:        i.ID.String(),
		UserID:    i.UserID.String(),
		SectionID: uuidPtrToString(i.SectionID),
		Name:      i.Name,
		URL:       i.URL,
		Price:     price,
		ImageURL:  i.ImageURL,
		Position:  i.Position,
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
	}, nil
}

func (d wishlistItemDoc) toDomain() (*domainwishlist.Item, error) {
	id, err := uuid.Parse(d.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de item: %w", err)
	}
	userID, err := uuid.Parse(d.UserID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de usuário do item: %w", err)
	}
	sectionID, err := stringPtrToUUID(d.SectionID, "ID de seção do item")
	if err != nil {
		return nil, err
	}
	price, err := fromDecimal128(d.Price)
	if err != nil {
		return nil, err
	}
	return &domainwishlist.Item{
		ID:        id,
		UserID:    userID,
		SectionID: sectionID,
		Name:      d.Name,
		URL:       d.URL,
		Price:     price,
		ImageURL:  d.ImageURL,
		Position:  d.Position,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}, nil
}

func (r *WishlistItemRepository) Create(ctx context.Context, i *domainwishlist.Item) error {
	doc, err := newWishlistItemDoc(i)
	if err != nil {
		return err
	}
	if _, err := r.collection.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("erro ao inserir item: %w", err)
	}
	return nil
}

func (r *WishlistItemRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainwishlist.Item, error) {
	var doc wishlistItemDoc
	if err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("erro ao buscar item: %w", err)
	}
	return doc.toDomain()
}

func (r *WishlistItemRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domainwishlist.Item, error) {
	opts := options.Find().SetSort(bson.D{{Key: "position", Value: 1}, {Key: "created_at", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID.String()}, opts)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar itens: %w", err)
	}
	var docs []wishlistItemDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("erro ao decodificar itens: %w", err)
	}
	items := make([]*domainwishlist.Item, 0, len(docs))
	for _, doc := range docs {
		i, err := doc.toDomain()
		if err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}

func (r *WishlistItemRepository) Update(ctx context.Context, i *domainwishlist.Item) error {
	doc, err := newWishlistItemDoc(i)
	if err != nil {
		return err
	}
	res, err := r.collection.ReplaceOne(ctx, bson.M{"_id": i.ID.String()}, doc)
	if err != nil {
		return fmt.Errorf("erro ao atualizar item: %w", err)
	}
	return checkMatchedCount(res)
}

func (r *WishlistItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return fmt.Errorf("erro ao remover item: %w", err)
	}
	return checkDeletedCount(res)
}
