package doc

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Mabarik667f/fsserver/internal/api/handler/doc/dto"
	doccmd "github.com/Mabarik667f/fsserver/internal/command/doc"
	"github.com/Mabarik667f/fsserver/internal/model"
	"github.com/Mabarik667f/fsserver/internal/model/query"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

func setupTestRouter(h *handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	r.Post("/api/docs", h.Upload)
	r.Get("/api/docs", h.Get)
	r.Get("/api/docs/{id}", h.GetByID)
	r.Delete("/api/docs/{id}", h.DeleteByID)

	return r
}

func TestHandler_Upload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		mockSrv := NewMockService(ctrl)
		validate := validator.New()
		decoder := schema.NewDecoder()
		mockManager := NewMockSessionManager(ctrl)

		user := model.User{ID: uuid.New(), Login: "ValidLogin2"}

		fileContent := "test file content"

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)

		meta := dto.MetaData{
			Name:     "photo.jpg",
			IsFile:   true,
			IsPublic: false,
			MimeType: "image/jpeg",
			Grant:    []string{"ValidLogin2"},
		}

		metaJSON, err := json.Marshal(meta)
		require.NoError(t, err)

		require.NoError(t, writer.WriteField("meta", string(metaJSON)))
		require.NoError(t, writer.WriteField(
			"json",
			`{"description":123}`,
		))

		fileWriter, err := writer.CreateFormFile("file", "photo.jpg")
		require.NoError(t, err)

		_, err = fileWriter.Write([]byte(fileContent))
		require.NoError(t, err)

		require.NoError(t, writer.Close())

		mockSrv.EXPECT().
			Upload(gomock.Any()).
			DoAndReturn(func(cmd doccmd.CreateDocumentCmd) error {
				assert.Equal(t, user.ID, cmd.OwnerID)
				assert.Equal(t, "photo.jpg", cmd.Name)
				assert.True(t, cmd.IsFile)
				assert.False(t, cmd.IsPublic)
				assert.Equal(t, "image/jpeg", cmd.MimeType)
				assert.Equal(t, []string{"ValidLogin2"}, cmd.Grant)
				assert.Equal(t, map[string]any{
					"description": float64(123),
				}, cmd.JSONData)

				require.NotNil(t, cmd.File)

				gotFile, err := io.ReadAll(cmd.File)
				require.NoError(t, err)
				assert.Equal(t, fileContent, string(gotFile))

				return nil
			})

		mockManager.EXPECT().GetString(gomock.Any(), "userID").Return(user.ID.String())
		mockManager.EXPECT().GetString(gomock.Any(), "userLogin").Return(user.Login)

		handler := NewHandler(
			mockSrv,
			validate,
			decoder,
			mockManager,
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/docs",
			&body,
		)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()

		r := setupTestRouter(handler)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var got map[string]any
		require.NoError(t, json.NewDecoder(w.Body).Decode(&got))

		data := got["data"].(map[string]any)

		assert.Equal(t, "photo.jpg", data["file"])
		assert.Equal(
			t,
			map[string]any{"description": float64(123)},
			data["json"],
		)
	})
}

func TestHandler_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		mockSrv := NewMockService(ctrl)
		mockManager := NewMockSessionManager(ctrl)

		validate := validator.New()
		decoder := schema.NewDecoder()

		user := model.User{
			ID:    uuid.New(),
			Login: "ValidLogin2",
		}

		cmd := doccmd.GetDocumentsListCmd{
			Key:   "mime",
			Value: "image/jpeg",
			Limit: 10,
		}

		expected := []query.DocReadModel{
			{
				ID:       uuid.New(),
				Name:     "photo.jpg",
				IsFile:   true,
				IsPublic: false,
				MimeType: "image/jpeg",
			},
		}

		mockManager.EXPECT().
			GetString(gomock.Any(), userIDKey).
			Return(user.ID.String())

		mockManager.EXPECT().
			GetString(gomock.Any(), userLoginKey).
			Return(user.Login)

		mockSrv.EXPECT().
			Get(cmd, user).
			Return(expected, nil)

		handler := NewHandler(
			mockSrv,
			validate,
			decoder,
			mockManager,
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/docs?key=mime&value=image/jpeg&limit=10",
			nil,
		)

		w := httptest.NewRecorder()

		r := setupTestRouter(handler)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var got map[string]any
		require.NoError(t, json.NewDecoder(w.Body).Decode(&got))

		data, ok := got["data"].(map[string]any)
		require.True(t, ok)

		docs, ok := data["docs"].([]any)
		require.True(t, ok)
		require.Len(t, docs, 1)

		doc := docs[0].(map[string]any)

		assert.Equal(t, "photo.jpg", doc["name"])
		assert.Equal(t, "image/jpeg", doc["mime"])
	})
}

func TestHandler_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		mockSrv := NewMockService(ctrl)
		mockManager := NewMockSessionManager(ctrl)

		validate := validator.New()
		decoder := schema.NewDecoder()

		user := model.User{
			ID:    uuid.New(),
			Login: "ValidLogin2",
		}

		docID := uuid.New()

		expected := query.FullDocReadModel{
			ID:       docID,
			OwnerID:  user.ID,
			Name:     "document.json",
			IsFile:   false,
			IsPublic: true,
			MimeType: "application/json",
			JSONData: map[string]any{"foo": "bar"},
		}

		mockManager.EXPECT().
			GetString(gomock.Any(), userIDKey).
			Return(user.ID.String())

		mockManager.EXPECT().
			GetString(gomock.Any(), userLoginKey).
			Return(user.Login)

		mockSrv.EXPECT().
			GetByID(docID, user, false).
			Return(expected, nil)

		handler := NewHandler(
			mockSrv,
			validate,
			decoder,
			mockManager,
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/docs/"+docID.String(),
			nil,
		)

		w := httptest.NewRecorder()

		r := setupTestRouter(handler)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var got map[string]any
		require.NoError(t, json.NewDecoder(w.Body).Decode(&got))

		data, ok := got["data"].(map[string]any)
		require.True(t, ok)

		assert.Equal(t, docID.String(), data["id"])
		assert.Equal(t, "document.json", data["name"])
	})
}

func TestHandler_DeleteByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		mockSrv := NewMockService(ctrl)
		mockManager := NewMockSessionManager(ctrl)

		validate := validator.New()
		decoder := schema.NewDecoder()

		user := model.User{
			ID:    uuid.New(),
			Login: "ValidLogin2",
		}

		docID := uuid.New()

		mockManager.EXPECT().
			GetString(gomock.Any(), userIDKey).
			Return(user.ID.String())

		mockManager.EXPECT().
			GetString(gomock.Any(), userLoginKey).
			Return(user.Login)

		mockSrv.EXPECT().
			DeleteByID(docID, user.ID).
			Return(nil)

		handler := NewHandler(
			mockSrv,
			validate,
			decoder,
			mockManager,
		)

		req := httptest.NewRequest(
			http.MethodDelete,
			"/api/docs/"+docID.String(),
			nil,
		)

		w := httptest.NewRecorder()

		r := setupTestRouter(handler)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var got map[string]any
		require.NoError(t, json.NewDecoder(w.Body).Decode(&got))

		responseData, ok := got["response"].(map[string]any)
		require.True(t, ok)

		assert.Equal(t, true, responseData[docID.String()])
	})
}
