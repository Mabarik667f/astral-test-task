package doc

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Mabarik667f/fsserver/internal/api/handler/doc/dto"
	doccmd "github.com/Mabarik667f/fsserver/internal/command/doc"
	"github.com/Mabarik667f/fsserver/internal/errs"
	"github.com/Mabarik667f/fsserver/pkg/http/response"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

//go:generate mockgen -destination=handler_mocks.go -source=handler.go -package=doc

type Service interface {
	Upload(cmd doccmd.CreateDocumentCmd) error
}

type handler struct {
	service   Service
	validator *validator.Validate
}

func NewHandler(service Service, validator *validator.Validate) *handler {
	return &handler{service: service, validator: validator}
}

func (h *handler) Upload(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			file = nil
		} else {
			response.ErrResponse(
				w,
				http.StatusBadRequest,
				"error to read file from form-data",
			)
			return
		}
	}

	if file != nil {
		defer file.Close()
	}

	meta := r.FormValue("meta")
	var metaData dto.MetaData

	if err := json.Unmarshal([]byte(meta), &metaData); err != nil {
		response.ErrResponse(w, http.StatusBadRequest, errs.ErrInvalidJSON.Error())
		return
	}

	if err := h.validator.Struct(metaData); err != nil {
		response.ErrResponse(w, http.StatusBadRequest, errs.ErrJSONValidation.Error())
		return
	}

	jsData := r.FormValue("json")
	var jsonData map[string]any

	if jsData != "" {
		if err := json.Unmarshal([]byte(jsData), &jsonData); err != nil {
			response.ErrResponse(
				w,
				http.StatusBadRequest,
				errs.ErrInvalidJSON.Error(),
			)
			return
		}
	}

	cmd := doccmd.CreateDocumentCmd{
		OwnerID:  uuid.MustParse("4a3a5a37-cee2-4abd-a221-8ee2dcb47d27"),
		Name:     metaData.Name,
		IsFile:   metaData.IsFile,
		IsPublic: metaData.IsPublic,
		MimeType: metaData.MimeType,
		Grant:    metaData.Grant,
		JSONData: jsonData,
		File:     file,
	}

	err = h.service.Upload(cmd)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrDocBusiness):
			response.ErrResponse(
				w,
				http.StatusBadRequest,
				err.Error(),
			)
		case errors.Is(err, errs.ErrSaveFileToStorage):
			response.ErrResponse(
				w,
				http.StatusInternalServerError,
				err.Error(),
			)
		default:
			response.ErrResponse(
				w,
				http.StatusInternalServerError,
				err.Error(),
			)
		}
		return
	}

	response.JSON(w, http.StatusOK, response.Resp{
		Data: dto.UploadFileResponse{
			JSON:     jsonData,
			FileName: metaData.Name,
		},
	})
}

func (h *handler) Get(w http.ResponseWriter, r *http.Request)        {}
func (h *handler) GetByID(w http.ResponseWriter, r *http.Request)    {}
func (h *handler) DeleteByID(w http.ResponseWriter, r *http.Request) {}
