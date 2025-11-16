package services

import (
	"context"

	"multi-processing-backend/internal/core"
)

type AddressRepository interface {
	List(ctx context.Context, page, limit int) ([]core.Address, int64, error)
	Create(ctx context.Context, u core.Address) (core.Address, error)

	Get(ctx context.Context, id string) (core.Address, error)
	Update(ctx context.Context, id string, update core.AddressUpdate) (core.Address, error)
	Delete(ctx context.Context, id string) error
}

type AddressService struct {
	repo AddressRepository
}

func NewAddressService(repo AddressRepository) *AddressService {
	return &AddressService{repo: repo}
}

func (s *AddressService) List(ctx context.Context, page, limit int) ([]core.Address, int64, error) {
	return s.repo.List(ctx, page, limit)
}

func (s *AddressService) Create(
	ctx context.Context, 
	add core.Address,
) (core.Address, error) {
	return s.repo.Create(ctx, add)
}

func (s *AddressService) Get(ctx context.Context, id string) (core.Address, error) {
	return s.repo.Get(ctx, id)
}

func (s *AddressService) Update(ctx context.Context, id string, updates core.AddressUpdate) (core.Address, error) {
	return s.repo.Update(ctx, id, updates)
}

func (s *AddressService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
