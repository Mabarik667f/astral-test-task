package dto

import (
	"time"

	"github.com/google/uuid"
)

type UploadFileResponse struct {
	JSON     map[string]any `json:"json,omitempty"`
	FileName string         `json:"file"`
}

type DocReadModelResponse struct {
	ID        uuid.UUID      `json:"id"`
	OwnerID   uuid.UUID      `json:"owner_id"`
	Name      string         `json:"name"`
	IsFile    bool           `json:"file"`
	IsPublic  bool           `json:"public"`
	MimeType  string         `json:"mime"`
	JSONData  map[string]any `json:"json"`
	CreatedAt time.Time      `json:"created_at"`
	Grants    []string       `json:"grants"`
}
