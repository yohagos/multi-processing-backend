package db

import (
	"context"
	
	"errors"
	"fmt"
	

	"multi-processing-backend/internal/core"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	
)

type PositionRepository struct {
	pool *pgxpool.Pool
}

func NewPositionRepository(pool *pgxpool.Pool) *PositionRepository {
	return &PositionRepository{pool: pool}
}

func (r *PositionRepository) List(
	ctx context.Context,
	page, limit int,
) ([]core.Position, int64, error) {
	offset := (page - 1) * limit

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM positions").Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, title, level, department_id, created_at, updated_at
		FROM positions
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	skills, err := pgx.CollectRows(rows, pgx.RowToStructByPos[core.Position])
	if err != nil {
		return nil, 0, err
	}

	return skills, total, nil
}

func (r *PositionRepository) Create(
	ctx context.Context,
	p core.Position,
) (core.Position, error) {
	query := `
		INSERT INTO positions (title, level, department_id)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx, query, p.Title, p.Level, p.DepartmentID,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return core.Position{}, err
	}

	return p, nil
}

func (r *PositionRepository) Get(
	ctx context.Context, 
	id string,
) (core.Position, error) {
	var p core.Position
	err := r.pool.QueryRow(ctx, `
		SELECT id, title, level, department_id, created_at, updated_at
		FROM positions 
		WHERE id = $1
	`, id).Scan(
		&p.ID, &p.Title, &p.Level, &p.DepartmentID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return core.Position{}, err
	}
	return p, nil
}

func (r *PositionRepository) Update(
	ctx context.Context,
	id string,
	update core.PositionUpdate,
) (core.Position, error) {
	var p core.Position

	err := r.pool.QueryRow(ctx, `
		SELECT id, title, level, department_id, created_at, updated_at
		FROM positions 
		WHERE id = $1
	`, id).Scan(
		&p.ID, &p.Title, &p.Level, &p.DepartmentID, &p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return core.Position{}, fmt.Errorf("user not found")
		}
		return core.Position{}, err
	}

	_, err = r.pool.Exec(ctx, `
		UPDATE positions
		SET title = $1, level = $2, department_id = $3, updated_at = NOW()
		WHERE id = $4
	`, p.Title, p.Level, p.DepartmentID, id)
	if err != nil {
		return core.Position{}, err
	}

	return p, nil
}

func (r *PositionRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM positions WHERE id = $1`, id)
	return err
}
