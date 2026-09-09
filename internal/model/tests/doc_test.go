package tests

import (
	"testing"

	"github.com/Mabarik667f/fsserver/internal/errs"
	"github.com/Mabarik667f/fsserver/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewDoc(t *testing.T) {
	type args struct {
		ownerID  uuid.UUID
		name     string
		isFile   bool
		isPublic bool
		mimeType string
		jsonData map[string]any
		filePath string
	}

	ownerID := uuid.New()
	jsonData := map[string]any{
		"description": "test document",
		"size":        float64(123),
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				ownerID:  ownerID,
				name:     "photo.jpg",
				isFile:   true,
				isPublic: false,
				mimeType: "image/jpeg",
				jsonData: jsonData,
				filePath: "/storage/photo.jpg",
			},
			wantErr: nil,
		},
		{
			name: "empty name",
			args: args{
				ownerID:  ownerID,
				name:     "",
				isFile:   true,
				isPublic: false,
				mimeType: "image/jpeg",
				jsonData: jsonData,
				filePath: "/storage/photo.jpg",
			},
			wantErr: errs.ErrDocNameEmpty,
		},
		{
			name: "empty mime type",
			args: args{
				ownerID:  ownerID,
				name:     "photo.jpg",
				isFile:   true,
				isPublic: false,
				mimeType: "",
				jsonData: jsonData,
				filePath: "/storage/photo.jpg",
			},
			wantErr: errs.ErrDocMimeTypeEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc, err := model.NewDoc(
				tt.args.ownerID,
				tt.args.name,
				tt.args.isFile,
				tt.args.isPublic,
				tt.args.mimeType,
				tt.args.jsonData,
				tt.args.filePath,
			)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, doc)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, doc)

			assert.NotEqual(t, uuid.Nil, doc.ID)
			assert.Equal(t, tt.args.ownerID, doc.OwnerID)
			assert.Equal(t, tt.args.name, doc.Name)
			assert.Equal(t, tt.args.isFile, doc.IsFile)
			assert.Equal(t, tt.args.isPublic, doc.IsPublic)
			assert.Equal(t, tt.args.mimeType, doc.MimeType)
			assert.Equal(t, tt.args.jsonData, doc.JSONData)
			assert.Equal(t, tt.args.filePath, doc.FilePath)
			assert.False(t, doc.CreatedAt.IsZero())
		})
	}
}
