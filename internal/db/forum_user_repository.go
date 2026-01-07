package db

import (
	"context"
	"database/sql"
	"fmt"
	"multi-processing-backend/internal/core"
	"os"
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

	var channels []core.ForumChannel
	for rows.Next() {
		var ch core.ForumChannel
		var description sql.NullString

		err := rows.Scan(
			&ch.ID, &ch.Name, &description, &ch.IsPrivate, &ch.IsDirectMessage,
			&ch.CreatedBy, &ch.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if description.Valid {
			ch.Description = description.String
		}

		channels = append(channels, ch)
	}
	return channels, nil
}

func (r *ForumUserRepository) GetChannelMessages(
	ctx context.Context,
	channelID, userID string,
) ([]core.ForumMessage, error) {
	const operation = "ForumRepository.GetChannelMessages"
	isMember, err := r.IsChannelMember(ctx, channelID, userID)
	 if err != nil {
        slog.Error(operation+" IsChannelMember failed", 
            "error", err, "userID", userID, "channelID", channelID)
        return nil, fmt.Errorf("%s: access check failed: %w", operation, err)
    }
    if !isMember {
        slog.Warn(operation+" access denied", 
            "userID", userID, "channelID", channelID)
        return nil, fmt.Errorf("%s: access denied", operation)
    }

	rows, err := r.pool.Query(ctx, `
		SELECT id, channel_id, user_id, content, message_type, parent_message_id, 
               is_edited, is_deleted, created_at, updated_at
        FROM forum_messages
        WHERE channel_id = $1 AND is_deleted = false
        ORDER BY created_at ASC
	`, channelID)
	if err != nil {
        slog.Error(operation+" query failed", 
            "error", err, "channelID", channelID)
        return nil, fmt.Errorf("%s: database query failed: %w", operation, err)
    }
	defer rows.Close()

	var msgs []core.ForumMessage
	for rows.Next() {
		var m core.ForumMessage
		var parentMessageID sql.NullString

		err := rows.Scan(
			&m.ID, &m.ChannelID, &m.UserID, &m.Content, &m.MessageType,
			&parentMessageID, &m.IsEdited, &m.IsDeleted, &m.CreatedAt, &m.UpdatedAt,
		)
		 if err != nil {
            slog.Error(operation+" row scan failed", "error", err)
            return nil, fmt.Errorf("%s: failed to scan row: %w", operation, err)
        }

		if parentMessageID.Valid {
			m.ParentMessageID = parentMessageID.String
		}

		msgs = append(msgs, m)
	}

	if err := rows.Err(); err != nil {
        slog.Error(operation+" rows iteration failed", "error", err)
        return nil, fmt.Errorf("%s: rows iteration failed: %w", operation, err)
    }

	if msgs == nil {
		msgs = []core.ForumMessage{}
	}

	return msgs, nil
}

func (r *ForumUserRepository) GetPublicChannelMessages(
	ctx context.Context,
	page, limit int,
) (*core.ForumChannelMessages, error) {
	offset := (page - 1) * limit
	channel, err := r.GetPublicChannel(ctx)
	if err != nil {
		return nil, err
	}

	var total int64
	err = r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM forum_messages
		WHERE channel_id = $1 AND is_deleted = false
	`, channel.ID).Scan(&total)
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT 
			fm.id, fm.channel_id, fm.user_id, fm.content, fm.message_type, 
			fm.parent_message_id, fm.is_edited, fm.is_deleted, fm.created_at, fm.updated_at,

			fu.id, fu.email, fu.username, fu.display_name, fu.avatar_url, fu.is_online,
			fu.last_seen, fu.created_at, fu.updated_at,

			pf.id, pf.channel_id, pf.user_id, pf.content, pf.message_type, 
			pf.parent_message_id, pf.is_edited, pf.is_deleted, pf.created_at, pf.updated_at,

			pfu.id, pfu.email, pfu.username, pfu.display_name, pfu.avatar_url, pfu.is_online,
			pfu.last_seen, pfu.created_at, pfu.updated_at
		FROM forum_messages fm
		JOIN forum_users fu ON fm.user_id = fu.id
		LEFT JOIN forum_messages pf ON fm.parent_message_id = pf.id AND pf.is_deleted = false
		LEFT JOIN forum_users pfu ON pf.user_id = pfu.id
		WHERE fm.channel_id = $1 AND fm.is_deleted = false
		ORDER BY fm.created_at ASC
		LIMIT $2 OFFSET $3
	`, channel.ID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []core.ForumMessage
	for rows.Next() {
		var msg core.ForumMessage
		var parentMsgID sql.NullString
		var avatarUrl sql.NullString

		msg.User = &core.ForumUser{}

		var userID, userEmail, userUsername, userDisplayName sql.NullString
		var userIsOnline sql.NullBool
		var userLastSeen, userCreatedAt, userUpdatedAt sql.NullTime

		var parentMessageID sql.NullString
		var parentMessageChannelID sql.NullString
		var parentMessageUserID sql.NullString
		var parentMessageContent sql.NullString
		var parentMessageType sql.NullString
		var parentMessageIsEdited sql.NullBool
		var parentMessageIsDeleted sql.NullBool
		var parentMessageCreatedAt, parentMessageUpdatedAt sql.NullTime

		var parentUserID, parentUserEmail, parentUserUsername, parentUserDisplayName sql.NullString
		var parentUserAvatar sql.NullString
		var parentUserIsOnline sql.NullBool
		var parentUserLastSeen, parentUserCreatedAt, parentUserUpdatedAt sql.NullTime

		err := rows.Scan(
			&msg.ID, &msg.ChannelID, &msg.UserID, &msg.Content, &msg.MessageType,
			&parentMsgID, &msg.IsEdited, &msg.IsDeleted, &msg.CreatedAt, &msg.UpdatedAt,

			&userID, &userEmail, &userUsername, &userDisplayName,
			&avatarUrl, &userIsOnline, &userLastSeen, &userCreatedAt, &userUpdatedAt,

			&parentMessageID, &parentMessageChannelID, &parentMessageUserID,
			&parentMessageContent, &parentMessageType, &parentMsgID,
			&parentMessageIsEdited, &parentMessageIsDeleted,
			&parentMessageCreatedAt, &parentMessageUpdatedAt,

			&parentUserID, &parentUserEmail, &parentUserUsername, &parentUserDisplayName,
			&parentUserAvatar, &parentUserIsOnline,
			&parentUserLastSeen, &parentUserCreatedAt, &parentUserUpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if parentMessageID.Valid {
			msg.ParentMessageID = parentMsgID.String

			if parentMessageID.Valid {
				msg.ParentMessage = &core.ForumMessage{
					ID:          parentMessageID.String,
					ChannelID:   parentMessageChannelID.String,
					UserID:      parentMessageUserID.String,
					Content:     parentMessageContent.String,
					MessageType: parentMessageType.String,
					IsEdited:    parentMessageIsEdited.Bool,
					IsDeleted:   parentMessageIsDeleted.Bool,
					CreatedAt:   parentMessageCreatedAt.Time,
					UpdatedAt:   parentMessageUpdatedAt.Time,
					User: &core.ForumUser{
						ID:          parentUserID.String,
						Email:       parentUserEmail.String,
						Username:    parentUserUsername.String,
						DisplayName: parentUserDisplayName.String,
						AvatarUrl:   parentUserAvatar,
						IsOnline:    parentUserIsOnline.Bool,
						LastSeen:    parentUserLastSeen.Time,
						CreatedAt:   parentUserCreatedAt.Time,
						UpdatedAt:   parentUserUpdatedAt.Time,
					},
				}
			}
		}

		if userID.Valid {
			msg.User = &core.ForumUser{
				ID:          userID.String,
				Email:       userEmail.String,
				Username:    userUsername.String,
				DisplayName: userDisplayName.String,
				AvatarUrl:   avatarUrl,
				IsOnline:    userIsOnline.Bool,
				LastSeen:    userLastSeen.Time,
				CreatedAt:   userCreatedAt.Time,
				UpdatedAt:   userUpdatedAt.Time,
			}
		}

		if avatarUrl.Valid {
			msg.User.AvatarUrl = avatarUrl
		}

		messages = append(messages, msg)
	}

	if messages == nil {
		messages = []core.ForumMessage{}
	}

	response := &core.ForumChannelMessages{
		Channel:  channel,
		Messages: messages,
		Page:     page,
		Limit:    limit,
		Total:    total,
	}

	return response, nil
}

func (r *ForumUserRepository) GetPublicChannel(ctx context.Context) (core.ForumChannel, error) {
	var ch core.ForumChannel
	publicChannel := "Public Channel"
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, description, is_private, is_direct_message, created_by, created_at
		FROM forum_channels
		WHERE name = $1
		LIMIT 1
	`, publicChannel).Scan(
		&ch.ID, &ch.Name, &ch.Description, &ch.IsPrivate, &ch.IsDirectMessage,
		&ch.CreatedBy, &ch.CreatedAt,
	)
	if err != nil {
		slog.Error("ForumUserRepository | findPublicChannelID | cannot find id of public channel", "error", err.Error())
		return core.ForumChannel{}, err
	}
	return ch, nil
}

func (r *ForumUserRepository) CreateMessage(
	ctx context.Context,
	channelID, userID, content, parentMessageID string,
) (*core.ForumMessage, error) {
	var message core.ForumMessage

	var parMsgValue interface{}

	if parentMessageID != "" {
		var exists bool
		err := r.pool.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM forum_messages WHERE id = $1 AND is_deleted = false)
		`, parentMessageID).Scan(&exists)
		if err != nil {
			return nil, err
		}

		if exists {
			parMsgValue = parentMessageID
		} else {
			parMsgValue = nil
		}
	} else {
		parMsgValue = nil
	}

	var scannedPMsgID sql.NullString

	err := r.pool.QueryRow(ctx, `
        INSERT INTO forum_messages (channel_id, user_id, content, message_type, parent_message_id,
                                    is_edited, is_deleted, created_at, updated_at)
        VALUES ($1, $2, $3, 'text', $4, false, false, NOW(), NOW())
        RETURNING id, channel_id, user_id, content, message_type, 
                  parent_message_id, is_edited, is_deleted, created_at, updated_at
    `, channelID, userID, content, parMsgValue).Scan(
		&message.ID, &message.ChannelID, &message.UserID, &message.Content, &message.MessageType,
		&scannedPMsgID, &message.IsEdited, &message.IsDeleted,
		&message.CreatedAt, &message.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if scannedPMsgID.Valid {
		message.ParentMessageID = scannedPMsgID.String
	} else {
		message.ParentMessageID = ""
	}
	return &message, nil
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

// Creating public channel at App start
func (r *ForumUserRepository) CreatePublicChannel(ctx context.Context) {
	var pc core.ForumChannel
	var channel = "Public Channel"

	err := r.pool.QueryRow(ctx, `
		SELECT id, name, description, is_private, is_direct_message, created_by, created_at
		FROM forum_channels
		WHERE name = $1
		LIMIT 1
	`, channel).Scan(
		&pc.ID, &pc.Name, &pc.Description, &pc.IsPrivate,
		&pc.IsDirectMessage, &pc.CreatedBy, &pc.CreatedAt,
	)

	if err != nil {
		if pc.ID == "" {
			id := r.createAdminUser(ctx)
			if id == "" {
				slog.Warn("ForumUserRepository | CreatePublicChannel | Error while Admin user", "error", err.Error())
				os.Exit(1000)
			}
			pc := &core.ForumChannel{
				Name:            "Public Channel",
				Description:     "Public Channel",
				IsPrivate:       false,
				IsDirectMessage: false,
				CreatedBy:       id,
				CreatedAt:       time.Now(),
			}

			_, err := r.pool.Exec(ctx, `
				INSERT INTO forum_channels (name, description, is_private, is_direct_message, created_by, created_at)
				VALUES ($1, $2, $3, $4, $5, $6)
				ON CONFLICT (id) DO NOTHING
			`,
				&pc.Name, &pc.Description, &pc.IsPrivate, &pc.IsDirectMessage,
				&pc.CreatedBy, &pc.CreatedAt,
			)

			if err != nil {
				slog.Warn("ForumUserRepository | CreatePublicChannel | Tried to create Public Channel but another error occurred.", "error", err.Error())
			}
		}
	}
}

func (r *ForumUserRepository) createAdminUser(ctx context.Context) string {
	admin := &core.ForumUser{
		Email:       "admin@admin.admin",
		Username:    "admin",
		DisplayName: "Admin",
		AvatarUrl:   sql.NullString{},
		IsOnline:    false,
		LastSeen:    time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	var id string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO forum_users (email, username, display_name, avatar_url, is_online, last_seen, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`,
		admin.Email, admin.Username, admin.DisplayName, admin.AvatarUrl,
		admin.IsOnline, admin.LastSeen, admin.CreatedAt, admin.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return ""
	}

	return id
}

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

	_, err = r.pool.Exec(ctx, "DROP TABLE forum_channels CASCADE")
	if err != nil {
		slog.Warn("ForumUserRepository | DeleteForumTables | error occurred while deleting forum_channels")
	}
}
