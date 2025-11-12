package core

import "time"

type Address struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Street    string    `json:"street" db:"street"`
	City      string    `json:"city" db:"city"`
	ZipCode   string    `json:"zip_code" db:"zip_code"`
	Country   string    `json:"country" db:"country"`
	IsPrimary string    `json:"is_primary" db:"is_primary"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}