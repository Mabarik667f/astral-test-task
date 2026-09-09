package doc

import (
	"errors"
	"strings"
	"testing"

	doccmd "github.com/Mabarik667f/fsserver/internal/command/doc"
	"github.com/Mabarik667f/fsserver/internal/errs"
	"github.com/Mabarik667f/fsserver/internal/model"
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
