package db

import (
	"context"
	"encoding/json"
	"os"

	"multi-processing-backend/internal/core"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/exp/slog"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) List(
	ctx context.Context,
	page, limit int,
) ([]core.User, int64, error) {
	slog.Info("List from user_repository", "Database", "get list of users")
	offset := (page - 1) * limit

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, email, first_name, last_name, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	users, err := pgx.CollectRows(rows, pgx.RowToStructByPos[core.User])
	if err != nil {
		return nil, 0, err
	}
	slog.Info("Found Users from user_reposiory", "Database", users)

	return users, total, nil
}

func (r *UserRepository) Create(
	ctx context.Context,
	u core.User,
) (core.User, error) {
	query := `
		INSERT INTO users (email, first_name, last_name)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query, u.Email, u.FirstName, u.LastName).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return core.User{}, err
	}

	return u, nil
}

func (r *UserRepository) SeedIfEmpty(
	ctx context.Context,
	jsonPath string,
) error {
	var count int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return err
	}

	if count > 0 {
		slog.Info("Users already exist, skipping seed")
		return nil
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}

	var users []struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	if err := json.Unmarshal(data, &users); err != nil {
		return err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, u := range users {
		_, err := tx.Exec(ctx, `
			INSERT INTO users (email, first_name, last_name)
			VALUES ($1, $2, $3)
			ON CONFLICT (email) DO NOTHING
		`, u.Email, u.FirstName, u.LastName)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	slog.Info("User seed inserted successfully", "count", len(users))
	return nil
}
