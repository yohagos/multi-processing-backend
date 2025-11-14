package core

import "time"

type Salary struct {
	ID            string  `json:"id" db:"id"`
	UserID        string  `json:"user_id" db:"user_id"`
	Amount        float64 `json:"amount" db:"amount"`
	EffectiveDate time.Time `json:"effective_date" db:"effective_date"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type SalaryPaginationResponse struct {
	Data  []Salary `json:"data"`
	Total int64  `json:"total"`
	Error error  `json:"error"`
}