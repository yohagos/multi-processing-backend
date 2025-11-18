package db

import (
	"context"
	"errors"
	"fmt"
	"multi-processing-backend/internal/core"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DepartmentRepository struct {
	pool *pgxpool.Pool
}

func NewDepartmentRepository(pool *pgxpool.Pool) *DepartmentRepository {
	return &DepartmentRepository{pool: pool}
}

func (r *DepartmentRepository) List(
	ctx context.Context,
	page, limit int,
) ([]core.Departments, int64, error) {
	offset := (page - 1) * limit

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM departments").Scan(&total); err != nil {
		return nil, 0, err
	}
	
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM departments
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	deps, err := pgx.CollectRows(rows, pgx.RowToStructByPos[core.Departments])
	if err != nil {
		return nil, 0, err
	}

	return deps, total, nil
}

func (r *DepartmentRepository) Create(
	ctx context.Context,
	d core.Departments,
) (core.Departments, error) {
	query := `
		INSERT INTO departments (name, description)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query, d.Name, d.Description).Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return core.Departments{}, err
	}

	return d, nil
}

func (r *DepartmentRepository) Get(
	ctx context.Context,
	id string,
) (core.Departments, error) {
	var d core.Departments
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM departments
		WHERE id = $1
	`, id).Scan(&d.ID, &d.Name, &d.Description, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return core.Departments{}, err
	}
	return d, nil
}

func (r *DepartmentRepository) Update(
	ctx context.Context,
	id string,
	update core.DepartmentUpdate,
) (core.Departments, error) {
	var d core.Departments
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM departments
		WHERE id = $1
	`, id).Scan(&d.ID, &d.Name, &d.Description, &d.CreatedAt, &d.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return core.Departments{}, fmt.Errorf("department not found")
		}
		return core.Departments{}, err
	}

	if update.Description != nil {
		d.Description = *update.Description
	}
	if update.Name != nil {
		d.Name = *update.Name
	}

	_, err = r.pool.Exec(ctx, `
		UPDATE departments
		SET name = $1, description = $2
		WHERE id = $3
	`, d.Name, d.Description, d.ID)

	if err != nil {
		return core.Departments{}, nil
	}

	return d, nil
}

func (r *DepartmentRepository) Delete(
	ctx context.Context,
	id string,
) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM departments WHERE id = $1`, id)
	return err
}
