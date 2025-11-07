package services

import (
	"context"

	"multi-processing-backend/internal/core"
)

type UserRepository interface {
	List(ctx context.Context, page, limit int) ([]core.User, int64, error)
	Create(ctx context.Context, u core.User) (core.User, error)

	Get(ctx context.Context, id string) (core.User, error)
	Update(ctx context.Context, id string, updates core.UserUpdate) (core.User, error)
	Delete(ctx context.Context, id string) error
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) List(ctx context.Context, page, limit int) ([]core.User, int64, error) {
	return s.repo.List(ctx, page, limit)
}

func (s *UserService) Create(ctx context.Context, email, firstName, lastName string) (core.User, error) {
	user := core.User{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	}
	return s.repo.Create(ctx, user)
}

func (s *UserService) Get(ctx context.Context, id string) (core.User, error) {
	return s.repo.Get(ctx, id)
}

func (s *UserService) Update(ctx context.Context, id string, updates core.UserUpdate) (core.User, error) {
	return s.repo.Update(ctx, id, updates)
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
