package core

import (
	"time"
)

type Skill struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Category  string    `json:"category" db:"category"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type SkillWithDetails struct {
	ID               string    `json:"id" db:"id"`
	Name             string    `json:"name" db:"name"`
	Category         string    `json:"category" db:"category"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
	ProficiencyLevel int       `json:"proficiency_level" db:"proficiency_level"`
	AcquiredDate     time.Time `json:"acquired_date" db:"acquired_date"`
}

type SkillUpdate struct {
	Name     *string `json:"name,omitempty"`
	Category *string `json:"category,omitempty"`
}

type SkillPagination struct {
	Data  []Skill `json:"data"`
	Total int64   `json:"total"`
	Error error   `json:"error"`
}
