package mongo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/darkin-kat/user/internal/domain"
	"github.com/darkin-kat/user/internal/repository"
)

// Repository представляет реализацию интерфейса UserRepository для работы с MongoDB.
type Repository struct {
	collection *mongo.Collection
}

// Убеждаемся, что Repository реализует интерфейс UserRepository.
// Убеждаемся, что Repository реализует интерфейс UserRepository.
var _ repository.UserRepository = (*Repository)(nil)

// NewMongoRepository создает новый экземпляр Repository для работы с MongoDB.
func NewMongoRepository(db *mongo.Database) *Repository {
	return &Repository{
		collection: db.Collection("users"),
	}
}

func (r *Repository) Create(ctx context.Context, user *domain.User) (domain.User, error) {
	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.User{}, repository.ErrEmailAlreadyExists
		}
		return domain.User{}, err
	}
	return *user, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (domain.User, error) {
	var user domain.User

	filter := bson.M{"id": id}
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.User{}, repository.ErrUserNotFound
		}
		return domain.User{}, err
	}
	return user, nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	var user domain.User

	filter := bson.M{"email": email}
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.User{}, repository.ErrUserNotFound
		}
		return domain.User{}, err
	}
	return user, nil
}

func (r *Repository) Update(ctx context.Context, user *domain.User) (domain.User, error) {
	filter := bson.M{"_id": user.ID}
	update := bson.M{
		"$set": bson.M{
			"email":         user.Email,
			"first_name":    user.FirstName,
			"last_name":     user.LastName,
			"password_hash": user.PasswordHash,
			"updated_at":    user.UpdatedAt,
		},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedUser domain.User
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedUser)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.User{}, repository.ErrUserNotFound
		}
		return domain.User{}, err
	}
	return updatedUser, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	filter := bson.M{"_id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return repository.ErrUserNotFound
	}
	return nil
}

func (r *Repository) List(ctx context.Context, limit uint, offset uint) ([]domain.User, error) {
	findOptions := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []domain.User
	for cursor.Next(ctx) {
		var user domain.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *Repository) Count(ctx context.Context) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, err
	}
	return count, nil
}
