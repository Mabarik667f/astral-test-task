package user

import (
	"context"

	usercmd "github.com/Mabarik667f/fsserver/internal/command/user"
	"github.com/Mabarik667f/fsserver/internal/errs"
	"github.com/Mabarik667f/fsserver/internal/model"
	"github.com/google/uuid"
)

//go:generate mockgen -destination=service_mocks.go -source=service.go -package=user

type Repository interface {
	Create(ctx context.Context, user model.User) (model.User, error)
	GetByLoginWithPasswordHash(ctx context.Context, login string) (model.User, error)
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(storedHash, providedPassword string) (bool, error)
}

type service struct {
	repo       Repository
	hasher     PasswordHasher
	adminToken string
}

func NewService(
	repo Repository,
	hasher PasswordHasher,
	adminToken string,
) *service {
	return &service{repo: repo, hasher: hasher, adminToken: adminToken}
}

func (s *service) Register(cmd usercmd.RegisterUserCmd) (string, error) {
	if cmd.AdminToken != s.adminToken {
		return "", errs.ErrAdminToken
	}

	if err := model.ValidatePassword(cmd.Password); err != nil {
		return "", errs.ErrPasswordPatterMatch
	}

	pswdHash, err := s.hasher.Hash(cmd.Password)
	if err != nil {
		return "", err
	}

	user, err := model.NewUser(model.PasswordHash(pswdHash), cmd.Login)
	if err != nil {
		return "", err
	}

	if _, err := s.repo.Create(context.Background(), *user); err != nil {
		return "", err
	}

	return user.Login, nil
}

func (s *service) Login(cmd usercmd.LoginCmd) (uuid.UUID, error) {
	user, err := s.repo.GetByLoginWithPasswordHash(context.Background(), cmd.Login)
	if err != nil {
		return uuid.Nil, err
	}

	ok, err := s.hasher.Verify(string(user.Password), cmd.Password)
	if err != nil {
		return uuid.Nil, err
	}

	if !ok {
		return uuid.Nil, errs.ErrPasswordsNotEqual
	}

	return user.ID, nil
}
