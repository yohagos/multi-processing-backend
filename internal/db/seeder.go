package db

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"multi-processing-backend/internal/core"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	insertedDepartments []core.Departments
	insertedPositions   []core.Position
	insertedUsers       []core.User
	insertedSkills      []core.Skill
	insertedAddresses   []core.Address
	insertedSalaries    []core.Salary

	assignedUserSkills = make(map[UserSkillkey]bool)
)

type Seeder struct {
	pool *pgxpool.Pool
}

type UserSkillkey struct {
	UserID  string
	SkillID string
}

func NewSeeder(pool *pgxpool.Pool) *Seeder {
	return &Seeder{pool: pool}
}

func (s *Seeder) SeedAll(ctx context.Context, basePath string) error {
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
		var count int
		err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM "+seed.tableName).Scan(&count)
		if err != nil {
			return fmt.Errorf("checking %s: %w", seed.tableName, err)
		}

		if count > 0 {

			continue
		}

		jsonPath := filepath.Join(basePath, seed.jsonFile)
		if err := seed.seedFunc(ctx, jsonPath); err != nil {
			return fmt.Errorf("seeding %s: %w", seed.tableName, err)
		}
	}
	return nil
}

func (s *Seeder) seedDepartments(ctx context.Context, jsonPath string) error {

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}

	var departments []core.Departments
	if err := json.Unmarshal(data, &departments); err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, d := range departments {
		var dep core.Departments
		err = tx.QueryRow(ctx, `
			INSERT INTO departments (name, description, created_at)
			VALUES ($1, $2, $3)
			RETURNING id, name, description, created_at
		`, d.Name, d.Description, d.CreatedAt).Scan(
			&dep.ID, &dep.Name, &dep.Description, &dep.CreatedAt,
		)
		if err != nil {
			return err
		}
		insertedDepartments = append(insertedDepartments, dep)

	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Seeder) seedPositions(ctx context.Context, jsonPath string) error {

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}
	var positions []core.Position
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
		//
		var pos core.Position
		err := tx.QueryRow(ctx, `
			INSERT INTO positions (title, level, department_id, created_at)
			VALUES ($1, $2, $3, $4)
			RETURNING id, title, level, department_id, created_at
		`, p.Title, p.Level, dep.ID, p.CreatedAt).Scan(
			&pos.ID, &pos.Title, &pos.Level, &pos.DepartmentID, &pos.CreatedAt,
		)
		if err != nil {
			return nil
		}
		insertedPositions = append(insertedPositions, pos)

	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func getRandomDepartment() core.Departments {

	return insertedDepartments[rand.Intn(len(insertedDepartments))]
}

func getRandomPosition() core.Position {

	return insertedPositions[rand.Intn(len(insertedPositions))]
}

func getRandomUser() core.User {
	return insertedUsers[rand.Intn(len(insertedUsers))]
}

func (s *Seeder) seedUsers(ctx context.Context, jsonPath string) error {

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}
	var users []core.User
	if err := json.Unmarshal(data, &users); err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, u := range users {
		posi := getRandomPosition()
		var user core.User

		err = tx.QueryRow(ctx, `
			INSERT INTO users (email, first_name, last_name, department_id, position_id, hire_date, phone, date_of_birth)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id, email, first_name, last_name, department_id, position_id, hire_date, phone, date_of_birth
		`, u.Email, u.FirstName, u.LastName, posi.DepartmentID,
			posi.ID, u.HireDate, u.Phone, u.DateOfBirth).Scan(
			&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.DepartmentID,
			&user.PositionID, &user.HireDate, &user.Phone, &user.DateOfBirth,
		)
		if err != nil {
			return err
		}
		insertedUsers = append(insertedUsers, user)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Seeder) seedSalaries(ctx context.Context, jsonPath string) error {

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}
	var salaries []core.Salary
	if err := json.Unmarshal(data, &salaries); err != nil {
		return err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, l := range salaries {
		user := getRandomUser()
		var sal core.Salary
		err = tx.QueryRow(ctx, `
			INSERT INTO salaries (user_id, amount, effective_date, created_at)
			VALUES ($1, $2, $3, $4)
			RETURNING id, user_id, amount, effective_date, created_at
		`, user.ID, l.Amount, l.EffectiveDate, l.CreatedAt).Scan(
			&sal.ID, &sal.UserID, &sal.Amount, &sal.EffectiveDate, &sal.CreatedAt,
		)
		if err != nil {
			return err
		}

		insertedSalaries = append(insertedSalaries, sal)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Seeder) seedAddresses(ctx context.Context, jsonPath string) error {

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}
	var addresses []core.Address
	if err := json.Unmarshal(data, &addresses); err != nil {
		return err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, a := range addresses {
		user := getRandomUser()
		var add core.Address
		err = tx.QueryRow(ctx, `
			INSERT INTO addresses (user_id, street, city, zip_code, country, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, user_id, street, city, zip_code, country, is_primary, created_at
		`, user.ID, a.Street, a.City, a.ZipCode, a.Country, a.CreatedAt).Scan(
			&add.ID, &add.UserID, &add.Street, &add.City, &add.ZipCode,
			&add.Country, &add.IsPrimary, &add.CreatedAt,
		)
		if err != nil {
			return err
		}
		insertedAddresses = append(insertedAddresses, add)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Seeder) seedSkills(ctx context.Context, jsonPath string) error {

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}

	var skills []core.Skill
	if err := json.Unmarshal(data, &skills); err != nil {
		return err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, k := range skills {
		var sk core.Skill
		err = tx.QueryRow(ctx, `
			INSERT INTO skills (name, category, created_at)
			VALUES ($1, $2, $3)
			RETURNING id, name, category, created_at
		`, k.Name, k.Category, k.CreatedAt).Scan(
			&sk.ID, &sk.Name, &sk.Category, &sk.CreatedAt,
		)
		if err != nil {
			return err
		}
		insertedSkills = append(insertedSkills, sk)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Seeder) seedUsersSkills(ctx context.Context, jsonPath string) error {

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}
	var user_skills []core.UserSkill
	if err := json.Unmarshal(data, &user_skills); err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	skills := s.getAllSkills(ctx)
	users := s.getAllUsers(ctx)

	for range users {
		var userID, skillID string
		var key UserSkillkey

		for attempts := 0; attempts < len(users); attempts++ {
			userID = users[rand.Intn(len(users))].ID
			skillID = skills[rand.Intn(len(skills))].ID
			key = UserSkillkey{UserID: userID, SkillID: skillID}

			if !assignedUserSkills[key] {
				break
			}
		}

		if assignedUserSkills[key] {
			continue
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO user_skills (user_id, skill_id, proficiency_level, acquired_date)
			VALUES ($1, $2, $3, $4)
		`, userID, skillID, rand.Intn(5)+1, time.Now().AddDate(0, -rand.Intn(24), 0))
		if err != nil {
			return err
		}

		assignedUserSkills[key] = true
	}
	return tx.Commit(ctx)
}

func (s *Seeder) getAllUsers(ctx context.Context) []core.User {
	rows, err := s.pool.Query(ctx, "SELECT id FROM users")
	if err != nil {
		return nil
	}
	defer rows.Close()

	var users []core.User
	for rows.Next() {
		var user core.User
		if err := rows.Scan(&user.ID); err != nil {
			continue
		}
		users = append(users, user)
	}
	return users
}

func (s *Seeder) getAllSkills(ctx context.Context) []core.Skill {
	rows, err := s.pool.Query(ctx, "SELECT id FROM skills")
	if err != nil {
		return nil
	}
	defer rows.Close()

	var skills []core.Skill
	for rows.Next() {
		var skill core.Skill
		if err := rows.Scan(&skill.ID); err != nil {
			continue
		}
		skills = append(skills, skill)
	}
	return skills
}

func (s *Seeder) DeleteDevData(ctx context.Context) {
	_, err := s.pool.Exec(ctx, `DROP TABLE departments`)
	if err != nil {

	}

	_, err = s.pool.Exec(ctx, `DROP TABLE positions`)
	if err != nil {

	}

	_, err = s.pool.Exec(ctx, `DROP TABLE users`)
	if err != nil {

	}

	_, err = s.pool.Exec(ctx, `DROP TABLE salaries`)
	if err != nil {

	}

	_, err = s.pool.Exec(ctx, `DROP TABLE addresses`)
	if err != nil {

	}

	_, err = s.pool.Exec(ctx, `DROP TABLE skills`)
	if err != nil {

	}

	_, err = s.pool.Exec(ctx, `DROP TABLE user_skills`)
	if err != nil {

	}
}
