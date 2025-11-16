package db

import (
	"context"
	"errors"
	"fmt"
	"multi-processing-backend/internal/core"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SalaryRepository struct {
	pool *pgxpool.Pool
}

func NewSalaryRepository(pool *pgxpool.Pool) *SalaryRepository {
	return &SalaryRepository{pool: pool}
}

func (r *SalaryRepository) List(
	ctx context.Context,
	page, limit int,
) ([]core.Salary, int64, error) {
	offset := (page - 1) * limit

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM salaries").Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, amount, effective_date, created_at, updated_at
		FROM salaries
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	sals, err := pgx.CollectRows(rows, pgx.RowToStructByPos[core.Salary])
	if err != nil {
		return nil, 0, err
	}

	return sals, total, nil
}

func (r *SalaryRepository) Create(
	ctx context.Context,
	s core.Salary,
) (core.Salary, error) {
	query := `
		INSERT INTO salaries (user_id, amount, effective_date)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx, query, s.UserID, s.Amount, s.EffectiveDate,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return core.Salary{}, err
	}

	return s, nil
}

func (r *SalaryRepository) Get(
	ctx context.Context, 
	id string,
) (core.Salary, error) {
	var s core.Salary
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, amount, effective_date, created_at, updated_at
		FROM salaries 
		WHERE id = $1
	`, id).Scan(
		&s.ID, &s.UserID, &s.Amount, &s.EffectiveDate, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return core.Salary{}, err
	}
	return s, nil
}

func (r *SalaryRepository) Update(
	ctx context.Context,
	id string,
	update core.SalaryUpdate,
) (core.Salary, error) {
	var s core.Salary

	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, amount, effective_date, created_date, updated_at
		FROM salaries 
		WHERE id = $1
	`, id).Scan(
		&s.ID, &s.UserID, &s.Amount, &s.EffectiveDate, &s.CreatedAt, &s.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return core.Salary{}, fmt.Errorf("salary not found")
		}
		return core.Salary{}, err
	}

	_, err = r.pool.Exec(ctx, `
		UPDATE salaries
		SET user_id = $1, amount = $2, effective_date = $3, updated_at = NOW()
		WHERE id = $4
	`, s.UserID, s.Amount, s.EffectiveDate, id)
	if err != nil {
		return core.Salary{}, err
	}

	return s, nil
}

func (r *SalaryRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM salaries WHERE id = $1`, id)
	return err
}