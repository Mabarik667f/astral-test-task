package model

import (
	"time"

	"github.com/Mabarik667f/fsserver/internal/errs"
	"github.com/google/uuid"
)

type Doc struct {
	ID        uuid.UUID
	OwnerID   uuid.UUID
	Name      string
	IsFile    bool
	IsPublic  bool
	MimeType  string
	JSONData  map[string]any
	FilePath  string
	CreatedAt time.Time
}

func NewDoc(
	ownerID uuid.UUID,
	name string,
	isFile, isPublic bool,
	mimeType string,
	jsonData map[string]any,
	filePath string,
) (*Doc, error) {
	if name == "" {
		return nil, errs.ErrDocNameEmpty
	}

	if mimeType == "" {
		return nil, errs.ErrDocMimeTypeEmpty
	}

	return &Doc{
		ID:        uuid.New(),
		OwnerID:   ownerID,
		Name:      name,
		IsFile:    isFile,
		IsPublic:  isPublic,
		MimeType:  mimeType,
		JSONData:  jsonData,
		FilePath:  filePath,
		CreatedAt: time.Now().UTC(),
	}, nil
}
