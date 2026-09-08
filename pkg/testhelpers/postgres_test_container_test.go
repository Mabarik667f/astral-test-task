package testhelpers

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/stdlib"
)

func TestPostgresContainer(t *testing.T) {
	ctx := context.Background()
	testDb := SetupTestPostgres(ctx, t)

	testDb.Migrate(ctx, t, "migrations")
	defer testDb.Close(t)

	conn := stdlib.OpenDBFromPool(testDb.Pool)
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		t.Fatalf("Error ping: %v", err)
	}
}
