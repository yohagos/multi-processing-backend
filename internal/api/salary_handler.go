package api

import (
	"context"
	"multi-processing-backend/internal/core"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type SalaryService interface {
	List(ctx context.Context, page, limit int) ([]core.Salary, int64, error)
	Create(ctx context.Context, user core.Salary) (core.Salary, error)

	Get(ctx context.Context, id string) (core.Salary, error)
	Update(ctx context.Context, id string, updates core.SalaryUpdate) (core.Salary, error)
	Delete(ctx context.Context, id string) error
}

type SalaryHandler struct {
	service SalaryService
}

func NewSalaryHandler(service SalaryService) *SalaryHandler {
	return &SalaryHandler{service: service}
}

func RegisterSalaryRoutes(rg *gin.RouterGroup, h *SalaryHandler) {
	salary := rg.Group("")
	{
		salary.GET("", h.List)
		salary.POST("", h.Create)
		salary.GET("/:id", h.Get)
		salary.PATCH("/:id", h.Update)
		salary.DELETE("/:id", h.Delete)
	}
}

func (h *SalaryHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	sals, total, err := h.service.List(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := core.SalaryPagination{
		Data:  sals,
		Total: total,
		Error: nil,
	}

	c.JSON(http.StatusOK, response)
}

func (h *SalaryHandler) Create(c *gin.Context) {
	var req core.Salary

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

func (h *SalaryHandler) Get(c *gin.Context) {
	id := c.Param("id")
	user, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "salary not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *SalaryHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req core.SalaryUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "salary not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *SalaryHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
