package testhelpers

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type PostgresTestContainer struct {
	Container *postgres.PostgresContainer
	Pool      *pgxpool.Pool
}

func SetupTestPostgres(ctx context.Context, t *testing.T) *PostgresTestContainer {
	if testing.Short() {
		t.Skip("postgres testcontainer is used")
	}

	postgresContainer, err := NewPostgresTestContainer(t, ctx)
	Must(t, err, "failed to start container")

	endpoint, err := postgresContainer.ConnectionString(ctx)
	Must(t, err, "failed to get connection string")

	pool, err := pgxpool.New(context.Background(), endpoint)
	Must(t, err, "failed to set connection")

	return &PostgresTestContainer{
		Container: postgresContainer,
		Pool:      pool,
	}
}

func (p *PostgresTestContainer) Close(t *testing.T) {
	ctx := context.Background()

	if p.Container != nil {
		err := p.Container.Terminate(ctx)
		Must(t, err, "failed to terminate container")
	}
}

func (p *PostgresTestContainer) Migrate(ctx context.Context, t *testing.T, path string) {
	t.Helper()
	conn := stdlib.OpenDBFromPool(p.Pool)
	MigrateContainerSchema(ctx, t, conn, path)
}

func (p *PostgresTestContainer) URI(ctx context.Context, t *testing.T) string {
	conn, err := p.Container.ConnectionString(ctx)
	Must(t, err, "failed to get connection string")
	return conn + "sslmode=disable"
}

func NewPostgresTestContainer(
	t *testing.T,
	ctx context.Context,
) (*postgres.PostgresContainer, error) {
	opts := postgres.BasicWaitStrategies()
	postgresContainer, err := postgres.Run(ctx, "postgres:18-alpine", opts)
	if err != nil {
		return nil, err
	}
	return postgresContainer, nil
}
