package core

import "time"

type Position struct {
	ID           string    `json:"id" db:"id"`
	Title        string    `json:"title" db:"title"`
	Level        int       `json:"level" db:"level"`
	DepartmentID string    `json:"department_id" db:"department_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type PositionUpdate struct {
	Title        *string    `json:"title"`
	Level        *int       `json:"level"`
	DepartmentID string    `json:"department_id"`
}

type PositionPagination struct {
	Data  []Position `json:"data"`
	Total int64      `json:"total"`
	Error error      `json:"error"`
}
