package handler

import (
	"github.com/Mabarik667f/fsserver/internal/api/handler/doc"
	"github.com/Mabarik667f/fsserver/internal/api/handler/user"
	"github.com/go-playground/validator/v10"
)

func NewUserHandler(
	srv user.Service,
	sessionManager user.SessionManager,
	validator *validator.Validate,
) UserHandler {
	return user.NewHandler(srv, sessionManager, validator)
}

func NewDocHandler(
	srv doc.Service,
	validator *validator.Validate,
) DocHandler {
	return doc.NewHandler(srv, validator)
}
