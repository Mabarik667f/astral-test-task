package service

import (
	doccmd "github.com/Mabarik667f/fsserver/internal/command/doc"
	usercmd "github.com/Mabarik667f/fsserver/internal/command/user"
	"github.com/Mabarik667f/fsserver/internal/model"
	"github.com/Mabarik667f/fsserver/internal/model/query"
	"github.com/google/uuid"
)

type UserService interface {
	Register(cmd usercmd.RegisterUserCmd) (string, error)
	Login(cmd usercmd.LoginCmd) (uuid.UUID, error)
}

type DocService interface {
	Upload(cmd doccmd.CreateDocumentCmd) error
	DeleteByID(id, userID uuid.UUID) error
	GetByID(id uuid.UUID, user model.User, metaOnly bool) (query.FullDocReadModel, error)
	Get(cmd doccmd.GetDocumentsListCmd, user model.User) ([]query.DocReadModel, error)
}
