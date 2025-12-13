package core

import "time"

type ChannelMember struct {
	ChannelID string `json:"channel_id" db:"channel_id"`
	UserID    string `json:"user_id" db:"user_id"`
	Role      string `json:"role" db:"role"`
	JoinedAt  time.Time `json:"joined_at" db:"joined_at"`
}