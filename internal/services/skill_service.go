package services

import (
	"context"
	"multi-processing-backend/internal/core"

	"golang.org/x/exp/slog"
)

type SkillRepository interface {
	List(ctx context.Context, page, limit int) ([]core.Skill, int64, error)
	Create(ctx context.Context, u core.Skill) (core.Skill, error)

	Get(ctx context.Context, id string) (core.Skill, error)
	GetByUserId(ctx context.Context, id string) ([]core.SkillWithDetails, error)
	Update(ctx context.Context, id string, update core.SkillUpdate) (core.Skill, error)
	Delete(ctx context.Context, id string) error
}

type SkillService struct {
	repo SkillRepository
}

func NewSkillService(repo SkillRepository) *SkillService {
	return &SkillService{repo: repo}
}

func (s *SkillService) List(ctx context.Context, page, limit int) ([]core.Skill, int64, error) {
	return s.repo.List(ctx, page, limit)
}

func (s *SkillService) Create(
	ctx context.Context,
	skill core.Skill,
) (core.Skill, error) {
	return s.repo.Create(ctx, skill)
}

func (s *SkillService) Get(ctx context.Context, id string) (core.Skill, error) {
	return s.repo.Get(ctx, id)
}

func (s *SkillService) GetByUserId(ctx context.Context, id string) ([]core.SkillWithDetails, error) {
	slog.Info("SkillService | Get Skills by UserID", "ID", id)
	return s.repo.GetByUserId(ctx, id)
}

func (s *SkillService) Update(ctx context.Context, id string, updates core.SkillUpdate) (core.Skill, error) {
	return s.repo.Update(ctx, id, updates)
}

func (s *SkillService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
