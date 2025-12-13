package core

import "time"

type ForumMessage struct {
	ID              string    `json:"id" db:"id"`
	ChannelID       string    `json:"channel_id" db:"channel_id"`
	UserID          string    `json:"user_id" db:"user_id"`
	Content         string    `json:"content" db:"content"`
	MessageType     string    `json:"message_type" db:"message_type"`
	ParentMessageID string    `json:"parent_message_id,omitempty" db:"parent_message_id"`
	IsEdited        bool      `json:"is_edited" db:"is_edited"`
	IsDeleted       bool      `json:"is_deleted" db:"is_deleted"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}
