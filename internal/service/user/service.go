package user

import (
	usercmd "github.com/Mabarik667f/fsserver/internal/command/user"
	"github.com/Mabarik667f/fsserver/internal/errs"
	"github.com/Mabarik667f/fsserver/internal/model"
)

//go:generate mockgen -destination=service_mocks.go -source=service.go -package=user

type Repository interface {
	Create(user model.User) (model.User, error)
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

func NewService(repo Repository, hasher PasswordHasher, adminToken string) *service {
	return &service{repo: repo, hasher: hasher, adminToken: adminToken}
}

func (s *service) Register(cmd usercmd.RegisterUserCmd) (string, error) {
	if cmd.AdminToken != s.adminToken {
		return "", errs.ErrAdminToken
	}

	if err := model.ValidatePassword(cmd.Password); err != nil {
		return "", err
	}

	pswdHash, err := s.hasher.Hash(cmd.Password)
	if err != nil {
		return "", err
	}

	user, err := model.NewUser(model.PasswordHash(pswdHash), cmd.Login)
	if err != nil {
		return "", err
	}

	if _, err := s.repo.Create(*user); err != nil {
		return "", err
	}

	return user.Login, nil
}
