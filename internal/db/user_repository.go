package db

import (
	"context"
	"database/sql"
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
	searchName, departmentName string,
) ([]core.UserWithDetails, int64, error) {
	offset := (page - 1) * limit

	whereClause := "WHERE 1=1"
	params := []interface{}{}
	paramCount := 0

	if searchName != "" {
		paramCount++
		whereClause += fmt.Sprintf(" AND (u.first_name ILIKE $%d OR u.last_name ILIKE $%d)", paramCount, paramCount)
		params = append(params, "%"+searchName+"%")
	}

	if departmentName != "" {
		paramCount++
		whereClause += fmt.Sprintf(" AND d.name ILIKE $%d", paramCount)
		params = append(params, "%"+departmentName+"%")
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM users u
		LEFT JOIN departments d ON u.department_id = d.id
		%s
	`, whereClause)

	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, params...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT
			u.id,
			u.email,
			u.first_name,
			u.last_name,
			u.department_id,
			u.position_id,
			u.hire_date,
			u.phone,
			u.date_of_birth,
			u.created_at,
			u.updated_at,

			COALESCE(d.id::text, '00000000-0000-0000-0000-000000000000') AS dept_id,
			COALESCE(d.name, '') AS dept_name,
			COALESCE(d.description, '') AS dept_description,
			COALESCE(d.created_at, NOW()) AS dept_created_at,
			COALESCE(d.updated_at, NOW()) AS dept_updated_at,

			COALESCE(p.id::text, '00000000-0000-0000-0000-000000000000') AS pos_id,
			COALESCE(p.title, '') AS pos_title,
			COALESCE(p.level::int, 0) AS pos_level,
			COALESCE(p.department_id::text, '00000000-0000-0000-0000-000000000000') AS pos_dept_id,
			COALESCE(p.created_at, NOW()) AS pos_created_at,
			COALESCE(p.updated_at, NOW()) AS pos_updated_at,

			COALESCE(a.id::text, '') AS a_id,
			COALESCE(a.user_id::text, '') AS a_user_id,
			COALESCE(a.street, '') AS a_street,
			COALESCE(a.city, '') AS a_city,
			COALESCE(a.zip_code, '') AS a_zip_code,
			COALESCE(a.country, '') AS a_country,
			COALESCE(a.is_primary, false) AS a_is_primary,
			COALESCE(a.created_at, NOW()) AS a_created_at,
			COALESCE(a.updated_at, NOW()) AS a_updated_at,

			COALESCE((
				SELECT JSON_AGG(
					JSON_BUILD_OBJECT(
						'id', s.id::text,
						'name', s.name,
						'category', s.category,
						'created_at', s.created_at,
						'updated_at', s.updated_at
					)
				)
				FROM user_skills us
				JOIN skills s ON us.skill_id = s.id
				WHERE us.user_id = u.id
			), '[]'::json) AS skills

		FROM users u
		LEFT JOIN departments d ON u.department_id = d.id
		LEFT JOIN positions p ON u.position_id = p.id
		LEFT JOIN LATERAL (
			SELECT *
			FROM addresses a
			WHERE a.user_id = u.id
			LIMIT 1
		) a ON true

		%s

		ORDER BY u.created_at DESC
		LIMIT $%d OFFSET $%d;
    `, whereClause, paramCount+1, paramCount+2)

	params = append(params, limit, offset)

	rows, err := r.pool.Query(ctx, query, params...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []core.UserWithDetails
	for rows.Next() {
		var user core.UserWithDetails
		var skillsJSON []byte

		var addID, addUserID, addStreet, addCity, addZipCode, addCountry sql.NullString
		var addIsPrimary sql.NullBool
		var addCreatedAt, addUpdatedAt sql.NullTime

		user.Departments = &core.Departments{}
		user.Position = &core.Position{}
		user.Address = &core.Address{}

		err := rows.Scan(
			&user.ID, &user.Email, &user.FirstName, &user.LastName,
			&user.DepartmentID, &user.PositionID, &user.HireDate,
			&user.Phone, &user.DateOfBirth, &user.CreatedAt, &user.UpdatedAt,

			&user.Departments.ID, &user.Departments.Name, &user.Departments.Description,
			&user.Departments.CreatedAt, &user.Departments.UpdatedAt,

			&user.Position.ID, &user.Position.Title, &user.Position.Level, &user.Position.DepartmentID,
			&user.Position.CreatedAt, &user.Position.UpdatedAt,

			&addID, &addUserID, &addStreet, &addCity, &addZipCode,
			&addCountry, &addIsPrimary, &addCreatedAt, &addUpdatedAt,

			&skillsJSON,
		)
		if err != nil {
			slog.Warn("ListWithDetails | Error occurred within Scan()", "error", err.Error())
			return nil, 0, err
		}

		if addID.Valid {
			user.Address.ID = addID.String
			user.Address.UserID = addUserID.String
			user.Address.Street = addStreet.String
			user.Address.City = addCity.String
			user.Address.ZipCode = addZipCode.String
			user.Address.Country = addCountry.String
			user.Address.IsPrimary = addIsPrimary.Bool
			user.Address.CreatedAt = addCreatedAt.Time
			user.Address.UpdatedAt = addUpdatedAt.Time
		} else {
			user.Address = nil
		}

		if len(skillsJSON) > 0 {
			if err := json.Unmarshal(skillsJSON, &user.Skill); err != nil {
				slog.Warn("ListWithDetails | Error occurred while json Unmarshal", "error", err.Error())
				return nil, 0, err
			}
		} else {
			user.Skill = []core.Skill{}
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
) (core.UserWithDetails, error) {
	var user core.UserWithDetails
	var skillsJSON []byte

	var addID, addUserID, addStreet, addCity, addZipCode, addCountry sql.NullString
	var addIsPrimary sql.NullBool
	var addCreatedAt, addUpdatedAt sql.NullTime

	user.Departments = &core.Departments{}
	user.Position = &core.Position{}
	user.Address = &core.Address{}
	err := r.pool.QueryRow(ctx, `
		SELECT
			u.id,
			u.email,
			u.first_name,
			u.last_name,
			u.department_id,
			u.position_id,
			u.hire_date,
			u.phone,
			u.date_of_birth,
			u.created_at,
			u.updated_at,

			COALESCE(d.id::text, '00000000-0000-0000-0000-000000000000') AS dept_id,
			COALESCE(d.name, '') AS dept_name,
			COALESCE(d.description, '') AS dept_description,
			COALESCE(d.created_at, NOW()) AS dept_created_at,
			COALESCE(d.updated_at, NOW()) AS dept_updated_at,

			COALESCE(p.id::text, '00000000-0000-0000-0000-000000000000') AS pos_id,
			COALESCE(p.title, '') AS pos_title,
			COALESCE(p.level::int, 0) AS pos_level,
			COALESCE(p.department_id::text, '00000000-0000-0000-0000-000000000000') AS pos_dept_id,
			COALESCE(p.created_at, NOW()) AS pos_created_at,
			COALESCE(p.updated_at, NOW()) AS pos_updated_at,

			COALESCE(a.id::text, '') AS a_id,
			COALESCE(a.user_id::text, '') AS a_user_id,
			COALESCE(a.street, '') AS a_street,
			COALESCE(a.city, '') AS a_city,
			COALESCE(a.zip_code, '') AS a_zip_code,
			COALESCE(a.country, '') AS a_country,
			COALESCE(a.is_primary, false) AS a_is_primary,
			COALESCE(a.created_at, NOW()) AS a_created_at,
			COALESCE(a.updated_at, NOW()) AS a_updated_at,

			COALESCE((
				SELECT JSON_AGG(
					JSON_BUILD_OBJECT(
						'id', s.id::text,
						'name', s.name,
						'category', s.category,
						'created_at', s.created_at,
						'updated_at', s.updated_at
					)
				)
				FROM user_skills us
				JOIN skills s ON us.skill_id = s.id
				WHERE us.user_id = u.id
			), '[]'::json) AS skills

		FROM users u
		LEFT JOIN departments d ON u.department_id = d.id
		LEFT JOIN positions p ON u.position_id = p.id
		LEFT JOIN LATERAL (
			SELECT *
			FROM addresses a
			WHERE a.user_id = u.id
			LIMIT 1
		) a ON true
		WHERE u.id = $1
	`, id).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName,
		&user.DepartmentID, &user.PositionID, &user.HireDate,
		&user.Phone, &user.DateOfBirth, &user.CreatedAt, &user.UpdatedAt,

		&user.Departments.ID, &user.Departments.Name, &user.Departments.Description,
		&user.Departments.CreatedAt, &user.Departments.UpdatedAt,

		&user.Position.ID, &user.Position.Title, &user.Position.Level, &user.Position.DepartmentID,
		&user.Position.CreatedAt, &user.Position.UpdatedAt,

		&addID, &addUserID, &addStreet, &addCity, &addZipCode,
		&addCountry, &addIsPrimary, &addCreatedAt, &addUpdatedAt,

		&skillsJSON,
	)
	if err != nil {
		return core.UserWithDetails{}, err
	}

	if addID.Valid {
		user.Address.ID = addID.String
		user.Address.UserID = addUserID.String
		user.Address.Street = addStreet.String
		user.Address.City = addCity.String
		user.Address.ZipCode = addZipCode.String
		user.Address.Country = addCountry.String
		user.Address.IsPrimary = addIsPrimary.Bool
		user.Address.CreatedAt = addCreatedAt.Time
		user.Address.UpdatedAt = addUpdatedAt.Time
	} else {
		user.Address = nil
	}

	if len(skillsJSON) > 0 {
		if err := json.Unmarshal(skillsJSON, &user.Skill); err != nil {
			slog.Warn("ListWithDetails | Error occurred while json Unmarshal", "error", err.Error())
			return core.UserWithDetails{}, err
		}
	} else {
		user.Skill = []core.Skill{}
	}

	return user, nil
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
