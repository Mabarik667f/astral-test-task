package filestorage

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
)

type storage struct {
	root string
}

func NewStorage(root string) *storage {
	return &storage{root: root}
}

func (s *storage) Save(fileID, name string, data io.Reader) (string, error) {
	dir := filepath.Join(s.root, fileID)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	pathFile := filepath.Join(dir, name)
	f, err := os.Create(pathFile)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, data); err != nil {
		return "", err
	}

	return pathFile, nil
}

func (s *storage) Read(fileID, name string) (io.ReadCloser, error) {
	dir := filepath.Join(s.root, fileID)
	fs := os.DirFS(dir)
	file, err := fs.Open(name)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (s *storage) Delete(fileID string) error {
	dir := path.Join(s.root, fileID)
	err := os.RemoveAll(dir)
	if err != nil {
		return fmt.Errorf("deleting directory: %w", err)
	}

	return nil
}
