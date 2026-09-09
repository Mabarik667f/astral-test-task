package doc

//go:generate mockgen -destination=service_mocks.go -source=service.go -package=doc

import (
	"context"
	"io"
	"log/slog"

	doccmd "github.com/Mabarik667f/fsserver/internal/command/doc"
	"github.com/Mabarik667f/fsserver/internal/errs"
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

func NewService(userRepo UserRepository, repo Repository, storage FileStorage) *service {
	return &service{repo: repo, userRepo: userRepo, storage: storage}
}

func NewEmpty() *service {
	return &service{}
}

func (s *service) Upload(cmd doccmd.CreateDocumentCmd) error {
	if cmd.IsFile && cmd.File == nil {
		slog.Info("message", "1", "1")
		return errs.ErrDocBusiness
	}

	doc, err := model.NewDoc(
		cmd.OwnerID,
		cmd.Name,
		cmd.IsFile,
		cmd.IsPublic,
		cmd.MimeType,
		cmd.JSONData,
		"",
	)
	if err != nil {
		slog.Info("message", "2", "2")
		return errs.ErrDocBusiness
	}

	if cmd.File != nil {
		path, err := s.storage.Save(doc.ID.String(), doc.Name, cmd.File)
		if err != nil {
			return errs.ErrSaveFileToStorage
		}
		doc.FilePath = path
	}

	ctx := context.Background()

	users, err := s.userRepo.GetUsersByLogins(ctx, cmd.Grant)
	if err != nil {
		return err
	}

	if len(users) != len(cmd.Grant) {
		return errs.ErrDocBusiness
	}

	if _, err = s.repo.Create(ctx, *doc); err != nil {
		_ = s.storage.Delete(doc.ID.String())
		return err
	}

	userIDs := make(uuid.UUIDs, len(users))
	for i := range users {
		userIDs[i] = users[i].ID
	}

	if err := s.repo.CreateGrants(ctx, doc.ID, userIDs); err != nil {
		_ = s.storage.Delete(doc.ID.String())
		return err
	}

	return nil
}

func (s *service) Get(userID uuid.UUID, cmd doccmd.GetDocumentsListCmd) error { return nil }
func (s *service) GetByID(id, userID uuid.UUID, metaOnly bool) error          { return nil }
func (s *service) DeleteByID(id, userID uuid.UUID) (bool, error)              { return false, nil }
