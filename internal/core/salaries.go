package core

import "time"

type Salary struct {
	ID            string  `json:"id" db:"id"`
	UserID        string  `json:"user_id" db:"user_id"`
	Amount        float64 `json:"amount" db:"amount"`
	EffectiveDate time.Time `json:"effective_date" db:"effective_date"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type SalaryUpdate struct {
	UserID        string  `json:"user_id"`
	Amount        float64 `json:"amount,omitempty"`
	EffectiveDate time.Time `json:"effective_date,omitempty"`
}

type SalaryPagination struct {
	Data  []Salary `json:"data"`
	Total int64  `json:"total"`
	Error error  `json:"error"`
}