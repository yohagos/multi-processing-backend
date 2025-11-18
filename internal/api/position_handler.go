package api

import (
	"context"
	"multi-processing-backend/internal/core"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type PositionService interface {
	List(ctx context.Context, page, limit int) ([]core.Position, int64, error)
	Create(ctx context.Context, user core.Position) (core.Position, error)

	Get(ctx context.Context, id string) (core.Position, error)
	Update(ctx context.Context, id string, updates core.PositionUpdate) (core.Position, error)
	Delete(ctx context.Context, id string) error
}

type PositionHandler struct {
	service PositionService
}

func NewPositionHandler(service PositionService) *PositionHandler {
	return &PositionHandler{service: service}
}

func RegisterPositionRoutes(rg *gin.RouterGroup, h *PositionHandler) {
	pos := rg.Group("")
	{
		pos.GET("", h.List)
		pos.POST("", h.Create)
		pos.GET("/:id", h.Get)
		pos.PATCH("/:id", h.Update)
		pos.DELETE("/:id", h.Delete)
	}
}

func (h *PositionHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	pos, total, err := h.service.List(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := core.PositionPagination{
		Data:  pos,
		Total: total,
		Error: nil,
	}

	c.JSON(http.StatusOK, response)
}

func (h *PositionHandler) Create(c *gin.Context) {
	var req core.Position

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *PositionHandler) Get(c *gin.Context) {
	id := c.Param("id")
	user, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "position not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *PositionHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req core.PositionUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "position not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *PositionHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
