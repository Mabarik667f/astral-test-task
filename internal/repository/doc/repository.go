package doc

import (
	"context"
	"errors"
	"fmt"

	doccmd "github.com/Mabarik667f/fsserver/internal/command/doc"
	"github.com/Mabarik667f/fsserver/internal/errs"
	"github.com/Mabarik667f/fsserver/internal/model"
	"github.com/Mabarik667f/fsserver/internal/model/query"
	"github.com/Mabarik667f/fsserver/internal/repository/doc/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	const queryString = `
		INSERT INTO docs (
			id, owner_id, name, file, public, mime, json, file_path, created
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, owner_id, name, file, public, mime, json, file_path, created
	`

	var result entity.Doc

	err := r.db.QueryRow(
		ctx,
		queryString,
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

	const queryString = `
		INSERT INTO doc_grants (doc_id, user_id)
		SELECT $1, unnest($2::uuid[])
		ON CONFLICT (doc_id, user_id) DO NOTHING
	`

	_, err := r.db.Exec(ctx, queryString, docID, userIDs)
	if err != nil {
		return fmt.Errorf("create document grants: %w", err)
	}

	return nil
}

func (r *repository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	const queryString = `DELETE FROM docs WHERE id = $1`

	if _, err := r.db.Exec(ctx, queryString, id); err != nil {
		return fmt.Errorf("delete document: %w", err)
	}

	return nil
}

func (r *repository) Get(
	ctx context.Context,
	cmd doccmd.GetDocumentsListCmd,
	userID uuid.UUID,
) ([]query.DocReadModel, error) {
	allowedKeys := map[string]string{
		"owner_id": "owner_id",
		"name":     "name",
		"file":     "file",
		"public":   "public",
		"mime":     "mime",
		"created":  "created",
	}

	column, ok := allowedKeys[cmd.Key]
	if !ok {
		return nil, errs.ErrInvalidFilter
	}

	queryString := fmt.Sprintf(`
		SELECT
			d.id,
			d.owner_id,
			d.name,
			d.file,
			d.public,
			d.mime,
			d.json,
			d.file_path,
			d.created,
			COALESCE(array_agg(u.login) FILTER (WHERE u.login IS NOT NULL), '{}') AS grants
		FROM docs d 
		LEFT JOIN doc_grants dg 
			ON dg.doc_id = d.id 
		LEFT JOIN users u 
			ON u.id = dg.user_id 
		WHERE d.%s = $1 AND (
			($2 <> '' AND EXISTS(
				SELECT 1
				FROM doc_grants dg2 
				JOIN users u2 ON u2.id = dg2.user_id 
				WHERE dg2.doc_id = d.id AND u2.login = $2
			))
			OR 
			($2 = '' AND d.owner_id = $3)
		)
		GROUP BY d.id, d.owner_id, d.name, d.file, d.public, d.mime, d.json, d.file_path, d.created
		ORDER BY d.name, d.created
		LIMIT $4
	`, column)

	rows, err := r.db.Query(ctx, queryString, cmd.Value, cmd.Login, userID, cmd.Limit)
	if err != nil {
		return nil, fmt.Errorf("query docs: %w", err)
	}
	defer rows.Close()

	docs, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (entity.DocRead, error) {
		var doc entity.DocRead

		err := row.Scan(
			&doc.ID,
			&doc.OwnerID,
			&doc.Name,
			&doc.IsFile,
			&doc.IsPublic,
			&doc.MimeType,
			&doc.JSONData,
			&doc.FilePath,
			&doc.CreatedAt,
			&doc.Grants,
		)
		if err != nil {
			return entity.DocRead{}, err
		}

		return doc, nil
	})
	if err != nil {
		return nil, fmt.Errorf("collect docs: %w", err)
	}

	result := make([]query.DocReadModel, len(docs))
	for i := range docs {
		result[i] = entity.ToDocReadModel(docs[i])
	}

	return result, nil
}

func (r *repository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (query.DocReadModel, error) {
	const queryString = `
		SELECT
			d.id,
			d.owner_id,
			d.name,
			d.file,
			d.public,
			d.mime,
			d.json,
			d.file_path,
			d.created,
			COALESCE(array_agg(u.login) FILTER (WHERE u.login IS NOT NULL), '{}') AS grants
		FROM docs d
		LEFT JOIN doc_grants dg
			ON dg.doc_id = d.id
		LEFT JOIN users u
			ON u.id = dg.user_id
		WHERE d.id = $1
		GROUP BY d.id, d.owner_id, d.name, d.file, d.public, d.mime, d.json, d.file_path, d.created
	`

	var res entity.DocRead

	err := r.db.QueryRow(ctx, queryString, id).Scan(
		&res.ID,
		&res.OwnerID,
		&res.Name,
		&res.IsFile,
		&res.IsPublic,
		&res.MimeType,
		&res.JSONData,
		&res.FilePath,
		&res.CreatedAt,
		&res.Grants,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return query.DocReadModel{}, errs.ErrDocNotFound
		}

		return query.DocReadModel{}, fmt.Errorf("get document: %w", err)
	}

	return entity.ToDocReadModel(res), nil
}
