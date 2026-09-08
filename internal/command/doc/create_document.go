package doccmd

import "github.com/google/uuid"

type CreateDocumentCmd struct {
	OwnerID  uuid.UUID
	Name     string
	IsFile   bool
	IsPublic bool
	MimeType string
	Grant    []string
	JSONData map[any]any
	File     []byte
}
