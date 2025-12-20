package db

import (
	"context"
	"database/sql"
	"fmt"
	"multi-processing-backend/internal/core"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/exp/slog"
)

type ForumUserRepository struct {
	pool *pgxpool.Pool
}

func NewForumUserRepository(pool *pgxpool.Pool) *ForumUserRepository {
	return &ForumUserRepository{pool: pool}
}

func (r *ForumUserRepository) GetByEmail(
	ctx context.Context,
	email string,
) (*core.ForumUser, error) {
	var user core.ForumUser
	var avatarUrl sql.NullString
	err := r.pool.QueryRow(ctx, `
		SELECT id, email, username, display_name, avatar_url, is_online, last_seen, created_at, updated_at
		FROM forum_users
		WHERE email = $1
	`, email).Scan(
		&user.ID, &user.Email, &user.Username, &user.DisplayName,
		&avatarUrl, &user.IsOnline, &user.LastSeen,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if avatarUrl.Valid {
		user.AvatarUrl = avatarUrl
	}

	return &user, nil
}

func (r *ForumUserRepository) GetByID(
	ctx context.Context,
	userID string,
) (*core.ForumUser, error) {
	var user core.ForumUser
	err := r.pool.QueryRow(ctx, `
		SELECT id, email, username, display_name, avatar_url, is_online, last_seen, created_at, updated_at
		FROM forum_users
		WHERE id = $1
	`, userID).Scan(
		&user.ID, &user.Email, &user.Username, &user.DisplayName,
		&user.AvatarUrl, &user.IsOnline, &user.LastSeen,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, err
}

func (r *ForumUserRepository) Create(
	ctx context.Context,
	user *core.ForumUser,
) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO forum_users (email, username, display_name, is_online, last_seen, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, user.Email, user.Username, user.DisplayName,
		user.IsOnline, user.LastSeen, user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID)
}

func (r *ForumUserRepository) Update(
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

func (r *ForumUserRepository) IsChannelMember(
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

func (r *ForumUserRepository) RegisterOrLogin(
	ctx context.Context,
	username, email string,
) (*core.ForumUser, error) {
	user, err := r.GetByEmail(ctx, email)
	if err == nil {
		user.IsOnline = true
		user.LastSeen = time.Now()
		if err := r.Update(ctx, user); err != nil {
			return nil, err
		}
		return user, nil
	}

	newUser := &core.ForumUser{
		Email:       email,
		Username:    username,
		DisplayName: username,
		AvatarUrl:   sql.NullString{},
		IsOnline:    true,
		LastSeen:    time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := r.Create(ctx, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

func (r *ForumUserRepository) GetUserChannels(
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

	return pgx.CollectRows(rows, pgx.RowToStructByPos[core.ForumChannel])
}

func (r *ForumUserRepository) GetChannelMessages(
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
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByPos[core.ForumMessage])
}

func (r *ForumUserRepository) CreateMessage(
	ctx context.Context,
	channelID, userID, content string,
) (*core.ForumMessage, error) {
	var message core.ForumMessage

	err := r.pool.QueryRow(ctx, `
        INSERT INTO forum_messages (id, channel_id, user_id, content, message_type, 
                                    is_edited, is_deleted, created_at, updated_at)
        VALUES (gen_random_uuid(), $1, $2, $3, 'text', false, false, NOW(), NOW())
        RETURNING id, channel_id, user_id, content, message_type, 
                  parent_message_id, is_edited, is_deleted, created_at, updated_at
    `, channelID, userID, content).Scan(
		&message.ID, &message.ChannelID, &message.UserID, &message.Content, &message.MessageType,
		&message.ParentMessageID, &message.IsEdited, &message.IsDeleted,
		&message.CreatedAt, &message.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &message, nil
}

/* func (r *ForumUserRepository) CreatMessage(
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
} */

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

func (r *ForumUserRepository) GetOrCreateDirectMessageChannel(
	ctx context.Context,
	user1Id, user2Id string,
) (string, error) {
	var existingChannelID string
	err := r.pool.QueryRow(ctx, `
		SELECT fc.id 
		FROM forum_channels fc
		JOIN channel_members cm1 ON fc.id = cm1.channel_id AND cm1.user_id = $1
		JOIN channel_members cm2 ON fc.id = cm2.channel_id AND cm2.user_id = $2
		WHERE fc.is_direct_message = true
		LIMIT 1
	`, user1Id, user2Id).Scan(&existingChannelID)

	if err == nil {
		return existingChannelID, nil
	}

	var channelID string
	err = r.pool.QueryRow(ctx, `
		SELECT create_direct_message_channel($1, $2)
	`, user1Id, user2Id).Scan(&channelID)
	return channelID, err
}

func (r *ForumUserRepository) GetOnlineUsers(
	ctx context.Context,
	userID string,
) ([]core.ForumUser, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, email, username, display_name, avatar_url, is_online, last_seen, created_at, updatedat
		FROM fourm_users
		WHERE is_online = true AND id != $1
		ORDER BY last_seen DESC
	`, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByPos[core.ForumUser])
}

func (r *ForumUserRepository) SearchUsers(
	ctx context.Context,
	query string,
	currentUserID string,
) ([]core.ForumUser, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, email, username, display_name, avatar_url, is_online, last_seen, created_at, updatedat
		FROM fourm_users
		WHERE (username ILIKE $1 OR display_name ILIKE $1) AND id != $2
		LIMIT 20
	`, "%"+query+"%", currentUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByPos[core.ForumUser])
}

func (r *ForumUserRepository) GetUnreadCount(
	ctx context.Context,
	userID string,
) (map[string]int, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT fm.channel_id, COUNT(*) as unread_count
		FROM forum_messages fm
		JOIN channel_members cm ON fm.channel_id = cm.channel_id
		LEFT JOIN message_read_status mrs ON fm.id = mrs.message_id AND mrs.user_id = $1
		WHERE cm.user_id = $1
			AND fm.uder_id != $1
			AND mrs.message_id IS NULL
			AND fm.is_deleted = false
		GROUP BY fm.channel_id
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var channelID string
		var count int
		if err := rows.Scan(&channelID, &count); err != nil {
			return nil, err
		}
		result[channelID] = count
	}
	return result, nil
}

func (r *ForumUserRepository) UpdateUserPresence(
	ctx context.Context,
	userID string,
	isOnline bool,
) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE forum_users
		SET is_online = $1, last_seen = NOW(), updated_at = NOW()
		WHERE id = $2
	`, isOnline, userID)
	return err
}

func (r *ForumUserRepository) GetChannelMembers(
	ctx context.Context,
	channelID string,
) ([]core.ForumUser, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT fu.id, fu.email, fu.username, fu.display_name, fu.avatar_url, 
			fu.is_online, fu.last_seen, fu.created_at, fu.updated_at
		FROM forum_users fu
		JOIN channel_members cm ON fu.id = cm.user_id
		WHERE cm.channel_id = $1
		ORDER BY cm.joined_at
	`, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByPos[core.ForumUser])
}

func (r *ForumUserRepository) EditMessage(
	ctx context.Context,
	messageID, userID, newContent string,
) error {
	_, err := r.pool.Exec(ctx, `
        UPDATE forum_messages 
        SET content = $1, is_edited = true, updated_at = NOW()
        WHERE id = $2 AND user_id = $3
    `, newContent, messageID, userID)
	return err
}

func (r *ForumUserRepository) DeleteMessage(
	ctx context.Context,
	messageID, userID string,
) error {
	_, err := r.pool.Exec(ctx, `
        UPDATE forum_messages 
        SET is_deleted = true, content = '[deleted]', updated_at = NOW()
        WHERE id = $1 AND user_id = $2
    `, messageID, userID)
	return err
}

func (r *ForumUserRepository) AddReaction(
	ctx context.Context,
	messageID, userID, emoji string,
) error {
	_, err := r.pool.Exec(ctx, `
        INSERT INTO message_reactions (message_id, user_id, emoji, created_at)
        VALUES ($1, $2, $3, NOW())
        ON CONFLICT (message_id, user_id, emoji) DO NOTHING
    `, messageID, userID, emoji)
	return err
}

func (r *ForumUserRepository) RemoveReaction(
	ctx context.Context,
	messageID, userID, emoji string,
) error {
	_, err := r.pool.Exec(ctx, `
        DELETE FROM message_reactions 
        WHERE message_id = $1 AND user_id = $2 AND emoji = $3
    `, messageID, userID, emoji)
	return err
}

// Clearing tables after shutdown
func (r *ForumUserRepository) DeleteForumTables(ctx context.Context) {
	_, err := r.pool.Exec(ctx, "DROP TABLE message_reactions CASCADE")
	if err != nil {
		slog.Warn("ForumUserRepository | DeleteForumTables | error occurred while deleting message_reactions")
	}

	_, err = r.pool.Exec(ctx, "DROP TABLE forum_messages CASCADE")
	if err != nil {
		slog.Warn("ForumUserRepository | DeleteForumTables | error occurred while deleting forum_messages")
	}

	_, err = r.pool.Exec(ctx, "DROP TABLE forum_users CASCADE")
	if err != nil {
		slog.Warn("ForumUserRepository | DeleteForumTables | error occurred while deleting forum_users")
	}

	_, err = r.pool.Exec(ctx, "DROP TABLE message_read_status CASCADE")
	if err != nil {
		slog.Warn("ForumUserRepository | DeleteForumTables | error occurred while deleting message_read_status")
	}

	_, err = r.pool.Exec(ctx, "DROP TABLE channel_members CASCADE")
	if err != nil {
		slog.Warn("ForumUserRepository | DeleteForumTables | error occurred while deleting channel_members")
	}

	/* _, err = r.pool.Exec(ctx, "DROP TABLE create_direct_message_channel CASCADE")
	if err != nil {
		slog.Warn("ForumUserRepository | DeleteForumTables | error occurred while deleting create_direct_message_channel")
	} */

	_, err = r.pool.Exec(ctx, "DROP TABLE forum_channels CASCADE")
	if err != nil {
		slog.Warn("ForumUserRepository | DeleteForumTables | error occurred while deleting forum_channels")
	}
}
