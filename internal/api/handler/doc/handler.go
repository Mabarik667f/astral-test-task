package doc

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/Mabarik667f/fsserver/internal/api/handler/doc/dto"
	doccmd "github.com/Mabarik667f/fsserver/internal/command/doc"
	"github.com/Mabarik667f/fsserver/internal/errs"
	"github.com/Mabarik667f/fsserver/internal/model"
	"github.com/Mabarik667f/fsserver/internal/model/query"
	"github.com/Mabarik667f/fsserver/pkg/http/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/schema"
)

//go:generate mockgen -destination=handler_mocks.go -source=handler.go -package=doc

type Service interface {
	Upload(cmd doccmd.CreateDocumentCmd) error
	DeleteByID(id, userID uuid.UUID) error
	GetByID(id uuid.UUID, user model.User, metaOnly bool) (query.FullDocReadModel, error)
	Get(cmd doccmd.GetDocumentsListCmd, user model.User) ([]query.DocReadModel, error)
}

type SessionManager interface {
	GetString(ctx context.Context, key string) string
}

type handler struct {
	service        Service
	validator      *validator.Validate
	decoder        *schema.Decoder
	sessionManager SessionManager
}

func NewHandler(
	service Service,
	validator *validator.Validate,
	decoder *schema.Decoder,
	sessionManager SessionManager,
) *handler {
	return &handler{
		service:        service,
		validator:      validator,
		decoder:        decoder,
		sessionManager: sessionManager,
	}
}

func (h *handler) Upload(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		response.ErrResponse(
			w,
			http.StatusUnauthorized,
			errs.ErrTokenNotProvided.Error(),
		)
		return
	}

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
		OwnerID:  user.ID,
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
				errs.ErrInternalServer.Error(),
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

func (h *handler) Get(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		response.ErrResponse(
			w,
			http.StatusUnauthorized,
			errs.ErrTokenNotProvided.Error(),
		)
		return
	}

	var req dto.GetDocumentsListRequest

	if err := h.decoder.Decode(&req, r.URL.Query()); err != nil {
		response.ErrResponse(w, http.StatusBadRequest, "invalid query params format")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		response.ErrResponse(w, http.StatusBadRequest, "validation error query params")
		return
	}

	cmd := doccmd.GetDocumentsListCmd{
		Key:   req.Key,
		Value: req.Value,
		Limit: req.Limit,
	}

	if req.Login != "" {
		cmd.Login = &req.Login
	}

	docs, err := h.service.Get(cmd, user)
	if err != nil {
		response.ErrResponse(
			w,
			http.StatusInternalServerError,
			errs.ErrInternalServer.Error(),
		)
		return
	}

	if r.Method == "HEAD" {
		w.WriteHeader(http.StatusOK)
		return
	}

	res := make([]dto.DocReadModelResponse, len(docs))
	for i := range docs {
		res[i] = dto.ToDocReadModelResponse(docs[i])
	}

	response.JSON(w, http.StatusOK, response.Resp{
		Data: map[string][]dto.DocReadModelResponse{
			"docs": res,
		},
	})
}

func (h *handler) GetByID(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		response.ErrResponse(
			w,
			http.StatusUnauthorized,
			errs.ErrTokenNotProvided.Error(),
		)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.ErrResponse(w, http.StatusBadRequest, "invalid param type")
		return
	}

	metaOnly := r.Method == "HEAD"

	doc, err := h.service.GetByID(id, user, metaOnly)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrDocNotFound):
			response.ErrResponse(
				w,
				http.StatusNotFound,
				err.Error(),
			)
		case errors.Is(err, errs.ErrDocPermissionDenied):
			response.ErrResponse(
				w,
				http.StatusForbidden,
				err.Error(),
			)
		default:
			response.ErrResponse(
				w,
				http.StatusInternalServerError,
				errs.ErrInternalServer.Error(),
			)
		}
		return
	}

	if doc.IsFile {
		h.sendFile(w, doc, metaOnly)
		return
	}

	res := dto.ToDocReadModelFromFullResponse(doc)

	response.JSON(w, http.StatusOK, response.Resp{
		Data: res,
	})
}

func (h *handler) DeleteByID(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		response.ErrResponse(
			w,
			http.StatusUnauthorized,
			errs.ErrTokenNotProvided.Error(),
		)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.ErrResponse(w, http.StatusBadRequest, "invalid param type")
		return
	}

	if err := h.service.DeleteByID(id, user.ID); err != nil {
		switch {
		case errors.Is(err, errs.ErrDocNotFound):
			response.ErrResponse(
				w,
				http.StatusNotFound,
				err.Error(),
			)
		case errors.Is(err, errs.ErrDocPermissionDenied):
			response.ErrResponse(
				w,
				http.StatusForbidden,
				err.Error(),
			)
		default:
			response.ErrResponse(
				w,
				http.StatusInternalServerError,
				errs.ErrInternalServer.Error(),
			)
		}
		return
	}

	response.JSON(w, http.StatusOK, response.Resp{
		Response: map[string]bool{
			id.String(): true,
		},
	})
}

func (h *handler) sendFile(
	w http.ResponseWriter,
	doc query.FullDocReadModel,
	metaOnly bool,
) {
	w.Header().Set("Content-Type", doc.MimeType)

	if metaOnly {
		w.WriteHeader(http.StatusOK)
		return
	}

	defer doc.File.Close()
	w.WriteHeader(http.StatusOK)

	if _, err := io.Copy(w, doc.File); err != nil {
		slog.Error("error to send file data", "err", err)
	}
}

const (
	userIDKey    = "userID"
	userLoginKey = "userLogin"
)

func (h *handler) getUser(r *http.Request) (model.User, bool) {
	userID := h.sessionManager.GetString(r.Context(), userIDKey)
	if userID == "" {
		return model.User{}, false
	}

	userLogin := h.sessionManager.GetString(r.Context(), userLoginKey)
	if userLogin == "" {
		return model.User{}, false
	}

	return model.User{ID: uuid.MustParse(userID), Login: userLogin}, true
}
