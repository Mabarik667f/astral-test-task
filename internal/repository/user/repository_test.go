package user

import (
	"context"
	"testing"

	"github.com/Mabarik667f/fsserver/internal/errs"
	"github.com/Mabarik667f/fsserver/internal/model"
	"github.com/Mabarik667f/fsserver/internal/repository/user/entity"
	"github.com/Mabarik667f/fsserver/pkg/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		testDB := testhelpers.SetupTestPostgres(ctx, t)
		testDB.Migrate(ctx, t, testhelpers.MigrationsPath())
		defer testDB.Close(t)

		user, err := model.NewUser(model.PasswordHash("hash"), "ValidLogin123")
		require.NoError(t, err)
		repo := NewRepository(testDB.Pool)

		_, err = repo.Create(ctx, *user)
		require.NoError(t, err)

		var res entity.UserWithPasswordHash

		query := `SELECT * FROM users WHERE id = $1`
		err = testDB.Pool.QueryRow(ctx, query, user.ID).Scan(&res.ID, &res.Login, &res.PasswordHash)

		require.NoError(t, err)
		assert.Equal(t, user.ID, res.ID)
		assert.Equal(t, user.Login, res.Login)
		assert.Equal(t, string(user.Password), res.PasswordHash)
	})
	t.Run("already exists", func(t *testing.T) {
		ctx := context.Background()
		testDB := testhelpers.SetupTestPostgres(ctx, t)
		testDB.Migrate(ctx, t, testhelpers.MigrationsPath())
		defer testDB.Close(t)

		user, err := model.NewUser(model.PasswordHash("hash"), "ValidLogin123")
		require.NoError(t, err)
		repo := NewRepository(testDB.Pool)

		_, err = repo.Create(ctx, *user)
		require.NoError(t, err)

		_, err = repo.Create(ctx, *user)
		require.Error(t, err)
		require.ErrorIs(t, err, errs.ErrUserUnique)
	})
}

func TestRepository_GetByLoginWithPasswordHash(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		testDB := testhelpers.SetupTestPostgres(ctx, t)
		testDB.Migrate(ctx, t, testhelpers.MigrationsPath())
		defer testDB.Close(t)

		user, err := model.NewUser(model.PasswordHash("hash"), "ValidLogin123")
		require.NoError(t, err)
		repo := NewRepository(testDB.Pool)

		repo.Create(ctx, *user)

		res, err := repo.GetByLoginWithPasswordHash(ctx, user.Login)
		require.NoError(t, err)
		assert.Equal(t, user.ID, res.ID)
		assert.Equal(t, user.Login, res.Login)
		assert.Equal(t, user.Password, res.Password)
	})
	t.Run("not found", func(t *testing.T) {
		ctx := context.Background()
		testDB := testhelpers.SetupTestPostgres(ctx, t)
		testDB.Migrate(ctx, t, testhelpers.MigrationsPath())
		defer testDB.Close(t)

		repo := NewRepository(testDB.Pool)

		_, err := repo.GetByLoginWithPasswordHash(ctx, "someLogin")
		require.Error(t, err)
		require.ErrorIs(t, err, errs.ErrUserNotFound)
	})
}
