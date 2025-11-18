package core

import (
	"time"
)

type Skill struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Category  string    `json:"category" db:"category"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type SkillUpdate struct {
	Name      *string    `json:"name,omitempty"`
	Category  *string    `json:"category,omitempty"`
}

type SkillPagination struct {
	Data []Skill
	Total int64
	Error error
}