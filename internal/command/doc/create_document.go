package doccmd

import (
	"io"

	"github.com/google/uuid"
)

type CreateDocumentCmd struct {
	OwnerID  uuid.UUID
	Name     string
	IsFile   bool
	IsPublic bool
	MimeType string
	Grant    []string
	JSONData map[string]any
	File     io.Reader
}
