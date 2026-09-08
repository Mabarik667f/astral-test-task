package entity

import "github.com/google/uuid"

type User struct {
	ID    uuid.UUID `db:"id"`
	Login string    `db:"login"`
}

type UserWithPasswordHash struct {
	ID           uuid.UUID `db:"id"`
	Login        string    `db:"login"`
	PasswordHash string    `db:"password_hash"`
}
