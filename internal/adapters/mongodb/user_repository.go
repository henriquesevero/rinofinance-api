package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"rinofinance-api/internal/domain/shared"
	domainuser "rinofinance-api/internal/domain/user"
)

// UserRepository implements domain/user.Repository against MongoDB.
type UserRepository struct {
	collection *mongo.Collection
}

// NewUserRepository wires a UserRepository to a database handle.
func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{collection: db.Collection(usersCollection)}
}

type userDoc struct {
	ID           string    `bson:"_id"`
	Name         string    `bson:"name"`
	Email        string    `bson:"email"`
	PasswordHash string    `bson:"password_hash"`
	AvatarURL    string    `bson:"avatar_url,omitempty"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
}

func newUserDoc(u *domainuser.User) userDoc {
	return userDoc{
		ID:           u.ID.String(),
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		AvatarURL:    u.AvatarURL,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

func (d userDoc) toDomain() (*domainuser.User, error) {
	id, err := uuid.Parse(d.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID de usuário: %w", err)
	}
	return &domainuser.User{
		ID:           id,
		Name:         d.Name,
		Email:        d.Email,
		PasswordHash: d.PasswordHash,
		AvatarURL:    d.AvatarURL,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}, nil
}

// Create inserts a new user document. A duplicate email is reported as
// domainuser.ErrEmailAlreadyInUse via the unique index created in
// EnsureIndexes.
func (r *UserRepository) Create(ctx context.Context, u *domainuser.User) error {
	if _, err := r.collection.InsertOne(ctx, newUserDoc(u)); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domainuser.ErrEmailAlreadyInUse
		}
		return fmt.Errorf("erro ao inserir usuário: %w", err)
	}
	return nil
}

// FindByID fetches a user by ID.
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainuser.User, error) {
	var doc userDoc
	if err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("erro ao buscar usuário: %w", err)
	}
	return doc.toDomain()
}

// FindByEmail fetches a user by email (used for login and duplicate
// registration checks).
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domainuser.User, error) {
	var doc userDoc
	if err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("erro ao buscar usuário por email: %w", err)
	}
	return doc.toDomain()
}

// Update persists changes to name, email, avatar and password hash.
func (r *UserRepository) Update(ctx context.Context, u *domainuser.User) error {
	res, err := r.collection.ReplaceOne(ctx, bson.M{"_id": u.ID.String()}, newUserDoc(u))
	if err != nil {
		return fmt.Errorf("erro ao atualizar usuário: %w", err)
	}
	return checkMatchedCount(res)
}

// Delete permanently removes a user document. Callers (see
// application/profile.DeleteAccountUseCase) are responsible for first
// deleting every other aggregate the user owns.
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return fmt.Errorf("erro ao remover usuário: %w", err)
	}
	return checkDeletedCount(res)
}
