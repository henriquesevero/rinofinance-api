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

	domaincategory "rinofinance-api/internal/domain/category"
	"rinofinance-api/internal/domain/shared"
)

type CategoryRepository struct {
	collection *mongo.Collection
}

func NewCategoryRepository(db *mongo.Database) *CategoryRepository {
	return &CategoryRepository{collection: db.Collection(categoriesCollection)}
}

type categoryDoc struct {
	ID        string    `bson:"_id"`
	UserID    string    `bson:"user_id"`
	Name      string    `bson:"name"`
	Color     string    `bson:"color,omitempty"`
	Icon      string    `bson:"icon,omitempty"`
	Position  int       `bson:"position"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

func newCategoryDoc(c *domaincategory.Category) categoryDoc {
	return categoryDoc{
		ID:        c.ID.String(),
		UserID:    c.UserID.String(),
		Name:      c.Name,
		Color:     c.Color,
		Icon:      c.Icon,
		Position:  c.Position,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func (d categoryDoc) toDomain() (*domaincategory.Category, error) {
	id, err := uuid.Parse(d.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de categoria: %w", err)
	}
	userID, err := uuid.Parse(d.UserID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de usuário da categoria: %w", err)
	}
	return &domaincategory.Category{
		ID:        id,
		UserID:    userID,
		Name:      d.Name,
		Color:     d.Color,
		Icon:      d.Icon,
		Position:  d.Position,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}, nil
}

func (r *CategoryRepository) Create(ctx context.Context, c *domaincategory.Category) error {
	if _, err := r.collection.InsertOne(ctx, newCategoryDoc(c)); err != nil {
		return fmt.Errorf("erro ao inserir categoria: %w", err)
	}
	return nil
}

func (r *CategoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*domaincategory.Category, error) {
	var doc categoryDoc
	if err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("erro ao buscar categoria: %w", err)
	}
	return doc.toDomain()
}

func (r *CategoryRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domaincategory.Category, error) {
	opts := options.Find().SetSort(bson.D{{Key: "position", Value: 1}, {Key: "name", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID.String()}, opts)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar categorias: %w", err)
	}

	var docs []categoryDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("erro ao decodificar categorias: %w", err)
	}

	categories := make([]*domaincategory.Category, 0, len(docs))
	for _, doc := range docs {
		c, err := doc.toDomain()
		if err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func (r *CategoryRepository) Update(ctx context.Context, c *domaincategory.Category) error {
	res, err := r.collection.ReplaceOne(ctx, bson.M{"_id": c.ID.String()}, newCategoryDoc(c))
	if err != nil {
		return fmt.Errorf("erro ao atualizar categoria: %w", err)
	}
	return checkMatchedCount(res)
}

func (r *CategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return fmt.Errorf("erro ao remover categoria: %w", err)
	}
	return checkDeletedCount(res)
}
