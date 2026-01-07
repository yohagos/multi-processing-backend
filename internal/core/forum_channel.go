package core

import "time"

type ForumChannel struct {
	ID              string    `json:"id" db:"id"`
	Name            string    `json:"name" db:"name"`
	Description     string    `json:"description,omitempty" db:"description"`
	IsPrivate       bool      `json:"is_private" db:"is_private"`
	IsDirectMessage bool      `json:"is_direct_message" db:"is_direct_message"`
	CreatedBy       string    `json:"created_by" db:"created_by"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type ForumChannelMessages struct {
	Channel  ForumChannel   `json:"channel"`
	Messages []ForumMessage `json:"messages"`
	Page     int            `json:"page"`
	Limit    int            `json:"limit"`
	Total    int64          `json:"total"`
}
