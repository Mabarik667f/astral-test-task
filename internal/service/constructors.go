package service

import (
	"github.com/Mabarik667f/fsserver/internal/service/doc"
	"github.com/Mabarik667f/fsserver/internal/service/user"
)

func NewUserService(
	repo user.Repository,
	hasher user.PasswordHasher,
	adminToken string,
) UserService {
	return user.NewService(repo, hasher, adminToken)
}

func NewDocService(
	userRepo doc.UserRepository,
	repo doc.Repository,
	storage doc.FileStorage,
) DocService {
	return doc.NewService(userRepo, repo, storage)
}
