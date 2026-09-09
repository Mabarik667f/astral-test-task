package handler

import (
	"github.com/Mabarik667f/fsserver/internal/api/handler/doc"
	"github.com/Mabarik667f/fsserver/internal/api/handler/user"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
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
	decoder *schema.Decoder,
	sessionManager doc.SessionManager,
) DocHandler {
	return doc.NewHandler(srv, validator, decoder, sessionManager)
}
