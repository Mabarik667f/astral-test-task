package doc

import (
	"context"
	"fmt"

	"github.com/Mabarik667f/fsserver/internal/model"
	"github.com/Mabarik667f/fsserver/internal/repository/doc/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Create(ctx context.Context, doc model.Doc) (model.Doc, error) {
	const query = `
		INSERT INTO docs (
			id,
			owner_id,
			name,
			file,
			public,
			mime,
			json,
			file_path,
			created
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9
		)
		RETURNING
			id,
			owner_id,
			name,
			file,
			public,
			mime,
			json,
			file_path,
			created
	`

	var result entity.Doc

	err := r.db.QueryRow(
		ctx,
		query,
		doc.ID,
		doc.OwnerID,
		doc.Name,
		doc.IsFile,
		doc.IsPublic,
		doc.MimeType,
		doc.JSONData,
		doc.FilePath,
		doc.CreatedAt,
	).Scan(
		&result.ID,
		&result.OwnerID,
		&result.Name,
		&result.File,
		&result.Public,
		&result.Mime,
		&result.JSON,
		&result.FilePath,
		&result.Created,
	)
	if err != nil {
		return model.Doc{}, fmt.Errorf("create document: %v", err)
	}

	return entity.ToDoc(result), nil
}

func (r *repository) CreateGrants(
	ctx context.Context,
	docID uuid.UUID,
	userIDs uuid.UUIDs,
) error {
	if len(userIDs) == 0 {
		return nil
	}

	const query = `
		INSERT INTO doc_grants (doc_id, user_id)
		SELECT $1, unnest($2::uuid[])
		ON CONFLICT (doc_id, user_id) DO NOTHING
	`

	_, err := r.db.Exec(ctx, query, docID, userIDs)
	if err != nil {
		return fmt.Errorf("create document grants: %w", err)
	}

	return nil
}
