package doc

//go:generate mockgen -destination=service_mocks.go -source=service.go -package=doc

import (
	"context"
	"io"

	doccmd "github.com/Mabarik667f/fsserver/internal/command/doc"
	"github.com/Mabarik667f/fsserver/internal/model"
	"github.com/google/uuid"
)

type FileStorage interface {
	Save(fileID string, name string, data io.Reader) (string, error)
	Read(fileID, name string) (io.ReadCloser, error)
	Delete(fileID string) error
}

type UserRepository interface {
	GetUsersByLogins(ctx context.Context, logins []string) ([]model.User, error)
}

type Repository interface {
	Create(ctx context.Context, doc model.Doc) (model.Doc, error)
	CreateGrants(ctx context.Context, docID uuid.UUID, userIDs uuid.UUIDs) error
}

type service struct {
	userRepo UserRepository
	repo     Repository
	storage  FileStorage
}

func NewService(repo Repository, userRepo UserRepository, storage FileStorage) *service {
	return &service{repo: repo, userRepo: userRepo, storage: storage}
}

func (s *service) Upload(cmd doccmd.CreateDocumentCmd) error                  { return nil }
func (s *service) Get(userID uuid.UUID, cmd doccmd.GetDocumentsListCmd) error { return nil }
func (s *service) GetByID(id, userID uuid.UUID, metaOnly bool) error          { return nil }
func (s *service) DeleteByID(id, userID uuid.UUID) (bool, error)              { return false, nil }
