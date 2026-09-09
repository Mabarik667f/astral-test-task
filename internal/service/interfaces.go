package service

import (
	usercmd "github.com/Mabarik667f/fsserver/internal/command/user"
	"github.com/google/uuid"
)

type UserService interface {
	Register(cmd usercmd.RegisterUserCmd) (string, error)
	Login(cmd usercmd.LoginCmd) (uuid.UUID, error)
}
