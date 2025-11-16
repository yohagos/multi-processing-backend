package services

import (
	"context"
	"multi-processing-backend/internal/core"
)

type SalaryRepository interface {
	List(ctx context.Context, page, limit int) ([]core.Salary, int64, error)
	Create(ctx context.Context, u core.Salary) (core.Salary, error)

	Get(ctx context.Context, id string) (core.Salary, error)
	Update(ctx context.Context, id string, update core.SalaryUpdate) (core.Salary, error)
	Delete(ctx context.Context, id string) error
}

type SalaryService struct {
	repo SalaryRepository
}

func NewSalaryService(repo SalaryRepository) *SalaryService {
	return &SalaryService{repo: repo}
}

func (s *SalaryService) List(ctx context.Context, page, limit int) ([]core.Salary, int64, error) {
	return s.repo.List(ctx, page, limit)
}

func (s *SalaryService) Create(
	ctx context.Context,
	user core.Salary,
) (core.Salary, error) {
	return s.repo.Create(ctx, user)
}

func (s *SalaryService) Get(ctx context.Context, id string) (core.Salary, error) {
	return s.repo.Get(ctx, id)
}

func (s *SalaryService) Update(ctx context.Context, id string, updates core.SalaryUpdate) (core.Salary, error) {
	return s.repo.Update(ctx, id, updates)
}

func (s *SalaryService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}