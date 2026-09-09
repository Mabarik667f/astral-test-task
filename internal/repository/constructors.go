package repository

import (
	"github.com/Mabarik667f/fsserver/internal/repository/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewUserRepo(db *pgxpool.Pool) UserRepository {
	return user.NewRepository(db)
}
