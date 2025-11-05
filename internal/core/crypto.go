package core

import "time"

type Crypto struct {
	ID            string    `json:"id" db:"id"`
	Initial       string    `json:"initial" db:"initial"`
	Name          string    `json:"name" db:"name"`
	CurrentValue  float64   `json:"current_value" db:"current_value"`
	PreviousValue float64   `json:"previous_value" db:"previous_value"`
	Percent       float64   `json:"percent" db:"percent"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	IsInitial     bool      `json:"-" db:"is_initial"`
}
