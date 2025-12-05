package db

import (
	"context"
	"encoding/json"
	"math"
	"math/rand"
	"multi-processing-backend/internal/core"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/exp/slog"
)

type CryptoRepository struct {
	pool *pgxpool.Pool
}

func NewCryptoRepository(pool *pgxpool.Pool) *CryptoRepository {
	return &CryptoRepository{pool: pool}
}

func (r *CryptoRepository) List(
	ctx context.Context,
	page, limit int,
) ([]core.Crypto, int64, error) {
	offset := (page - 1) * limit

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM crypto").Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, initial, name, current_value, previous_value, percent, created_at, is_initial
		FROM crypto
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	cryptos, err := pgx.CollectRows(rows, pgx.RowToStructByPos[core.Crypto])
	if err != nil {
		return nil, 0, err
	}

	return cryptos, total, nil
}

func (r *CryptoRepository) Create(
	ctx context.Context,
	c core.Crypto,
) (core.Crypto, error) {
	lastEntry, err := r.GetLatestByName(ctx, c.Name)
	var percent float64 = 0
	if err != nil {
		lastEntry.CurrentValue = 0
	} else {
		percent = ((c.CurrentValue - lastEntry.CurrentValue) / lastEntry.CurrentValue) * 100
		if math.IsNaN(percent) {
			percent = rand.Float64()
		}
	}
	query := `
		INSERT INTO crypto (initial, name, current_value, previous_value, percent)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, initial, name, current_value, previous_value, percent, created_at
	`
	var created core.Crypto
	err = r.pool.QueryRow(
		ctx, query, c.Initial, c.Name, c.CurrentValue, lastEntry.CurrentValue, percent,
	).Scan(
		&created.ID,
		&created.Initial,
		&created.Name,
		&created.CurrentValue,
		&created.PreviousValue,
		&created.Percent,
		&created.CreatedAt,
	)
	if err != nil {
		return core.Crypto{}, err
	}

	return c, nil
}

func (r *CryptoRepository) SeedCryptosIfEmpty(
	ctx context.Context,
	jsonPath string,
) error {
	var count int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM crypto").Scan(&count); err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}

	var cryptos []struct {
		Initial      string  `json:"initial"`
		Name         string  `json:"name"`
		CurrentValue float64 `json:"current_value"`
	}

	if err := json.Unmarshal(data, &cryptos); err != nil {
		return err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, cry := range cryptos {
		_, err := tx.Exec(ctx, `
			INSERT INTO crypto (initial, name, current_value, previous_value, percent, is_initial, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, cry.Initial, cry.Name, cry.CurrentValue, float64(0), float64(0), true, time.Now().UTC())
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (r *CryptoRepository) ListInitial(ctx context.Context, limit int) ([]core.Crypto, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, initial, name, current_value, previous_value, percent, created_at, is_initial
		FROM crypto
		WHERE is_initial = true
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByPos[core.Crypto])
}

func (r *CryptoRepository) GetLatestByName(ctx context.Context, name string) (core.Crypto, error) {
	var c core.Crypto
	err := r.pool.QueryRow(ctx, `
		SELECT id, initial, name, current_value, previous_value, percent, created_at, is_initial
		FROM crypto
		WHERE name = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, name).Scan(
		&c.ID, &c.Initial, &c.Name, &c.CurrentValue,
		&c.PreviousValue, &c.Percent, &c.CreatedAt, &c.IsInitial,
	)
	return c, err
}

func (r *CryptoRepository) DeleteDevData(ctx context.Context) {
	_, err := r.pool.Exec(ctx, `DROP TABLE crypto`)
	if err != nil {
		slog.Error("Could not delete Crypto Table before shutting down.", "DatabaseError", err)
		return
	}
	slog.Info("Dropped Crypto Table successfully")
}

func (r *CryptoRepository) GetLatestCryptos(ctx context.Context, limit int) ([]core.Crypto, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, initial, name, current_value, previous_value, percent, created_at, is_initial
		FROM crypto
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByPos[core.Crypto])
}
