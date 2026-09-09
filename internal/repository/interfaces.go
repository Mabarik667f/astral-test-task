package repository

import (
	"context"

	doccmd "github.com/Mabarik667f/fsserver/internal/command/doc"
	"github.com/Mabarik667f/fsserver/internal/model"
	"github.com/Mabarik667f/fsserver/internal/model/query"
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
	DeleteByID(ctx context.Context, docID uuid.UUID) error
	GetByID(ctx context.Context, docID uuid.UUID) (query.DocReadModel, error)
	Get(
		ctx context.Context,
		cmd doccmd.GetDocumentsListCmd,
		userID uuid.UUID,
	) ([]query.DocReadModel, error)
}
