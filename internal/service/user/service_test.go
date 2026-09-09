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
			Password:   "Password48+-",
			Login:      "ValidLogin2",
			AdminToken: "secret",
		}

		mockRepo.
			EXPECT().
			Create(gomock.Any(), gomock.AssignableToTypeOf(model.User{})).
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
			Create(gomock.Any(), gomock.AssignableToTypeOf(model.User{})).
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

func TestService_Login(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockHasher := NewMockPasswordHasher(ctrl)

		cmd := usercmd.LoginCmd{
			Login:    "SomeLogin1",
			Password: "validPassword47)",
		}

		user := model.User{Login: cmd.Login, Password: model.PasswordHash(cmd.Password)}

		mockRepo.
			EXPECT().
			GetByLoginWithPasswordHash(gomock.Any(), cmd.Login).
			Return(user, nil)

		mockHasher.
			EXPECT().
			Verify(string(user.Password), cmd.Password).
			Return(true, nil)

		service := NewService(mockRepo, mockHasher, "secret")
		res, err := service.Login(cmd)

		require.NoError(t, err)
		require.NotNil(t, res)
	})

	t.Run("password not equal", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockHasher := NewMockPasswordHasher(ctrl)

		cmd := usercmd.LoginCmd{
			Login:    "SomeLogin1",
			Password: "validPassword47)",
		}

		mockRepo.
			EXPECT().
			GetByLoginWithPasswordHash(gomock.Any(), cmd.Login).
			Return(model.User{Login: cmd.Login, Password: model.PasswordHash("other")}, nil)

		mockHasher.
			EXPECT().
			Verify("other", cmd.Password).
			Return(false, nil)

		service := NewService(mockRepo, mockHasher, "secret")
		_, err := service.Login(cmd)

		require.Error(t, err)
		require.ErrorIs(t, err, errs.ErrPasswordsNotEqual)
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockHasher := NewMockPasswordHasher(ctrl)

		cmd := usercmd.LoginCmd{
			Login:    "SomeLogin1",
			Password: "validPassword47)",
		}

		mockRepo.
			EXPECT().
			GetByLoginWithPasswordHash(gomock.Any(), cmd.Login).
			Return(model.User{}, errs.ErrUserNotFound)

		service := NewService(mockRepo, mockHasher, "secret")
		_, err := service.Login(cmd)

		require.Error(t, err)
		require.ErrorIs(t, err, errs.ErrUserNotFound)
	})
}
