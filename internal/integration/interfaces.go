package integration

import "io"

type FileStorage interface {
	Save(fileID string, name string, data io.Reader) (string, error)
	Read(fileID, name string) (io.ReadCloser, error)
	Delete(fileID string) error
}
