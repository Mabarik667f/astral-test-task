package user

import (
	"errors"
	"testing"

	usercmd "github.com/Mabarik667f/fsserver/internal/command/user"
	"github.com/Mabarik667f/fsserver/internal/errs"
	"github.com/Mabarik667f/fsserver/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService_Register(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockHasher := NewMockPasswordHasher(ctrl)

		cmd := usercmd.RegisterUserCmd{
			Password:   "Password47+-",
			Login:      "ValidLogin1",
			AdminToken: "secret",
		}

		mockRepo.
			EXPECT().
			Create(gomock.AssignableToTypeOf(model.User{})).
			Return(model.User{}, nil)

		mockHasher.
			EXPECT().
			Hash(string(cmd.Password)).
			Return("hashedPassword", nil)

		service := NewService(mockRepo, mockHasher, "secret")

		login, err := service.Register(cmd)

		require.NoError(t, err)
		assert.Equal(t, cmd.Login, login)
	})

	t.Run("password validation failed", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockHasher := NewMockPasswordHasher(ctrl)

		cmd := usercmd.RegisterUserCmd{
			Password:   "inv",
			Login:      "ValidLogin1",
			AdminToken: "secret",
		}
		service := NewService(mockRepo, mockHasher, "secret")

		_, err := service.Register(cmd)

		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrPasswordTooShort)
	})

	t.Run("password hasher error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockHasher := NewMockPasswordHasher(ctrl)

		cmd := usercmd.RegisterUserCmd{
			Password:   "Password47+-",
			Login:      "ValidLogin1",
			AdminToken: "secret",
		}

		mockHasher.
			EXPECT().
			Hash(string(cmd.Password)).
			Return("", errors.New("password hashing error"))

		service := NewService(mockRepo, mockHasher, "secret")

		_, err := service.Register(cmd)
		require.Error(t, err)
	})
	t.Run("invalid login", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockHasher := NewMockPasswordHasher(ctrl)

		cmd := usercmd.RegisterUserCmd{
			Password:   "Password47+-",
			Login:      "invalid",
			AdminToken: "secret",
		}

		mockHasher.
			EXPECT().
			Hash(string(cmd.Password)).
			Return("hashedPassword", nil)

		service := NewService(mockRepo, mockHasher, "secret")

		_, err := service.Register(cmd)

		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrLoginPatternMatch)
	})
	t.Run("repository error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockHasher := NewMockPasswordHasher(ctrl)

		cmd := usercmd.RegisterUserCmd{
			Password:   "Password47+-",
			Login:      "ValidLogin1",
			AdminToken: "secret",
		}

		mockRepo.
			EXPECT().
			Create(gomock.AssignableToTypeOf(model.User{})).
			Return(model.User{}, errors.New("repository error"))

		mockHasher.
			EXPECT().
			Hash(string(cmd.Password)).
			Return("hashedPassword", nil)

		service := NewService(mockRepo, mockHasher, "secret")

		_, err := service.Register(cmd)
		require.Error(t, err)
	})
}
