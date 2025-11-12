package db

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"multi-processing-backend/internal/core"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/exp/slog"
)

var (
	departments []core.Departments
	positions   []core.Position
)

type Seeder struct {
	pool *pgxpool.Pool
}

func NewSeeder(pool *pgxpool.Pool) *Seeder {
	return &Seeder{pool: pool}
}

func (s *Seeder) SeedAll(ctx context.Context, basePath string) error {
	slog.Info("Seeder started")
	seedOrder := []struct {
		tableName string
		jsonFile  string
		seedFunc  func(ctx context.Context, jsonPath string) error
	}{
		{"departments", "generated_departments.json", s.seedDepartments},
		{"positions", "generated_positions.json", s.seedPositions},
		{"users", "generated_users.json", s.seedUsers},
		{"salaries", "generated_salaries.json", s.seedSalaries},
		{"addresses", "generated_addresses.json", s.seedAddresses},
		{"skills", "generated_skills.json", s.seedSkills},
		{"user_skills", "generated_user_skills.json", s.seedUsersSkills},
	}

	for _, seed := range seedOrder {
		jsonPath := filepath.Join(basePath, seed.jsonFile)

		var count int
		err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM "+seed.tableName).Scan(&count)
		if err != nil {
			return fmt.Errorf("checking %s: %w", seed.tableName, err)
		}

		if count > 0 {
			slog.Info("Table already has data, skipping", "table", seed.tableName)
			continue
		}

		slog.Info("Seeding table", "table", seed.tableName)
		if err := seed.seedFunc(ctx, jsonPath); err != nil {
			return fmt.Errorf("seeding %s: %w", seed.tableName, err)
		}
	}
	return nil
}

func (s *Seeder) seedDepartments(ctx context.Context, jsonPath string) error {
	slog.Info("Seeder | Departments started")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}

	var departments []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		CreatedAt   string `json:"created_at"`
	}

	if err := json.Unmarshal(data, &departments); err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, d := range departments {
		_, err := tx.Exec(ctx, `
			INSERT INTO departments (name, description, created_at)
			VALUES ($1, $2, $3)
		`, d.Name, d.Description, d.CreatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *Seeder) seedPositions(ctx context.Context, jsonPath string) error {
	slog.Info("Seeder | Positions started")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}

	deps, err := s.pool.Query(ctx, "SELECT * FROM departments")
	if err != nil {
		slog.Warn("getting departments failed")
		return err
	}
	departments, _ = pgx.CollectRows(deps, pgx.RowToStructByPos[core.Departments])

	var positions []struct {
		Title        string `json:"title"`
		Level        int    `json:"level"`
		DepartmentID string `json:"department_id"`
		CreatedAt    string `json:"created_at"`
	}

	if err := json.Unmarshal(data, &positions); err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, p := range positions {
		dep := getRandomDepartment()
		_, err := tx.Exec(ctx, `
			INSERT INTO positions (id, title, level, department_id, created_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO NOTHING
		`, p.Title, p.Level, dep.ID, p.CreatedAt)
		if err != nil {
			return nil
		}
	}

	return tx.Commit(ctx)
}

func getRandomDepartment() core.Departments {
	return departments[rand.Intn(len(departments))]
}

func getRandomPosition() core.Position {
	return positions[rand.Intn(len(positions))]
}

func (s *Seeder) seedUsers(ctx context.Context, jsonPath string) error {
	slog.Info("Seeder | Users started")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}

	pos, err := s.pool.Query(ctx, "SELECT * FROM positions")
	if err != nil {
		slog.Warn("Positions table is empty")
	}
	positions, _ = pgx.CollectRows(pos, pgx.RowToStructByPos[core.Position])
	slog.Warn("Positions seed", "data", positions)

	var users []struct {
		Email        string `json:"email"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		DepartmentID string `json:"department_id"`
		PositionID   string `json:"position_id"`
		HireDate     string `json:"hire_date"`
		Phone        string `json:"phone"`
		DateOfBirth  string `json:"date_of_birth"`
	}

	if err := json.Unmarshal(data, &users); err != nil {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, u := range users {
		dep := getRandomDepartment()
		posi := getRandomPosition()
		_, err := tx.Exec(ctx, `
			INSERT INTO users (email, first_name, last_name, department_id, position_id, hire_date, phone, date_of_birth)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (id) DO NOTHING
		`, u.Email, u.FirstName, u.LastName, dep.ID,
			posi.ID, u.HireDate, u.Phone, u.DateOfBirth)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *Seeder) seedSalaries(ctx context.Context, jsonPath string) error {
	slog.Info("Seeder | Salaries started")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}

	var salaries []struct {
		UserID        string  `json:"user_id"`
		Amount        float64 `json:"amount"`
		EffectiveDate string  `json:"effective_date"`
		CreatedAt     string  `json:"created_date"`
	}

	if err := json.Unmarshal(data, &salaries); err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, l := range salaries {
		_, err := tx.Exec(ctx, `
			INSERT INTO salaries (id, user_id, amount, effective_date, created_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO NOTHING
		`, l.UserID, l.Amount, l.EffectiveDate, l.CreatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *Seeder) seedAddresses(ctx context.Context, jsonPath string) error {
	slog.Info("Seeder | Addresses started")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}

	var addresses []struct {
		UserID    string `json:"user_id"`
		Street    string `json:"street"`
		City      string `json:"city"`
		ZipCode   string `json:"zip_code"`
		Country   string `json:"country"`
		IsPrimary bool   `json:"is_primary"`
		CreatedAt string `json:"created_at"`
	}

	if err := json.Unmarshal(data, &addresses); err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, a := range addresses {
		_, err := tx.Exec(ctx, `
			INSERT INTO addresses (id, user_id, street, city, zip_code, country, is_primary, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO NOTHING
		`, a.UserID, a.Street, a.City, a.ZipCode, a.Country, a.IsPrimary, a.CreatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *Seeder) seedSkills(ctx context.Context, jsonPath string) error {
	slog.Info("Seeder | Skills started")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}

	var skills []struct {
		Name      string `json:"name"`
		Category  string `json:"category"`
		CreatedAt string `json:"created_at"`
	}

	if err := json.Unmarshal(data, &skills); err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, k := range skills {
		_, err := tx.Exec(ctx, `
			INSERT INTO skills (id, name, category, created_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO NOTHING
		`, k.Name, k.Category, k.CreatedAt)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *Seeder) seedUsersSkills(ctx context.Context, jsonPath string) error {
	slog.Info("Seeder | UserSkills started")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}

	var userSkills []struct {
		UserID           string `json:"user_id"`
		SkillID          string `json:"skill_id"`
		ProficiencyLevel string `json:"proficiency_level"`
		AcquiredDate     string `json:"acquired_date"`
	}

	if err := json.Unmarshal(data, &userSkills); err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, us := range userSkills {
		_, err := tx.Exec(ctx, `
			INSERT INTO user_skills (user_id, skill_id, proficiency_level, acquired_date)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id, skill_id) DO NOTHING
		`, us.UserID, us.SkillID, us.ProficiencyLevel, us.AcquiredDate)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *Seeder) DeleteDevData(ctx context.Context) {
	_, err := s.pool.Exec(ctx, `DROP TABLE departments`)
	if err != nil {
		slog.Warn("Seeder | could not delete departments table")
	}

	_, err = s.pool.Exec(ctx, `DROP TABLE positions`)
	if err != nil {
		slog.Warn("Seeder | could not delete positions table")
	}

	_, err = s.pool.Exec(ctx, `DROP TABLE users`)
	if err != nil {
		slog.Warn("Seeder | could not delete users table")
	}

	_, err = s.pool.Exec(ctx, `DROP TABLE salaries`)
	if err != nil {
		slog.Warn("Seeder | could not delete salaries table")
	}

	_, err = s.pool.Exec(ctx, `DROP TABLE addresses`)
	if err != nil {
		slog.Warn("Seeder | could not delete addresses table")
	}

	_, err = s.pool.Exec(ctx, `DROP TABLE skills`)
	if err != nil {
		slog.Warn("Seeder | could not delete skills table")
	}

	_, err = s.pool.Exec(ctx, `DROP TABLE user_skills`)
	if err != nil {
		slog.Warn("Seeder | could not delete user_skills table")
	}
}
