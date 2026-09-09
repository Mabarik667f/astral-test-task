package dto

type UploadFileResponse struct {
	JSON     map[string]any `json:"json,omitempty"`
	FileName string         `json:"file"`
}
