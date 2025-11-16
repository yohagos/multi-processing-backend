package services

import (
	"context"
	"multi-processing-backend/internal/core"
)

type DepartmentRepository interface {
	List(ctx context.Context, page, limit int) ([]core.Departments, int64, error)
	Create(ctx context.Context, u core.Departments) (core.Departments, error)

	Get(ctx context.Context, id string) (core.Departments, error)
	Update(ctx context.Context, id string, update core.DepartmentUpdate) (core.Departments, error)
	Delete(ctx context.Context, id string) error
}

type DepartmentService struct {
	repo DepartmentRepository
}

func NewDepartmentService(repo DepartmentRepository) *DepartmentService {
	return &DepartmentService{repo: repo}
}

func (s *DepartmentService) List(
	ctx context.Context,
	page, limit int,
) ([]core.Departments, int64, error) {
	return s.repo.List(ctx, page, limit)
}

func (s *DepartmentService) Create(
	ctx context.Context,
	dep core.Departments,
) (core.Departments, error) {
	/* dep := core.Departments{
		Name: name,
		Description: description,
	} */
	return s.repo.Create(ctx, dep)
}

func (s *DepartmentService) Get(
	ctx context.Context,
	id string,
) (core.Departments, error) {
	return s.repo.Get(ctx, id)
}

func (s *DepartmentService) Update(
	ctx context.Context,
	id string,
	update core.DepartmentUpdate,
) (core.Departments, error) {
	return s.repo.Update(ctx, id, update)
}

func (s *DepartmentService) Delete(
	ctx context.Context,
	id string,
) error {
	return s.repo.Delete(ctx, id)
}
