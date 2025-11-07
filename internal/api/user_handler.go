package api

import (
	"context"
	"net/http"
	"strconv"

	"multi-processing-backend/internal/core"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
)

type UserService interface {
	List(ctx context.Context, page, limit int) ([]core.User, int64, error)
	Create(ctx context.Context, email, firstName, lastName string) (core.User, error)
	
	Get(ctx context.Context, id string) (core.User, error)
	Update(ctx context.Context, id string, updates core.UserUpdate) (core.User, error)
	Delete(ctx context.Context, id string) error
}

type UserHandler struct {
	service UserService
}

func NewUserHandler(service UserService) *UserHandler {
	return &UserHandler{service: service}
}

func RegisterUserRoutes(rg *gin.RouterGroup, h *UserHandler) {
	users := rg.Group("")
	{
		users.GET("", h.ListUsers)
		users.POST("", h.CreateUser)
		users.GET("/:id", h.GetUserById)
		users.PATCH("/:id", h.UpdateUser)
		users.DELETE("/:id", h.DeleteUserById)
	}
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	slog.Info("ListUsers from user_handler", "GET", "list of users", "Query", c.Query("page"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	users, total, err := h.service.List(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req struct {
		Email     string `json:"email" binding:"required,email"`
		FirstName string `json:"first_name" binding:"required"`
		LastName  string `json:"last_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.Create(c.Request.Context(), req.Email, req.FirstName, req.LastName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) GetUserById(c *gin.Context) {
	id := c.Param("id")
	user, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req core.UserUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *UserHandler) DeleteUserById(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
