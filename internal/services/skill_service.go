package services

import (
	"context"
	"multi-processing-backend/internal/core"
)

type SkillRepository interface {
	List(ctx context.Context) ([]core.Skill, int64, error)
	Create(ctx context.Context, u core.Skill) (core.Skill, error)
	AddSkillByUserId(ctx context.Context, skill_id, user_id string) error
	Get(ctx context.Context, id string) (core.Skill, error)
	GetByUserId(ctx context.Context, id string) ([]core.SkillWithDetails, error)
	Update(ctx context.Context, id string, update core.SkillUpdate) (core.Skill, error)
	Delete(ctx context.Context, id string) error
	DeleteSkillByUserId(ctx context.Context, skill_id, user_id string) error
}

type SkillService struct {
	repo SkillRepository
}

func NewSkillService(repo SkillRepository) *SkillService {
	return &SkillService{repo: repo}
}

func (s *SkillService) List(ctx context.Context) ([]core.Skill, int64, error) {
	return s.repo.List(ctx)
}

func (s *SkillService) Create(
	ctx context.Context,
	skill core.Skill,
) (core.Skill, error) {
	return s.repo.Create(ctx, skill)
}

func (s *SkillService) AddSkillByUserId(
	ctx context.Context,
	skill_id, user_id string,
) error {
	return s.repo.AddSkillByUserId(ctx, skill_id, user_id)
}

func (s *SkillService) DeleteSkillByUserId(
	ctx context.Context,
	skill_id, user_id string,
) error {
	return s.repo.DeleteSkillByUserId(ctx, skill_id, user_id)
}

func (s *SkillService) Get(ctx context.Context, id string) (core.Skill, error) {
	return s.repo.Get(ctx, id)
}

func (s *SkillService) GetByUserId(ctx context.Context, id string) ([]core.SkillWithDetails, error) {
	return s.repo.GetByUserId(ctx, id)
}

func (s *SkillService) Update(ctx context.Context, id string, updates core.SkillUpdate) (core.Skill, error) {
	return s.repo.Update(ctx, id, updates)
}

func (s *SkillService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
