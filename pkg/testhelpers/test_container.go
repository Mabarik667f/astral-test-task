package testhelpers

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/pressly/goose/v3"
)

type TestContainer interface {
	Close(t *testing.T)
}

func Must(t *testing.T, err error, msg string) {
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

func MigrateContainerSchema(
	ctx context.Context,
	t *testing.T,
	db *sql.DB,
	migrationsPath string,
) {
	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		db,
		os.DirFS(migrationsPath),
	)
	if err != nil {
		t.Fatalf("provider create error: %s", err)
	}

	if _, err := provider.Up(ctx); err != nil {
		t.Fatalf("failed to up migrations: %s", err)
	}
}

func MigrationsPath() string {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)

	projectRoot := filepath.Join(basepath, "../..")
	migrationsPath := filepath.Join(projectRoot, "migrations")
	return migrationsPath
}
