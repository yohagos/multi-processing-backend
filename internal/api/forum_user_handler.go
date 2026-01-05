package api

import (
	"context"
	"net/http"
	"strconv"

	"multi-processing-backend/internal/core"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
)

type ForumUserService interface {
	GetByID(ctx context.Context, id string) (*core.ForumUser, error)
	GetByEmail(ctx context.Context, email string) (*core.ForumUser, error)
	Create(ctx context.Context, user *core.ForumUser) error
	Update(ctx context.Context, user *core.ForumUser) error
	IsChannelMember(ctx context.Context, channelID, userID string) (bool, error)
	RegisterOrLogin(ctx context.Context, email, username string) (*core.ForumUser, error)
	GetUserChannels(ctx context.Context, userID string) ([]core.ForumChannel, error)
	GetChannelMessages(ctx context.Context, channelID, userID string) ([]core.ForumMessage, error)
	CreateMessage(ctx context.Context, channelID, userID, content, parentMessageId string) (*core.ForumMessage, error)
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

type ForumUserHandler struct {
	service ForumUserService
}

func NewForumUserHandler(service ForumUserService) *ForumUserHandler {
	return &ForumUserHandler{service: service}
}

func RegisterForumUserRoutes(rg *gin.RouterGroup, h *ForumUserHandler) {
	users := rg.Group("/users")
	{
		users.GET("/search", h.SearchUsers)
		users.GET("/online", h.GetOnlineUsers)
		users.GET("/:id", h.GetByID)
		users.GET("/email", h.GetByEmail)

		users.POST("/login", h.RegisterOrLogin)
		users.POST("", h.Create)

		users.PATCH("/:id", h.Update)
		users.PATCH("/:id/presence", h.UpdateUserPresence)
	}

	channels := rg.Group("channels")
	{
		channels.GET("/public", h.GetPublicChannelMessages)
		channels.GET("/:id/members", h.GetChannelMembers)
		channels.GET("/:id/messages", h.GetChannelMessages)
		channels.GET("/:id/unread", h.GetUnreadCount)
		channels.GET("/user/:userID", h.GetUserChannels)
		channels.GET("/direct", h.GetOrCreateDirectMessageChannel)

		channels.POST("/:id/messages", h.CreateMessage)
		channels.PATCH("/:id/read", h.MarkMessagesAsRead)
	}

	messages := rg.Group("/messages")
	{
		messages.PATCH("/:id", h.EditMessage)
		messages.DELETE("/:id", h.DeleteMessage)
		messages.POST("/:id/reactions", h.AddReaction)
		messages.DELETE("/:id/reactions", h.RemoveReaction)
	}
}

func (h *ForumUserHandler) Create(c *gin.Context) {
	var req core.ForumUser

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}

func (h *ForumUserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	user, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *ForumUserHandler) GetByEmail(c *gin.Context) {
	email := c.Query("email")

	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email query parameter required"})
		return
	}

	user, err := h.service.GetByEmail(c.Request.Context(), email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *ForumUserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req core.ForumUser

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ID = id

	err := h.service.Update(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, req)
}

func (h *ForumUserHandler) RegisterOrLogin(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, err := h.service.RegisterOrLogin(c.Request.Context(), req.Username, req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *ForumUserHandler) GetPublicChannelMessages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "40"))

	messages, err := h.service.GetPublicChannelMessages(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, messages)
}

func (h *ForumUserHandler) GetUserChannels(c *gin.Context) {
	userID := c.Param("userID")

	slog.Info("\nForumHandler | \nGetUserChannels() | UserID received", "id", userID)

	channels, err := h.service.GetUserChannels(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slog.Info("\nForumHandler | \nGetUserChannels() | Channels found for User", "data", channels)

	c.JSON(http.StatusOK, channels)
}

func (h *ForumUserHandler) GetChannelMessages(c *gin.Context) {
	channelID := c.Param("id")
	userID := c.Query("userID")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userID query paramter required"})
		return
	}

	messages, err := h.service.GetChannelMessages(c.Request.Context(), channelID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

func (h *ForumUserHandler) CreateMessage(c *gin.Context) {
	channelID := c.Param("id")
	var req struct {
		UserID          string `json:"user_id"`
		Content         string `json:"content"`
		ParentMessageID string `json:"parent_message_id"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	message, err := h.service.CreateMessage(c.Request.Context(), channelID, req.UserID, req.Content, req.ParentMessageID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, message)
}

func (h *ForumUserHandler) MarkMessagesAsRead(c *gin.Context) {
	channelID := c.Param("id")
	var req struct {
		UserID string `json:"userId"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	err := h.service.MarkMessagesAsRead(c.Request.Context(), channelID, req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func (h *ForumUserHandler) GetOrCreateDirectMessageChannel(c *gin.Context) {
	user1ID := c.Query("user1ID")
	user2ID := c.Query("user2ID")

	if user1ID == "" || user2ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user1Id and user2Id query parameter required"})
		return
	}

	id, err := h.service.GetOrCreateDirectMessageChannel(c.Request.Context(), user1ID, user2ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"channelId": id})
}

func (h *ForumUserHandler) GetOnlineUsers(c *gin.Context) {
	userID := c.Query("userID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId query parameter required"})
		return
	}

	users, err := h.service.GetOnlineUsers(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *ForumUserHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	userID := c.Query("userId")
	if query == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q and userId query parameters required"})
		return
	}

	users, err := h.service.SearchUsers(c.Request.Context(), query, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *ForumUserHandler) GetUnreadCount(c *gin.Context) {
	userID := c.Param("id")

	result, err := h.service.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *ForumUserHandler) UpdateUserPresence(c *gin.Context) {
	userID := c.Param("id")
	var req struct {
		IsOnline bool `json:"isOnline"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	err := h.service.UpdateUserPresence(c.Request.Context(), userID, req.IsOnline)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, nil)
}

func (h *ForumUserHandler) GetChannelMembers(c *gin.Context) {
	channelID := c.Param("channelID")

	users, err := h.service.GetChannelMembers(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *ForumUserHandler) EditMessage(c *gin.Context) {
	messageID := c.Param("id")
	var req struct {
		UserID  string `json:"userId"`
		Content string `json:"content"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.EditMessage(c.Request.Context(), messageID, req.UserID, req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "message updated"})
}

func (h *ForumUserHandler) DeleteMessage(c *gin.Context) {
	messageID := c.Param("id")
	userID := c.Query("user_id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userID query parameter required"})
		return
	}

	err := h.service.DeleteMessage(c.Request.Context(), messageID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "message deleted"})
}

func (h *ForumUserHandler) AddReaction(c *gin.Context) {
	messageID := c.Param("id")
	var req struct {
		UserID string `json:"userId"`
		Emoji  string `json:"emoji"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	err := h.service.AddReaction(c.Request.Context(), messageID, req.UserID, req.Emoji)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, err)
}

func (h *ForumUserHandler) RemoveReaction(c *gin.Context) {
	messageID := c.Param("id")
	userID := c.Query("userId")
	emoji := c.Query("emoji")
	if userID == "" || emoji == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId and emoji query parameters required"})
		return
	}

	err := h.service.RemoveReaction(c.Request.Context(), messageID, userID, emoji)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, err)
}
