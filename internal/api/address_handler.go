package api

import (
	"context"
	"multi-processing-backend/internal/core"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type AddressService interface {
	List(ctx context.Context, page, limit int) ([]core.Address, int64, error)
	Create(ctx context.Context, user core.Address) (core.Address, error)

	Get(ctx context.Context, id string) (core.Address, error)
	Update(ctx context.Context, id string, updates core.AddressUpdate) (core.Address, error)
	Delete(ctx context.Context, id string) error
}

type AddressHandler struct {
	service AddressService
}

func NewAddressHandler(service AddressService) *AddressHandler {
	return &AddressHandler{service: service}
}

func RegisterAddressRoutes(rg *gin.RouterGroup, h *AddressHandler) {
	pos := rg.Group("")
	{
		pos.GET("", h.List)
		pos.POST("", h.Create)
		pos.GET("/:id", h.Get)
		pos.PATCH("/:id", h.Update)
		pos.DELETE("/:id", h.Delete)
	}
}

func (h *AddressHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	pos, total, err := h.service.List(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := core.AddressPagination{
		Data:  pos,
		Total: total,
		Error: nil,
	}

	c.JSON(http.StatusOK, response)
}

func (h *AddressHandler) Create(c *gin.Context) {
	var req core.Address

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

func (h *AddressHandler) Get(c *gin.Context) {
	id := c.Param("id")
	user, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "position not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *AddressHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req core.AddressUpdate
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

func (h *AddressHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
