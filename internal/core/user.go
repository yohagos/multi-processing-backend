package core

import "time"

type User struct {
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	FirstName    string    `json:"first_name" db:"first_name"`
	LastName     string    `json:"last_name" db:"last_name"`
	DepartmentID string    `json:"department_id" db:"department_id"`
	PositionID   string    `json:"position_id" db:"position_id"`
	HireDate     time.Time `json:"hire_date" db:"hire_date"`
	Phone        string    `json:"phone" db:"phone"`
	DateOfBirth  time.Time `json:"date_of_birth" db:"date_of_birth"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type UserUpdate struct {
	Email        *string   `json:"email,omitempty"`
	FirstName    *string   `json:"first_name,omitempty"`
	LastName     *string   `json:"last_name,omitempty"`
	DepartmentID string    `json:"department_id,omitempty"`
	PositionID   string    `json:"position_id,omitempty"`
	HireDate     time.Time `json:"hire_date,omitempty"`
	Phone        string    `json:"phone,omitempty"`
	DateOfBirth  time.Time `json:"date_of_birth,omitempty"`
}

type UserPaginationResponse struct {
	Data  []User `json:"data"`
	Total int64  `json:"total"`
	Error error  `json:"error"`
}

type UserSkill struct {
	ID               string    `json:"id" db:"id"`
	UserID           string    `json:"user_id" db:"user_id"`
	SkillID          string    `json:"skill_id" db:"skill_id"`
	ProficiencyLevel int       `json:"proficiency_level" db:"proficiency_level"`
	AcquiredDate     time.Time `json:"acquired_date" db:"acquired_date"`
}
