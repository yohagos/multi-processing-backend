package services

import (
	"context"

	"multi-processing-backend/internal/core"
)

type ForumUserRepository interface {
	GetByID(ctx context.Context, id string) (*core.ForumUser, error)
	GetByEmail(ctx context.Context, id string) (*core.ForumUser, error)
	Create(ctx context.Context, user *core.ForumUser) error
	Update(ctx context.Context, user *core.ForumUser) error
	IsChannelMember(ctx context.Context, channelID, userID string) (bool, error)
	RegisterOrLogin(ctx context.Context, email, username string) (*core.ForumUser, error)
	GetUserChannels(ctx context.Context, userID string) ([]core.ForumChannel, error)
	GetChannelMessages(ctx context.Context, channelID, userID string) ([]core.ForumMessage, error)
	CreateMessage(ctx context.Context, channelID, userID, content string) (*core.ForumMessage, error)
	MarkMessagesAsRead(ctx context.Context, channelID, userID string) error
	GetPublicChannelMessages(ctx context.Context, page, limit int) (*core.ForumChannelMessages, error)

	GetOrCreateDirectMessageChannel(ctx context.Context, user1ID, user2ID string) (string, error)
	GetOnlineUsers(ctx context.Context, userID string) ([]core.ForumUser, error)
	SearchUsers(ctx context.Context, query string, currentUserID string) ([]core.ForumUser, error)
	GetUnreadCount(ctx context.Context, userID string) (map[string]int, error)
	UpdateUserPresence(ctx context.Context, userID string, isOnline bool) error
	GetChannelMembers(ctx context.Context, channelID string) ([]core.ForumUser, error)
	EditMessage(ctx context.Context, messageID, userID, newContent string) error
	DeleteMessage(ctx context.Context, messageID, userID string) error
	AddReaction(ctx context.Context, messageID, userID, emoji string) error
	RemoveReaction(ctx context.Context, messageID, userID, emoji string) error
}

type ForumUserService struct {
	repo ForumUserRepository
}

func NewForumUserService(repo ForumUserRepository) *ForumUserService {
	return &ForumUserService{repo: repo}
}

func (s *ForumUserService) GetByID(ctx context.Context, id string) (*core.ForumUser, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ForumUserService) GetByEmail(ctx context.Context, email string) (*core.ForumUser, error) {
	return s.repo.GetByID(ctx, email)
}

func (s *ForumUserService) Create(ctx context.Context, user *core.ForumUser) error {
	return s.repo.Create(ctx, user)
}

func (s *ForumUserService) Update(ctx context.Context, user *core.ForumUser) error {
	return s.repo.Update(ctx, user)
}

func (s *ForumUserService) IsChannelMember(ctx context.Context, channelID, userID string) (bool, error) {
	return s.repo.IsChannelMember(ctx, channelID, userID)
}

func (s *ForumUserService) RegisterOrLogin(ctx context.Context, username, email string) (*core.ForumUser, error) {
	return s.repo.RegisterOrLogin(ctx, username, email)
}

func (s *ForumUserService) GetUserChannels(ctx context.Context, userID string) ([]core.ForumChannel, error) {
	return s.repo.GetUserChannels(ctx, userID)
}

func (s *ForumUserService) GetChannelMessages(ctx context.Context, channelID, userID string) ([]core.ForumMessage, error) {
	return s.repo.GetChannelMessages(ctx, channelID, userID)
}

func (s *ForumUserService) CreateMessage(ctx context.Context, channelID, userID, content string) (*core.ForumMessage, error) {
	return s.repo.CreateMessage(ctx, channelID, userID, content)
}

func (s *ForumUserService) MarkMessagesAsRead(ctx context.Context, channelID, userID string) error {
	return s.repo.MarkMessagesAsRead(ctx, channelID, userID)
}

func (s *ForumUserService) GetPublicChannelMessages(ctx context.Context, page, limit int) (*core.ForumChannelMessages, error) {
	return s.repo.GetPublicChannelMessages(ctx, page, limit)
}

func (s *ForumUserService) GetOrCreateDirectMessageChannel(ctx context.Context, user1ID, user2ID string) (string, error) {
	return s.repo.GetOrCreateDirectMessageChannel(ctx, user1ID, user2ID)
}

func (s *ForumUserService) GetOnlineUsers(ctx context.Context, userID string) ([]core.ForumUser, error) {
	return s.repo.GetOnlineUsers(ctx, userID)
}

func (s *ForumUserService) SearchUsers(ctx context.Context, query string, currentUserID string) ([]core.ForumUser, error) {
	return s.repo.SearchUsers(ctx, query, currentUserID)
}

func (s *ForumUserService) GetUnreadCount(ctx context.Context, userID string) (map[string]int, error) {
	return s.repo.GetUnreadCount(ctx, userID)
}

func (s *ForumUserService) UpdateUserPresence(ctx context.Context, userID string, isOnline bool) error {
	return s.repo.UpdateUserPresence(ctx, userID, isOnline)
}

func (s *ForumUserService) GetChannelMembers(ctx context.Context, channelID string) ([]core.ForumUser, error) {
	return s.repo.GetChannelMembers(ctx, channelID)
}

func (s *ForumUserService) EditMessage(ctx context.Context, messageID, userID, newContent string) error {
	return s.repo.EditMessage(ctx, messageID, userID, newContent)
}

func (s *ForumUserService) DeleteMessage(ctx context.Context, messageID, userID string) error {
	return s.repo.DeleteMessage(ctx, messageID, userID)
}

func (s *ForumUserService) AddReaction(ctx context.Context, messageID, userID, emoji string) error {
	return s.repo.AddReaction(ctx, messageID, userID, emoji)
}

func (s *ForumUserService) RemoveReaction(ctx context.Context, messageID, userID, emoji string) error {
	return s.repo.RemoveReaction(ctx, messageID, userID, emoji)
}
