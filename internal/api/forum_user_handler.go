package api

import (
	"context"

	"multi-processing-backend/internal/core"

	"github.com/gin-gonic/gin"
)

type ForumUserService interface {
	GetByEmail(ctx context.Context, email string) (*core.ForumUser, error)
	Create(ctx context.Context, user *core.ForumUser) error
	Update(ctx context.Context, user *core.ForumUser) error
	IsChannelMember(ctx context.Context, channelID, userID string) (bool, error)
	RegisterOrLogin(ctx context.Context, email, username string) (*core.ForumUser, error)
	GetUserChannels(ctx context.Context, userID string) ([]core.ForumChannel, error)
	GetChannelMessages(ctx context.Context, channelID, userID string) ([]core.ForumMessage, error)
	CreateMessage(ctx context.Context, message *core.ForumMessage) error
	MarkMessagesAsRead(ctx context.Context, channelID, userID string) error
}

type ForumUserHandler struct {
	service ForumUserService
}

func NewForumUserHandler(service ForumUserService) *ForumUserHandler {
	return &ForumUserHandler{service: service}
}

func RegisterForumUserRoutes(rg *gin.RouterGroup, h *ForumUserHandler) {
	users := rg.Group("")
	{
		users.POST("", h.Create)
	}
}

func (h *ForumUserHandler) Create(c *gin.Context) {

}

func (h *ForumUserHandler) GetByEmail(ctx context.Context, email string) {

}

func (h *ForumUserHandler) Update(ctx context.Context, user *core.ForumUser) {}

func (h *ForumUserHandler) IsChannelMember(ctx context.Context, channelID, userID string) {}

func (h *ForumUserHandler) RegisterOrLogin(ctx context.Context, email, username string) {}

func (h *ForumUserHandler) GetUserChannels(ctx context.Context, userID string) {}

func (h *ForumUserHandler) GetChannelMessages(ctx context.Context, channelID, userID string) {}

func (h *ForumUserHandler) CreateMessage(ctx context.Context, message *core.ForumMessage) {}

func (h *ForumUserHandler) MarkMessagesAsRead(ctx context.Context, channelID, userID string) {}
