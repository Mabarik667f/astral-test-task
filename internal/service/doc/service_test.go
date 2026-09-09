package doc

import (
	"errors"
	"io"
	"strings"
	"testing"

	doccmd "github.com/Mabarik667f/fsserver/internal/command/doc"
	"github.com/Mabarik667f/fsserver/internal/errs"
	"github.com/Mabarik667f/fsserver/internal/model"
	"github.com/Mabarik667f/fsserver/internal/model/query"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

func TestService_Upload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockUserRepo := NewMockUserRepository(ctrl)
		mockStorage := NewMockFileStorage(ctrl)

		file := strings.NewReader("test file")

		cmd := doccmd.CreateDocumentCmd{
			OwnerID:  uuid.New(),
			Name:     "test.txt",
			IsFile:   true,
			IsPublic: false,
			MimeType: "text/plain",
			JSONData: map[string]any{
				"foo": "bar",
			},
			File:  file,
			Grant: []string{"user1"},
		}

		user := model.User{
			ID:    uuid.New(),
			Login: "user1",
		}

		mockStorage.
			EXPECT().
			Save(gomock.Any(), cmd.Name, cmd.File).
			Return("/storage/test.txt", nil)

		mockUserRepo.
			EXPECT().
			GetUsersByLogins(gomock.Any(), cmd.Grant).
			Return([]model.User{user}, nil)

		mockRepo.
			EXPECT().
			Create(gomock.Any(), gomock.AssignableToTypeOf(model.Doc{})).
			Return(model.Doc{}, nil)

		mockRepo.
			EXPECT().
			CreateGrants(
				gomock.Any(),
				gomock.Any(),
				uuid.UUIDs{user.ID},
			).
			Return(nil)

		service := NewService(
			mockUserRepo,
			mockRepo,
			mockStorage,
		)

		err := service.Upload(cmd)

		require.NoError(t, err)
	})

	t.Run("file is required", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockUserRepo := NewMockUserRepository(ctrl)
		mockStorage := NewMockFileStorage(ctrl)

		cmd := doccmd.CreateDocumentCmd{
			OwnerID:  uuid.New(),
			Name:     "test.txt",
			IsFile:   true,
			MimeType: "text/plain",
		}

		service := NewService(
			mockUserRepo,
			mockRepo,
			mockStorage,
		)

		err := service.Upload(cmd)

		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrDocBusiness)
	})

	t.Run("storage error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockUserRepo := NewMockUserRepository(ctrl)
		mockStorage := NewMockFileStorage(ctrl)

		file := strings.NewReader("test file")

		cmd := doccmd.CreateDocumentCmd{
			OwnerID:  uuid.New(),
			Name:     "test.txt",
			IsFile:   true,
			MimeType: "text/plain",
			File:     file,
		}

		mockStorage.
			EXPECT().
			Save(gomock.Any(), cmd.Name, cmd.File).
			Return("", errors.New("storage error"))

		service := NewService(
			mockUserRepo,
			mockRepo,
			mockStorage,
		)

		err := service.Upload(cmd)

		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrSaveFileToStorage)
	})

	t.Run("repository error deletes file", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockUserRepo := NewMockUserRepository(ctrl)
		mockStorage := NewMockFileStorage(ctrl)

		file := strings.NewReader("test file")

		cmd := doccmd.CreateDocumentCmd{
			OwnerID:  uuid.New(),
			Name:     "test.txt",
			IsFile:   true,
			MimeType: "text/plain",
			File:     file,
			Grant:    []string{"user1"},
		}

		user := model.User{
			ID:    uuid.New(),
			Login: "user1",
		}

		mockStorage.
			EXPECT().
			Save(gomock.Any(), cmd.Name, cmd.File).
			Return("/storage/test.txt", nil)

		mockUserRepo.
			EXPECT().
			GetUsersByLogins(gomock.Any(), cmd.Grant).
			Return([]model.User{user}, nil)

		mockRepo.
			EXPECT().
			Create(gomock.Any(), gomock.AssignableToTypeOf(model.Doc{})).
			Return(model.Doc{}, errors.New("repository error"))

		mockStorage.
			EXPECT().
			Delete(gomock.Any()).
			Return(nil)

		service := NewService(
			mockUserRepo,
			mockRepo,
			mockStorage,
		)

		err := service.Upload(cmd)

		require.Error(t, err)
		assert.EqualError(t, err, "repository error")
	})
}

func TestService_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockUserRepo := NewMockUserRepository(ctrl)
		mockStorage := NewMockFileStorage(ctrl)

		user := model.User{ID: uuid.New(), Login: "user1"}
		doc := query.DocReadModel{
			ID:       uuid.New(),
			OwnerID:  user.ID,
			Name:     "test.txt",
			IsFile:   true,
			IsPublic: false,
			Grants:   []string{},
		}

		file := io.NopCloser(strings.NewReader("test file"))

		mockRepo.EXPECT().GetByID(gomock.Any(), doc.ID).Return(doc, nil)
		mockStorage.EXPECT().
			Read(doc.ID.String(), doc.Name).
			Return(file, nil)

		service := NewService(mockUserRepo, mockRepo, mockStorage)

		got, err := service.GetByID(doc.ID, user, false)

		require.NoError(t, err)
		assert.Equal(t, doc.ID, got.ID)
		assert.Equal(t, doc.Name, got.Name)
		assert.Equal(t, file, got.File)
	})

	t.Run("permission denied", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockUserRepo := NewMockUserRepository(ctrl)
		mockStorage := NewMockFileStorage(ctrl)

		user := model.User{ID: uuid.New(), Login: "user1"}
		doc := query.DocReadModel{
			ID:       uuid.New(),
			OwnerID:  uuid.New(),
			IsPublic: false,
			Grants:   []string{"user2"},
		}

		mockRepo.EXPECT().GetByID(gomock.Any(), doc.ID).Return(doc, nil)

		service := NewService(mockUserRepo, mockRepo, mockStorage)

		got, err := service.GetByID(doc.ID, user, true)

		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrDocPermissionDenied)
		assert.Equal(t, query.FullDocReadModel{}, got)
	})
	t.Run("without file", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockUserRepo := NewMockUserRepository(ctrl)
		mockStorage := NewMockFileStorage(ctrl)

		user := model.User{
			ID:    uuid.New(),
			Login: "user1",
		}

		doc := query.DocReadModel{
			ID:       uuid.New(),
			OwnerID:  user.ID,
			Name:     "data.json",
			IsFile:   false,
			IsPublic: false,
			JSONData: map[string]any{"name": "test"},
		}

		mockRepo.EXPECT().
			GetByID(gomock.Any(), doc.ID).
			Return(doc, nil)

		service := NewService(mockUserRepo, mockRepo, mockStorage)

		got, err := service.GetByID(doc.ID, user, false)

		require.NoError(t, err)
		assert.Equal(t, doc.ID, got.ID)
		assert.Equal(t, doc.JSONData, got.JSONData)
		assert.Nil(t, got.File)
	})
}

func TestService_DeleteByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockUserRepo := NewMockUserRepository(ctrl)
		mockStorage := NewMockFileStorage(ctrl)

		userID := uuid.New()
		doc := query.DocReadModel{
			ID:      uuid.New(),
			OwnerID: userID,
		}

		mockRepo.EXPECT().GetByID(gomock.Any(), doc.ID).Return(doc, nil)
		mockRepo.EXPECT().DeleteByID(gomock.Any(), doc.ID).Return(nil)
		mockStorage.EXPECT().Delete(doc.ID.String()).Return(nil)

		service := NewService(mockUserRepo, mockRepo, mockStorage)

		err := service.DeleteByID(doc.ID, userID)

		require.NoError(t, err)
	})

	t.Run("permission denied", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockUserRepo := NewMockUserRepository(ctrl)
		mockStorage := NewMockFileStorage(ctrl)

		doc := query.DocReadModel{
			ID:      uuid.New(),
			OwnerID: uuid.New(),
		}

		mockRepo.EXPECT().GetByID(gomock.Any(), doc.ID).Return(doc, nil)

		service := NewService(mockUserRepo, mockRepo, mockStorage)

		err := service.DeleteByID(doc.ID, uuid.New())

		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrDocPermissionDenied)
	})
}
