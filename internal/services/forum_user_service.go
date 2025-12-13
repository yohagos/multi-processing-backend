package services

import (
	"context"

	"multi-processing-backend/internal/core"
)

type ForumUserRepository interface {
	GetByEmail(ctx context.Context, email string) (*core.ForumUser, error)
    Create(ctx context.Context, user *core.ForumUser) error
    Update(ctx context.Context, user *core.ForumUser) error
	IsChannelMember(ctx context.Context, channelID, userID string) (bool, error)
	RegisterOrLogin(ctx context.Context, email, username string) (*core.ForumUser, error)
	GetUserChannels(ctx context.Context, userID string) ([]core.ForumChannel, error)
	GetChannelMessages(ctx context.Context, channelID, userID string) ([]core.ForumMessage, error)
	SendMessage(ctx context.Context, channelID, userID, content string) (*core.ForumMessage, error)
}

type ForumUserService struct {
	repo ForumUserRepository
}

func NewForumUserService(repo ForumUserRepository) *ForumUserService {
	return &ForumUserService{repo: repo}
}

func (s *ForumUserService) GetByEmail(ctx context.Context, email string) (*core.ForumUser, error) {
	return s.repo.GetByEmail(ctx, email)
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

func (s *ForumUserService) RegisterOrLogin(ctx context.Context, email, username string) (*core.ForumUser, error) {
	return s.repo.RegisterOrLogin(ctx, email, username)
}

func (s *ForumUserService) GetUserChannels(ctx context.Context, userID string) ([]core.ForumChannel, error) {
	return  s.repo.GetUserChannels(ctx, userID)
}

func (s *ForumUserService) GetChannelMessages(ctx context.Context, channelID, userID string) ([]core.ForumMessage, error) {
	return s.repo.GetChannelMessages(ctx, channelID, userID)
}

func (s *ForumUserService) SendMessage(ctx context.Context, channelID, userID, content string) (*core.ForumMessage, error) {
	return s.repo.SendMessage(ctx, channelID, userID, content)
}
