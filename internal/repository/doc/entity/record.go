package entity

import (
	"time"

	"github.com/google/uuid"
)

type Doc struct {
	ID       uuid.UUID      `db:"id"`
	OwnerID  uuid.UUID      `db:"owner_id"`
	Name     string         `db:"name"`
	File     bool           `db:"file"`
	Public   bool           `db:"public"`
	Mime     string         `db:"mime"`
	JSON     map[string]any `db:"json"`
	FilePath string         `db:"file_path"`
	Created  time.Time      `db:"created"`
}

type DocRead struct {
	ID        uuid.UUID      `db:"id"`
	OwnerID   uuid.UUID      `db:"owner_id"`
	Name      string         `db:"name"`
	IsFile    bool           `db:"file"`
	IsPublic  bool           `db:"public"`
	MimeType  string         `db:"mime"`
	JSONData  map[string]any `db:"json"`
	FilePath  string         `db:"file_path"`
	CreatedAt time.Time      `db:"created"`
	Grants    []string       `db:"grants"`
}
