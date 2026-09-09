package doc

import (
	"context"
	"testing"
	"time"

	"github.com/Mabarik667f/fsserver/internal/model"
	"github.com/Mabarik667f/fsserver/pkg/testhelpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		testDB := testhelpers.SetupTestPostgres(ctx, t)
		testDB.Migrate(ctx, t, testhelpers.MigrationsPath())
		defer testDB.Close(t)

		repo := NewRepository(testDB.Pool)

		user, err := model.NewUser(
			model.PasswordHash("hash1"),
			"ValidLogin123",
		)

		query := `INSERT INTO users (id, login, password_hash) VALUES ($1, $2, $3)`
		_, err = testDB.Pool.Exec(ctx, query, user.ID, user.Login, user.Password)
		require.NoError(t, err)

		doc := model.Doc{
			ID:       uuid.New(),
			OwnerID:  user.ID,
			Name:     "photo.jpg",
			IsFile:   true,
			IsPublic: false,
			MimeType: "image/jpeg",
			JSONData: map[string]any{
				"description": "test photo",
				"width":       float64(1920),
			},
			FilePath:  "/storage/photo.jpg",
			CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
		}

		res, err := repo.Create(ctx, doc)

		require.NoError(t, err)

		assert.Equal(t, doc.ID, res.ID)
		assert.Equal(t, doc.OwnerID, res.OwnerID)
		assert.Equal(t, doc.Name, res.Name)
		assert.Equal(t, doc.IsFile, res.IsFile)
		assert.Equal(t, doc.IsPublic, res.IsPublic)
		assert.Equal(t, doc.MimeType, res.MimeType)
		assert.Equal(t, doc.JSONData, res.JSONData)
		assert.Equal(t, doc.FilePath, res.FilePath)
		assert.Equal(t, doc.CreatedAt, res.CreatedAt)
	})
}

func TestRepository_CreateGrants(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		testDB := testhelpers.SetupTestPostgres(ctx, t)
		testDB.Migrate(ctx, t, testhelpers.MigrationsPath())
		defer testDB.Close(t)

		repo := NewRepository(testDB.Pool)

		user1, err := model.NewUser(
			model.PasswordHash("hash1"),
			"ValidLogin123",
		)
		require.NoError(t, err)

		query := `INSERT INTO users (id, login, password_hash) VALUES ($1, $2, $3)`
		_, err = testDB.Pool.Exec(ctx, query, user1.ID, user1.Login, user1.Password)
		require.NoError(t, err)

		docID := uuid.New()

		_, err = repo.Create(ctx, model.Doc{
			ID:        docID,
			OwnerID:   user1.ID,
			Name:      "test.txt",
			IsFile:    false,
			IsPublic:  false,
			MimeType:  "text/plain",
			FilePath:  "",
			CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
		})
		require.NoError(t, err)

		err = repo.CreateGrants(ctx, docID, uuid.UUIDs{
			user1.ID,
		})
		require.NoError(t, err)

		rows, err := testDB.Pool.Query(ctx, `
			SELECT user_id
			FROM doc_grants
			WHERE doc_id = $1
		`, docID)
		require.NoError(t, err)
		defer rows.Close()

		var userIDs []uuid.UUID

		for rows.Next() {
			var userID uuid.UUID

			err := rows.Scan(&userID)
			require.NoError(t, err)

			userIDs = append(userIDs, userID)
		}

		require.NoError(t, rows.Err())
		require.Len(t, userIDs, 1)

		assert.ElementsMatch(
			t,
			[]uuid.UUID{user1.ID},
			userIDs,
		)
	})
}
