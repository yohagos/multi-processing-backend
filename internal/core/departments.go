package core

import "time"

type Departments struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type DepartmentsPagination struct {
	Data []Departments `json:"data"`
	Total int64 `json:"total"`
	Error error `json:"error"`
}