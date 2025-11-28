package api

import (
	"context"
	"multi-processing-backend/internal/core"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
)

type SkillService interface {
	List(ctx context.Context) ([]core.Skill, int64, error)
	Create(ctx context.Context, user core.Skill) (core.Skill, error)
	AddSkillByUserId(ctx context.Context, skill_id, user_id string) error
	Get(ctx context.Context, id string) (core.Skill, error)
	GetByUserId(ctx context.Context, id string) ([]core.SkillWithDetails, error)
	Update(ctx context.Context, id string, updates core.SkillUpdate) (core.Skill, error)
	Delete(ctx context.Context, id string) error
	DeleteSkillByUserId(ctx context.Context, skill_id, user_id string) error
}

type SkillHandler struct {
	service SkillService
}

func NewSkillHandler(service SkillService) *SkillHandler {
	return &SkillHandler{service: service}
}

func RegisterSkillRoutes(rg *gin.RouterGroup, h *SkillHandler) {
	skill := rg.Group("")
	{
		skill.GET("", h.List)
		skill.GET("/:id", h.Get)
		skill.GET("/user/:id", h.GetByUserId)

		skill.POST("", h.Create)
		skill.POST("/add/:user_id/skill/:skill_id", h.AddSkillByUserId)

		skill.PATCH("/:id", h.Update)

		skill.DELETE("/:id", h.Delete)
		skill.DELETE("/delete/:user_id/skill/:skill_id", h.DeleteSkillByUserId)
	}
}

func (h *SkillHandler) List(c *gin.Context) {
	sals, total, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := core.SkillPagination{
		Data:  sals,
		Total: total,
		Error: nil,
	}

	c.JSON(http.StatusOK, response)
}

func (h *SkillHandler) Create(c *gin.Context) {
	var req core.Skill

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	skill, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, skill)
}

func (h *SkillHandler) AddSkillByUserId(c *gin.Context) {
	skill_id := c.Param("skill_id")
	user_id := c.Param("user_id")

	if err := h.service.AddSkillByUserId(c.Request.Context(), skill_id, user_id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusAccepted)
}

func (h *SkillHandler) Get(c *gin.Context) {
	id := c.Param("id")
	user, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "skill not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *SkillHandler) GetByUserId(c *gin.Context) {
	id := c.Param("id")
	skills, err := h.service.GetByUserId(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "skills for user not found"})
		return
	}
	c.JSON(http.StatusOK, skills)
}

func (h *SkillHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req core.SkillUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "skill not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *SkillHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *SkillHandler) DeleteSkillByUserId(c *gin.Context) {
	skill_id := c.Param("skill_id")
	user_id := c.Param("user_id")

	slog.Info("Delete Skill by user id", "skill_id", skill_id, "user_id", user_id)

	if err := h.service.DeleteSkillByUserId(c.Request.Context(), skill_id, user_id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusAccepted)
}
