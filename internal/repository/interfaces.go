package repository

import (
	"context"

	"github.com/Mabarik667f/fsserver/internal/model"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user model.User) (model.User, error)
	GetByLoginWithPasswordHash(ctx context.Context, login string) (model.User, error)
	GetUsersByLogins(ctx context.Context, logins []string) ([]model.User, error)
}

type DocRepository interface {
	Create(ctx context.Context, doc model.Doc) (model.Doc, error)
	CreateGrants(ctx context.Context, docID uuid.UUID, userIDs uuid.UUIDs) error
}
