package api

import (
	"context"
	"multi-processing-backend/internal/core"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type DepartmentService interface {
	List(ctx context.Context, searchName string) ([]core.Departments, int64, error)
	Create(ctx context.Context, user core.Departments) (core.Departments, error)

	Get(ctx context.Context, id string) (core.Departments, error)
	Update(ctx context.Context, id string, updates core.DepartmentUpdate) (core.Departments, error)
	Delete(ctx context.Context, id string) error
}

type DepartmentHandler struct {
	service DepartmentService
}

func NewDepartmentHandler(service DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{service: service}
}

func RegisterDepartmentRoutes(rg *gin.RouterGroup, h *DepartmentHandler) {
	deps := rg.Group("")
	{
		deps.GET("", h.List)
		deps.POST("", h.Create)
		deps.GET("/:id", h.Get)
		deps.PATCH("/:id", h.Update)
		deps.DELETE("/:id", h.Delete)
	}
}

func (h *DepartmentHandler) List(c *gin.Context) {
	searchName := c.Query("searchName")
	deps, total, err := h.service.List(c.Request.Context(), searchName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := core.DepartmentsPagination{
		Data:  deps,
		Total: total,
		Error: err,
	}

	c.JSON(http.StatusOK, response)
}

func (h *DepartmentHandler) Create(c *gin.Context) {
	var req core.Departments

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dep, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dep)
}

func (h *DepartmentHandler) Get(c *gin.Context) {
	id := c.Param("id")
	dep, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dep)
}

func (h *DepartmentHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req core.DepartmentUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "department not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *DepartmentHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
