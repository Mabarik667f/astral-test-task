package dto

type MetaData struct {
	Name     string   `json:"name"   validate:"required"`
	IsFile   bool     `json:"file"`
	IsPublic bool     `json:"public"`
	MimeType string   `json:"mime"   validate:"required"`
	Grant    []string `json:"grant"  validate:"required"`
}

type GetDocumentsListRequest struct {
	Login string `schema:"login" validate:"omitempty"`
	Key   string `schema:"key"   validate:"required"`
	Value any    `schema:"value" validate:"required"`
	Limit int    `schema:"limit" validate:"required,gt=0"`
}
