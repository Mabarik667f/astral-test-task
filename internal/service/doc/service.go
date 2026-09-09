package doc

//go:generate mockgen -destination=service_mocks.go -source=service.go -package=doc

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"time"

	doccmd "github.com/Mabarik667f/fsserver/internal/command/doc"
	"github.com/Mabarik667f/fsserver/internal/errs"
	"github.com/Mabarik667f/fsserver/internal/model"
	"github.com/Mabarik667f/fsserver/internal/model/query"
	"github.com/google/uuid"
)

type Cache interface {
	Set(k string, v any, ttl time.Duration)
	Get(k string) (any, bool)
	Delete(k string)
	DeleteByPrefix(prefix string)
}

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
	cache    Cache
}

func NewService(
	userRepo UserRepository,
	repo Repository,
	storage FileStorage,
	cache Cache,
) *service {
	return &service{repo: repo, userRepo: userRepo, storage: storage, cache: cache}
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

	if cmd.IsFile {
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

	s.cache.DeleteByPrefix("docs:list:")

	return nil
}

func (s *service) Get(
	cmd doccmd.GetDocumentsListCmd,
	user model.User,
) ([]query.DocReadModel, error) {
	ctx := context.Background()
	slog.Info("user", "login", user.Login, "id", user.ID)

	key := listCacheKey(cmd, user.ID)
	if value, ok := s.cache.Get(key); ok {
		docs, ok := value.([]query.DocReadModel)
		if ok {
			slog.Info("cache", "docs", len(docs))
			return docs, nil
		}
	}

	docs, err := s.repo.Get(ctx, cmd, user.ID)
	if err != nil {
		return []query.DocReadModel{}, err
	}

	s.cache.Set(key, docs, 5*time.Minute)

	slog.Info("not cached", "docs", len(docs))
	return docs, nil
}

func (s *service) GetByID(
	id uuid.UUID,
	user model.User,
	metaOnly bool,
) (query.FullDocReadModel, error) {
	ctx := context.Background()

	key := docCacheKey(id, user.ID)
	value, ok := s.cache.Get(key)
	if ok {
		doc, ok := value.(query.FullDocReadModel)
		if ok {
			file, err := s.readFile(metaOnly, doc)
			if err != nil {
				return query.FullDocReadModel{}, err
			}
			doc.File = file
			return doc, nil
		}
	}

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

	s.cache.Set(key, res, 5*time.Minute)

	file, err := s.readFile(metaOnly, res)
	if err != nil {
		return query.FullDocReadModel{}, err
	}
	res.File = file

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

	s.cache.Delete(docCacheKey(id, userID))
	s.cache.DeleteByPrefix("docs:list:")

	if err := s.storage.Delete(doc.ID.String()); err != nil {
		return err
	}

	return nil
}

func (s *service) readFile(metaOnly bool, doc query.FullDocReadModel) (io.ReadCloser, error) {
	if !metaOnly && doc.IsFile {
		file, err := s.storage.Read(doc.ID.String(), doc.Name)
		if err != nil {
			return nil, err
		}
		return file, nil
	}
	return nil, nil
}

func listCacheKey(cmd doccmd.GetDocumentsListCmd, userID uuid.UUID) string {
	login := ""
	if cmd.Login != nil {
		login = *cmd.Login
	}

	return fmt.Sprintf(
		"docs:list:%s:%s:%s:%s:%d",
		userID,
		login,
		cmd.Key,
		cmd.Value,
		cmd.Limit,
	)
}

func docCacheKey(id, userID uuid.UUID) string {
	return fmt.Sprintf("docs:%s:%s", userID, id)
}
