package services

import (
	"context"

	"multi-processing-backend/internal/core"
)

type UserRepository interface {
	List(ctx context.Context, page, limit int) ([]core.User, int64, error)
	Create(ctx context.Context, u core.User) (core.User, error)
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