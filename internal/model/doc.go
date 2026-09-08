package model

import (
	"time"

	"github.com/google/uuid"
)

type Doc struct {
	ID        uuid.UUID
	OwnerID   uuid.UUID
	Name      string
	IsFile    bool
	IsPublic  bool
	MimeType  string
	JSONData  map[any]any
	FilePath  string
	CreatedAt time.Time
}

func NewDoc(
	ownerID uuid.UUID,
	name string,
	isFile, isPublic bool,
	mimeType string,
	jsonData map[any]any,
	filePath string,
) *Doc {
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
	}
}
