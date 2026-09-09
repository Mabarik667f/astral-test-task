package repository

import (
	"context"

	"github.com/Mabarik667f/fsserver/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user model.User) (model.User, error)
	GetByLoginWithPasswordHash(ctx context.Context, login string) (model.User, error)
}
