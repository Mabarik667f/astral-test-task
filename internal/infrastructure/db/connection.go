package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(dsn string) (*pgxpool.Pool, error) {
	return pgxpool.New(context.Background(), dsn)
}
