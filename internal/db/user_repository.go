package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	offset := (page - 1) * limit

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, email, first_name, last_name, department_id, position_id, hire_date, phone, date_of_birth, created_at, updated_at
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

	return users, total, nil
}

func (r *UserRepository) ListWithDetails(
	ctx context.Context,
	page, limit int,
) ([]core.UserWithDetails, int64, error) {
	offset := (page - 1) * limit

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT 
			u.id, u.email, u.first_name, u.last_name, u.department_id, u.position_id, u.hire_date, 
			u.phone, u.date_of_birth, u.created_at, u.updated_at,
			d.id, d.name, d.description, d.created_at, d.updated_at,
			p.id, p.title, p.level, p.department_id, p.created_at, p.updated_at
		FROM users u
		LEFT JOIN departments d ON u.department_id = d.id
		LEFT JOIN positions p ON u.position_id = p.id
		ORDER BY u.created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []core.UserWithDetails
	for rows.Next() {
		var user core.UserWithDetails
		user.Departments = &core.Departments{}
		user.Position = &core.Position{}

		err := rows.Scan(
			&user.ID, &user.Email, &user.FirstName, &user.LastName,
			&user.DepartmentID, &user.PositionID, &user.HireDate,
			&user.Phone, &user.DateOfBirth, &user.CreatedAt, &user.UpdatedAt,

			&user.Departments.ID, &user.Departments.Name, &user.Departments.Description,
			&user.Departments.CreatedAt, &user.Departments.UpdatedAt,

			&user.Position.ID, &user.Position.Title, &user.Position.Level, &user.Position.DepartmentID,
			&user.Position.CreatedAt, &user.Position.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}
	return users, total, nil
}

func (r *UserRepository) Create(
	ctx context.Context,
	u core.User,
) (core.User, error) {
	query := `
		INSERT INTO users (email, first_name, last_name, department_id, position_id, hire_date, phone, date_of_birth)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query, u.Email, u.FirstName, u.LastName, u.DepartmentID, u.PositionID, u.HireDate, u.Phone, u.DateOfBirth).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return core.User{}, err
	}

	return u, nil
}

func (r *UserRepository) Get(
	ctx context.Context,
	id string,
) (core.User, error) {
	var u core.User
	err := r.pool.QueryRow(ctx, `
		SELECT id, email, first_name, last_name, department_id, position_id, hire_date, phone, date_of_birth, created_at, updated_at
		FROM users 
		WHERE id = $1
	`, id).Scan(
		&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.DepartmentID, &u.PositionID, &u.HireDate, &u.Phone, &u.DateOfBirth, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return core.User{}, err
	}
	return u, nil
}

func (r *UserRepository) Update(
	ctx context.Context,
	id string,
	update core.UserUpdate,
) (core.User, error) {
	var user core.User

	err := r.pool.QueryRow(ctx, `
		SELECT id, email, first_name, last_name, department_id, position_id, hire_date, phone, date_of_birth, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.DepartmentID, &user.PositionID,
		&user.HireDate, &user.Phone, &user.DateOfBirth, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return core.User{}, fmt.Errorf("user not found")
		}
		return core.User{}, err
	}

	if update.Email != nil {
		user.Email = *update.Email
	}
	if update.FirstName != nil {
		user.FirstName = *update.FirstName
	}
	if update.LastName != nil {
		user.LastName = *update.LastName
	}

	_, err = r.pool.Exec(ctx, `
		UPDATE users
		SET email = $1, first_name = $2, last_name = $3, department_id = $4, position_id = $5, hire_date = $6, phone = $7, date_of_birth = $8, updated_at = NOW()
		WHERE id = $9
	`, user.Email, user.FirstName, user.LastName, user.DepartmentID, user.PositionID, user.HireDate, user.Phone, user.DateOfBirth, id)
	if err != nil {
		return core.User{}, err
	}

	return user, nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

func (r *UserRepository) SeedUsersIfEmpty(
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

	return nil
}

func (r *UserRepository) DeleteDevData(ctx context.Context) {
	_, err := r.pool.Exec(ctx, `DROP TABLE users`)
	if err != nil {
		slog.Error("Could not delete Users Table before shutting down", "DatabaseError", err)
		return
	}
	slog.Info("Dropped Users Table successfully")
}
