package core

import (
	"database/sql"
	"time"
)

type ForumUser struct {
	ID          string    `json:"id" db:"id"`
	Email       string    `json:"email" db:"email"`
	Username    string    `json:"username" db:"username"`
	DisplayName string    `json:"display_name,omitempty" db:"display_name"`
	AvatarUrl   sql.NullString    `json:"avatar_url,omitempty" db:"avatar_url"`
	IsOnline    bool      `json:"is_online" db:"is_online"`
	LastSeen    time.Time `json:"last_seen,omitempty" db:"last_seen"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
