package app

import (
	"net/http"

	"github.com/Mabarik667f/fsserver/internal/api/handler"
	"github.com/go-playground/validator/v10"
)

type diContainer struct {
	api http.Handler

	validator *validator.Validate
}

func newDiContainer() *diContainer {
	return &diContainer{}
}

func (d *diContainer) API() http.Handler {
	if d.api == nil {
		d.api = handler.API()
	}

	return d.api
}
