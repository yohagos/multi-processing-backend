package core

import (
	"encoding/json"
	"time"
)

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

type UserPagination struct {
	Data  []User `json:"data"`
	Total int64  `json:"total"`
	Error error  `json:"error"`
}

type UserWithDetails struct {
	User
	Departments *Departments `json:"department,omitempty"`
	Position    *Position    `json:"position,omitempty"`
	Address     *Address     `json:"address,omitempty"`
	Skill      []UserSkill       `json:"skill,omitempty"`
}

type UserWithDetailsPagination struct {
	Data  []UserWithDetails `json:"data"`
	Total int64             `json:"total"`
	Error error             `json:"error"`
}

type UserSkill struct {
	ID               string    `json:"id" db:"id"`
	UserID           string    `json:"user_id" db:"user_id"`
	SkillID          string    `json:"skill_id" db:"skill_id"`
	ProficiencyLevel int       `json:"proficiency_level" db:"proficiency_level"`
	AcquiredDate     time.Time `json:"acquired_date" db:"acquired_date"`
}

func (us *UserSkill) UnmarshelJSON(data []byte) error {
	type Alias UserSkill

	aux := &struct {
		AcquiredDate string `json:"acquireddate"`
		*Alias
	}{
		Alias: (*Alias)(us),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	if aux.AcquiredDate != "" {
		parsedDate, err := time.Parse("2006-01-01", aux.AcquiredDate)
		if err != nil {
			return  err
		}
		us.AcquiredDate = parsedDate
	}
	return nil
}
