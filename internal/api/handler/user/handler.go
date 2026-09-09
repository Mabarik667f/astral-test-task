package user

//go:generate mockgen -destination=handler_mocks.go -source=handler.go -package=user

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Mabarik667f/fsserver/internal/api/handler/user/dto"
	usercmd "github.com/Mabarik667f/fsserver/internal/command/user"
	"github.com/Mabarik667f/fsserver/internal/errs"
	"github.com/Mabarik667f/fsserver/pkg/http/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Service interface {
	Register(cmd usercmd.RegisterUserCmd) (string, error)
	Login(cmd usercmd.LoginCmd) (uuid.UUID, error)
}

type SessionManager interface {
	Put(ctx context.Context, key string, value any)
	Remove(ctx context.Context, key string)
	Exists(ctx context.Context, key string) bool
	Commit(ctx context.Context) (string, time.Time, error)
}

type handler struct {
	validator      *validator.Validate
	sessionManager SessionManager
	service        Service
}

func NewHandler(
	service Service,
	sessionManager SessionManager,
	validator *validator.Validate,
) *handler {
	return &handler{
		service:        service,
		sessionManager: sessionManager,
		validator:      validator,
	}
}

func (h *handler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrResponse(w, http.StatusBadRequest, errs.ErrInvalidJSON.Error())
		return
	}

	if err := h.validator.Struct(req); err != nil {
		response.ErrResponse(w, http.StatusBadRequest, errs.ErrJSONValidation.Error())
		return
	}

	cmd := usercmd.RegisterUserCmd{
		AdminToken: req.AdminToken,
		Login:      req.Login,
		Password:   req.Password,
	}

	login, err := h.service.Register(cmd)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrAdminToken):
			response.ErrResponse(w, http.StatusForbidden, err.Error())
		case errors.Is(err, errs.ErrUserUnique):
			response.ErrResponse(w, http.StatusConflict, err.Error())
		case errors.Is(err, errs.ErrLoginPatternMatch) ||
			errors.Is(err, errs.ErrPasswordPatterMatch):
			response.ErrResponse(w, http.StatusBadRequest, err.Error())
		default:
			response.ErrResponse(w, http.StatusInternalServerError, errs.ErrInternalServer.Error())
		}
		return
	}

	response.JSON(w, http.StatusOK, response.Resp{
		Response: dto.RegisterResponse{Login: login},
	})
}

func (h *handler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrResponse(w, http.StatusBadRequest, errs.ErrInvalidJSON.Error())
		return
	}

	if err := h.validator.Struct(req); err != nil {
		response.ErrResponse(w, http.StatusBadRequest, errs.ErrJSONValidation.Error())
		return
	}

	cmd := usercmd.LoginCmd{
		Login:    req.Login,
		Password: req.Password,
	}

	userID, err := h.service.Login(cmd)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrUserNotFound):
			response.ErrResponse(w, http.StatusNotFound, err.Error())
			return
		case errors.Is(err, errs.ErrPasswordsNotEqual):
			response.ErrResponse(w, http.StatusUnauthorized, err.Error())
			return
		default:
			response.ErrResponse(w, http.StatusInternalServerError, errs.ErrInternalServer.Error())
		}
		return
	}

	h.sessionManager.Put(r.Context(), "user_id", userID.String())
	token, _, err := h.sessionManager.Commit(r.Context())
	if err != nil {
		response.ErrResponse(
			w,
			http.StatusInternalServerError,
			errs.ErrInternalServer.Error(),
		)
		return
	}
	response.JSON(w, http.StatusOK, response.Resp{
		Response: map[string]string{
			"token": token,
		},
	})
}

func (h *handler) Logout(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		response.ErrResponse(w, http.StatusBadRequest, errs.ErrTokenNotProvided.Error())
		return
	}

	if ok := h.sessionManager.Exists(r.Context(), token); !ok {
		response.ErrResponse(w, http.StatusNotFound, errs.ErrSessionNoExists.Error())
		return
	}

	h.sessionManager.Remove(r.Context(), token)
	response.JSON(w, http.StatusOK, response.Resp{
		Response: map[string]bool{
			token: true,
		},
	})
}
