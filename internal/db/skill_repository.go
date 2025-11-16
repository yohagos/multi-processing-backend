package db

import (
	"context"
	
	"errors"
	"fmt"
	

	"multi-processing-backend/internal/core"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	
)

type SkillRepository struct {
	pool *pgxpool.Pool
}

func NewSkillRepository(pool *pgxpool.Pool) *SkillRepository {
	return &SkillRepository{pool: pool}
}

func (r *SkillRepository) List(
	ctx context.Context,
	page, limit int,
) ([]core.Skill, int64, error) {
	offset := (page - 1) * limit

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM skills").Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, name, category, created_at, updated_at
		FROM skills
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	skills, err := pgx.CollectRows(rows, pgx.RowToStructByPos[core.Skill])
	if err != nil {
		return nil, 0, err
	}

	return skills, total, nil
}

func (r *SkillRepository) Create(
	ctx context.Context,
	s core.Skill,
) (core.Skill, error) {
	query := `
		INSERT INTO skills (name, category)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx, query, s.Name, s.Category,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return core.Skill{}, err
	}

	return s, nil
}

func (r *SkillRepository) Get(
	ctx context.Context, 
	id string,
) (core.Skill, error) {
	var s core.Skill
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, category, created_at, updated_at
		FROM skills 
		WHERE id = $1
	`, id).Scan(
		&s.ID, &s.Name, &s.Category, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return core.Skill{}, err
	}
	return s, nil
}

func (r *SkillRepository) Update(
	ctx context.Context,
	id string,
	update core.SkillUpdate,
) (core.Skill, error) {
	var s core.Skill

	err := r.pool.QueryRow(ctx, `
		SELECT id, name, category, created_at, updated_at
		FROM skills 
		WHERE id = $1
	`, id).Scan(
		&s.ID, &s.Name, &s.Category, &s.CreatedAt, &s.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return core.Skill{}, fmt.Errorf("user not found")
		}
		return core.Skill{}, err
	}

	_, err = r.pool.Exec(ctx, `
		UPDATE skills
		SET name = $1, category = $2, updated_at = NOW()
		WHERE id = $9
	`, s.Name, s.Category, id)
	if err != nil {
		return core.Skill{}, err
	}

	return s, nil
}

func (r *SkillRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM skills WHERE id = $1`, id)
	return err
}
