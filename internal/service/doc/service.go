package doc

//go:generate mockgen -destination=service_mocks.go -source=service.go -package=doc

import (
	"context"
	"io"
	"slices"

	doccmd "github.com/Mabarik667f/fsserver/internal/command/doc"
	"github.com/Mabarik667f/fsserver/internal/errs"
	"github.com/Mabarik667f/fsserver/internal/model"
	"github.com/Mabarik667f/fsserver/internal/model/query"
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
	DeleteByID(ctx context.Context, docID uuid.UUID) error
	GetByID(ctx context.Context, docID uuid.UUID) (query.DocReadModel, error)
	Get(
		ctx context.Context,
		cmd doccmd.GetDocumentsListCmd,
		userID uuid.UUID,
	) ([]query.DocReadModel, error)
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

func (s *service) Get(
	cmd doccmd.GetDocumentsListCmd,
	user model.User,
) ([]query.DocReadModel, error) {
	ctx := context.Background()

	if cmd.Login == nil {
		cmd.Login = &user.Login
	}

	docs, err := s.repo.Get(ctx, cmd, user.ID)
	if err != nil {
		return []query.DocReadModel{}, err
	}

	return docs, nil
}

func (s *service) GetByID(
	id uuid.UUID,
	user model.User,
	metaOnly bool,
) (query.FullDocReadModel, error) {
	ctx := context.Background()
	doc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return query.FullDocReadModel{}, err
	}

	if !doc.IsPublic &&
		user.ID != doc.OwnerID &&
		!slices.Contains(doc.Grants, user.Login) {
		return query.FullDocReadModel{}, errs.ErrDocPermissionDenied
	}

	res := query.FullDocReadModel{
		ID:        doc.ID,
		OwnerID:   doc.OwnerID,
		Name:      doc.Name,
		IsFile:    doc.IsFile,
		IsPublic:  doc.IsPublic,
		MimeType:  doc.MimeType,
		JSONData:  doc.JSONData,
		FilePath:  doc.FilePath,
		CreatedAt: doc.CreatedAt,
		Grants:    doc.Grants,
		File:      nil,
	}

	if !metaOnly && doc.IsFile {
		file, err := s.storage.Read(doc.ID.String(), doc.Name)
		if err != nil {
			return query.FullDocReadModel{}, err
		}
		res.File = file
	}

	return res, nil
}

func (s *service) DeleteByID(id, userID uuid.UUID) error {
	ctx := context.Background()
	doc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if userID != doc.OwnerID {
		return errs.ErrDocPermissionDenied
	}

	if err := s.repo.DeleteByID(ctx, doc.ID); err != nil {
		return err
	}

	return s.storage.Delete(doc.ID.String())
}
