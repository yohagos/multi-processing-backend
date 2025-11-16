package services

import (
	"context"
	"multi-processing-backend/internal/core"
)

type PositionRepository interface {
	List(ctx context.Context, page, limit int) ([]core.Position, int64, error)
	Create(ctx context.Context, u core.Position) (core.Position, error)

	Get(ctx context.Context, id string) (core.Position, error)
	Update(ctx context.Context, id string, update core.PositionUpdate) (core.Position, error)
	Delete(ctx context.Context, id string) error
}

type PositionService struct {
	repo PositionRepository
}

func NewPositionService(repo PositionRepository) *PositionService {
	return &PositionService{repo: repo}
}

func (s *PositionService) List(ctx context.Context, page, limit int) ([]core.Position, int64, error) {
	return s.repo.List(ctx, page, limit)
}

func (s *PositionService) Create(
	ctx context.Context,
	user core.Position,
) (core.Position, error) {
	return s.repo.Create(ctx, user)
}

func (s *PositionService) Get(ctx context.Context, id string) (core.Position, error) {
	return s.repo.Get(ctx, id)
}

func (s *PositionService) Update(ctx context.Context, id string, updates core.PositionUpdate) (core.Position, error) {
	return s.repo.Update(ctx, id, updates)
}

func (s *PositionService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}