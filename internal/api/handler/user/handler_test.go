package user

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Mabarik667f/fsserver/internal/api/handler/user/dto"
	usercmd "github.com/Mabarik667f/fsserver/internal/command/user"
	"github.com/Mabarik667f/fsserver/internal/errs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupTestRouter(h *handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	r.Post("/api/register", h.Register)
	r.Post("/api/auth", h.Login)
	r.Delete("/api/auth/{token}", h.Logout)

	return r
}

func TestHandler_Register(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		validator := validator.New()
		mockSrv := NewMockService(ctrl)

		data := dto.RegisterRequest{
			AdminToken: "valid",
			Login:      "ValidLogin1",
			Password:   "Password47+-",
		}
		body, err := json.Marshal(data)
		require.NoError(t, err)

		mockSrv.EXPECT().Register(usercmd.RegisterUserCmd{
			AdminToken: data.AdminToken,
			Login:      data.Login,
			Password:   data.Password,
		}).Return(data.Login, nil)

		handler := NewHandler(mockSrv, &MockSessionManager{}, validator)

		req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
		w := httptest.NewRecorder()

		r := setupTestRouter(handler)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var got map[string]any
		require.NoError(t, json.NewDecoder(w.Body).Decode(&got))

		require.Equal(t, data.Login, got["response"].(map[string]any)["login"])
	})
	t.Run("admin token error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		mockSrv := NewMockService(ctrl)
		mockSessionManager := NewMockSessionManager(ctrl)

		validate := validator.New()

		data := dto.RegisterRequest{
			AdminToken: "invalid",
			Login:      "ValidLogin1",
			Password:   "Password47+-",
		}

		body, err := json.Marshal(data)
		require.NoError(t, err)

		mockSrv.EXPECT().
			Register(usercmd.RegisterUserCmd{
				AdminToken: data.AdminToken,
				Login:      data.Login,
				Password:   data.Password,
			}).
			Return("", errs.ErrAdminToken)

		handler := NewHandler(mockSrv, mockSessionManager, validate)

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/register",
			bytes.NewReader(body),
		)

		w := httptest.NewRecorder()

		r := setupTestRouter(handler)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestHandler_Login(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		mockSrv := NewMockService(ctrl)
		mockSessionManager := NewMockSessionManager(ctrl)

		validate := validator.New()

		data := dto.LoginRequest{
			Login:    "ValidLogin1",
			Password: "Password47+-",
		}

		userID := uuid.New()
		token := "test-token"

		mockSrv.EXPECT().
			Login(usercmd.LoginCmd{
				Login:    data.Login,
				Password: data.Password,
			}).
			Return(userID, nil)

		mockSessionManager.EXPECT().
			Commit(gomock.Any()).Return(token, time.Time{}, nil)

		mockSessionManager.EXPECT().
			Put(gomock.Any(), "user_id", userID.String())

		handler := NewHandler(
			mockSrv,
			mockSessionManager,
			validate,
		)

		body, err := json.Marshal(data)
		require.NoError(t, err)

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/auth",
			bytes.NewReader(body),
		)

		w := httptest.NewRecorder()

		r := setupTestRouter(handler)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var got map[string]any
		require.NoError(t, json.NewDecoder(w.Body).Decode(&got))

		responseData := got["response"].(map[string]any)

		require.Equal(t, token, responseData["token"])
	})
	t.Run("password mismatch", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		mockSrv := NewMockService(ctrl)
		mockSessionManager := NewMockSessionManager(ctrl)

		data := dto.LoginRequest{
			Login:    "ValidLogin1",
			Password: "Password47+-",
		}

		body, err := json.Marshal(data)
		require.NoError(t, err)

		mockSrv.EXPECT().
			Login(gomock.Any()).
			Return(uuid.Nil, errs.ErrPasswordsNotEqual)

		handler := NewHandler(
			mockSrv,
			mockSessionManager,
			validator.New(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/auth",
			bytes.NewReader(body),
		)

		w := httptest.NewRecorder()

		r := setupTestRouter(handler)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestHandler_Logout(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		mockSrv := NewMockService(ctrl)
		mockSessionManager := NewMockSessionManager(ctrl)

		token := "test-token"

		mockSessionManager.EXPECT().Exists(gomock.Any(), token).Return(true)

		mockSessionManager.EXPECT().
			Remove(gomock.Any(), token)

		handler := NewHandler(
			mockSrv,
			mockSessionManager,
			validator.New(),
		)

		req := httptest.NewRequest(
			http.MethodDelete,
			fmt.Sprintf("/api/auth/%s", token),
			nil,
		)

		w := httptest.NewRecorder()

		r := setupTestRouter(handler)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var got map[string]any
		require.NoError(t, json.NewDecoder(w.Body).Decode(&got))

		responseData := got["response"].(map[string]any)

		require.Equal(t, true, responseData[token])
	})

	t.Run("invalid session", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		mockSrv := NewMockService(ctrl)
		mockSessionManager := NewMockSessionManager(ctrl)

		token := "invalid"

		mockSessionManager.EXPECT().Exists(gomock.Any(), token).Return(false)

		handler := NewHandler(
			mockSrv,
			mockSessionManager,
			validator.New(),
		)

		req := httptest.NewRequest(
			http.MethodDelete,
			fmt.Sprintf("/api/auth/%s", token),
			nil,
		)

		w := httptest.NewRecorder()

		r := setupTestRouter(handler)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)
	})
}
