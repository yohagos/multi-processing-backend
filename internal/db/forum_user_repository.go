package db

import (
	"context"
	"fmt"
	"multi-processing-backend/internal/core"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ForumUserRepository struct {
	pool *pgxpool.Pool
}

func NewForumUserRepository(pool *pgxpool.Pool) *ForumUserRepository {
	return &ForumUserRepository{pool: pool}
}

func (r * ForumUserRepository) GetByEmail(
	ctx context.Context, 
	email string,
) (*core.ForumUser, error) {
	var user core.ForumUser
	err := r.pool.QueryRow(ctx, `
		SELECT id, email, username, display_name, avatarurl, is_online, last_seen, created_at, updated_at
		FROM forum_users
		WHERE email = $1
	`, email).Scan(
		&user.ID, &user.Email,  &user.Username,  &user.DisplayName, 
		 &user.AvatarUrl,  &user.IsOnline,  &user.LastSeen,  
		 &user.CreatedAt,  &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r * ForumUserRepository) Create(
	ctx context.Context, 
	user *core.ForumUser,
) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO forum_users (email, username, display_name, avatar_url, is_online, last_seen, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, user.Email, user.Username, user.DisplayName, user.AvatarUrl, 
		user.IsOnline, user.LastSeen, user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID)
}

func (r * ForumUserRepository) Update(
	ctx context.Context, 
	user *core.ForumUser,
) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE forum_users
		SET display_name = $1, avatar_url = $2, is_online = $3, last_seen = $4, updated_at = $5
		WHERE id = $6
	`, user.DisplayName, user.AvatarUrl, user.IsOnline, user.LastSeen, time.Now(), user.ID)
	return err
}

func (r * ForumUserRepository) IsChannelMember(
	ctx context.Context, 
	channelID, userID string,
) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM channel_members
			WHERE channel_id = $1 AND user_id = $2
		)
	`, channelID, userID).Scan(&exists)
	return exists, err
}

func (r * ForumUserRepository) RegisterOrLogin(
	ctx context.Context, 
	email, username string,
) (*core.ForumUser, error) {
	user, err := r.GetByEmail(ctx, email)
	if err == nil {
		user.IsOnline = true
		user.LastSeen = time.Now()
		if err := r.Update(ctx, user); err != nil {
			return  nil, err
		}
		return user, nil
	}

	newUser := &core.ForumUser{
		Email: email,
		Username: username,
		DisplayName: username,
		IsOnline: true,
		LastSeen: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := r.Create(ctx, newUser); err != nil {
		return nil, err
	}
	
	return newUser, nil
}

func (r * ForumUserRepository) GetUserChannels(
	ctx context.Context, 
	userID string,
) ([]core.ForumChannel, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT fc.id, fc.name, fc.description, fc.is_private, fc.is_direct_message, fc.created_by, fc.created_at
		FROM forum_channels fc
		JOIN channel_members cm ON fc.id = cm.channel_id
		WHERE cm.user_id = $1
		ORDER BY fc.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return  pgx.CollectRows(rows, pgx.RowToStructByPos[core.ForumChannel])
}


func (r * ForumUserRepository) GetChannelMessages(
	ctx context.Context, 
	channelID, userID string,
) ([]core.ForumMessage, error) {
	isMember, err := r.IsChannelMember(ctx, channelID, userID)
	if err != nil || !isMember {
		return nil, fmt.Errorf("access denied or channel not found")
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, channel_id, user_id, content, message_type, parent_message_id, 
               is_edited, is_deleted, created_at, updated_at
        FROM forum_messages
        WHERE channel_id = $1 AND is_deleted = false
        ORDER BY created_at ASC
	`, channelID)
	if err != nil {
		return  nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByPos[core.ForumMessage])
}

/* func (r * ForumUserRepository) SendMessage(
	ctx context.Context, 
	channelID, userID, content string,
) (*core.ForumMessage, error) {
	isMember, err := r.IsChannelMember(ctx, channelID, userID)
	if err != nil || !isMember {
		return nil, fmt.Errorf("access denied")
	}

	message := &core.ForumMessage{
		ChannelID: channelID,
		UserID: userID,
		Content: content,
		MessageType: "text",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	

	return nil, nil
} */

func (r *ForumUserRepository) CreateMessage(
	ctx context.Context, 
	message *core.ForumMessage,
) error {
    return r.pool.QueryRow(ctx, `
        INSERT INTO forum_messages (id, channel_id, user_id, content, message_type, parent_message_id, 
                                    is_edited, is_deleted, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING id
    `, message.ID, message.ChannelID, message.UserID, message.Content, message.MessageType,
        message.ParentMessageID, message.IsEdited, message.IsDeleted, 
        message.CreatedAt, message.UpdatedAt,
    ).Scan(&message.ID)
}

func (r *ForumUserRepository) CreatMessage(
	ctx context.Context, 
	message *core.ForumMessage,
) error {
	return r.pool.QueryRow(ctx, `
			INSERT INTO forum_messages (id, channel_id, user_id, content, message_type, parent_message_id, 
										is_edited, is_deleted, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id
		`, message.ID, message.ChannelID, message.UserID, message.Content, message.MessageType,
			message.ParentMessageID, message.IsEdited, message.IsDeleted, 
			message.CreatedAt, message.UpdatedAt,
		).Scan(&message.ID)
}

func (r *ForumUserRepository) MarkMessagesAsRead(ctx context.Context, channelID, userID string) error {
    _, err := r.pool.Exec(ctx, `
        INSERT INTO message_read_status (message_id, user_id, read_at)
        SELECT fm.id, $2, NOW()
        FROM forum_messages fm
        WHERE fm.channel_id = $1 
          AND fm.user_id != $2
          AND NOT EXISTS (
              SELECT 1 FROM message_read_status mrs 
              WHERE mrs.message_id = fm.id AND mrs.user_id = $2
          )
    `, channelID, userID)
    return err
}
