package repository

import (
	"context"
	"errors"

	"github.com/darkin-kat/user/internal/domain"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) (domain.User, error)
	GetByID(ctx context.Context, id string) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Update(ctx context.Context, user *domain.User) (domain.User, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit int, offset int) ([]domain.User, error)
}