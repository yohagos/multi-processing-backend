package services

import (
	"context"
	"math"

	"log/slog"
	"time"

	"math/rand"

	"multi-processing-backend/internal/core"
)

type CryptoRepository interface {
	List(ctx context.Context, page, limit int) ([]core.Crypto, int64, error)
	Create(ctx context.Context, c core.Crypto) (core.Crypto, error)

	ListInitial(ctx context.Context, limit int) ([]core.Crypto, error)
	GetLatestByName(ctx context.Context, name string) (core.Crypto, error)
	GetLatestCryptos(ctx context.Context, limit int) ([]core.Crypto, error)
}

type CryptoService struct {
	repo CryptoRepository
}

func NewCryptoService(repo CryptoRepository) *CryptoService {
	return &CryptoService{repo: repo}
}

func (s *CryptoService) List(ctx context.Context, page, limit int) ([]core.Crypto, int64, error) {
	return s.repo.List(ctx, page, limit)
}

func (s *CryptoService) Create(
	ctx context.Context,
	initial string,
	name string,
	current_value float64,
) (core.Crypto, error) {
	crypto := core.Crypto{
		Initial:      initial,
		Name:         name,
		CurrentValue: current_value,
	}
	return s.repo.Create(ctx, crypto)
}

func (s *CryptoService) StartPriceTicker(ctx context.Context) {
	initial, err := s.repo.ListInitial(ctx, 5)
	if err != nil {
		slog.Error("failed to load initial cryptos for ticker", "error", err)
		return
	}
	if len(initial) == 0 {
		slog.Warn("no initial cryptos found, ticker stopped")
		return
	}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("crypto price ticker stopped")
			return
		case <-ticker.C:
			s.updateRandomCrypto(ctx, initial)
		}
	}
}

func (s *CryptoService) updateRandomCrypto(ctx context.Context, initial []core.Crypto) {
	base := initial[rand.Intn(len(initial))]

	latest, err := s.repo.GetLatestByName(ctx, base.Name)
	if err != nil {
		slog.Error("failed to get latest crypto", "name", base, "error", err)
		return
	}

	changePrt := (rand.Float64() * 400) - 200
	changeFactor := 1 + (changePrt / 100)

	newValue := latest.CurrentValue * changeFactor

	newCrypto := core.Crypto{
		Name:          latest.Name,
		Initial:       latest.Initial,
		CurrentValue:  roundFloat(newValue, 4),
		PreviousValue: latest.CurrentValue,
		Percent:       roundFloat(changePrt, 4),
		CreatedAt:     time.Now().UTC(),
		IsInitial:     false,
	}

	_, err = s.repo.Create(ctx, newCrypto)
	if err != nil {
		slog.Error("failed to create new crypto price", "name", base, "error", err)
		return
	}
}

func roundFloat(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func (s *CryptoService) GetLatestCryptos(ctx context.Context, limit int) ([]core.Crypto, error) {
	return s.repo.GetLatestCryptos(ctx, limit)
}