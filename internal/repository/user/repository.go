package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/Mabarik667f/fsserver/internal/errs"
	"github.com/Mabarik667f/fsserver/internal/model"
	"github.com/Mabarik667f/fsserver/internal/repository/user/entity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, user model.User) (model.User, error) {
	query := `INSERT INTO users (id, login, password_hash) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, user.ID, user.Login, user.Password)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.User{}, errs.ErrUserUnique
		}

		return model.User{}, err
	}
	return user, nil
}

func (r *repository) GetByLoginWithPasswordHash(
	ctx context.Context,
	login string,
) (model.User, error) {
	var user entity.UserWithPasswordHash

	query := `SELECT * FROM users WHERE login = $1`
	err := r.db.QueryRow(ctx, query, login).Scan(&user.ID, &user.Login, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, errs.ErrUserNotFound
		}
		return model.User{}, fmt.Errorf("retrieving from db: %w", err)
	}

	return entity.ToUserWithPasswordHash(user), nil
}

func (r *repository) GetUsersByLogins(ctx context.Context, logins []string) ([]model.User, error) {
	query := `
		SELECT id, login
		FROM users
		WHERE login = ANY($1)
	`

	rows, err := r.db.Query(ctx, query, logins)
	if err != nil {
		return nil, fmt.Errorf("query users by logins: %w", err)
	}
	defer rows.Close()

	entities, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.User])
	if err != nil {
		return nil, fmt.Errorf("scan users: %w", err)
	}

	users := make([]model.User, 0, len(entities))
	for _, user := range entities {
		users = append(users, entity.ToUser(user))
	}

	return users, nil
}
