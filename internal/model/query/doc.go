package query

import (
	"io"
	"time"

	"github.com/google/uuid"
)

type FullDocReadModel struct {
	ID        uuid.UUID
	OwnerID   uuid.UUID
	Name      string
	IsFile    bool
	IsPublic  bool
	MimeType  string
	JSONData  map[string]any
	FilePath  string
	CreatedAt time.Time
	Grants    []string
	File      io.ReadCloser
}

type DocReadModel struct {
	ID        uuid.UUID
	OwnerID   uuid.UUID
	Name      string
	IsFile    bool
	IsPublic  bool
	MimeType  string
	JSONData  map[string]any
	FilePath  string
	CreatedAt time.Time
	Grants    []string
}
