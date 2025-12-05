package db

import (
	"context"

	"multi-processing-backend/internal/core"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AddressRepository struct {
	pool *pgxpool.Pool
}

func NewAddressRepository(pool *pgxpool.Pool) *AddressRepository {
	return &AddressRepository{pool: pool}
}

func (r *AddressRepository) List(
	ctx context.Context,
	page, limit int,
) ([]core.Address, int64, error) {
	offset := (page - 1) * limit

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM addresses").Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, street, city, zip_code, country, is_primary, created_at, updated_at
		FROM addresses
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	users, err := pgx.CollectRows(rows, pgx.RowToStructByPos[core.Address])
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *AddressRepository) Create(
	ctx context.Context,
	add core.Address,
) (core.Address, error) {
	query := `
		INSERT INTO addresses (user_id, street, city, zip_code, country, is_primary)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query, add.UserID, add.Street, add.City, add.ZipCode, add.Country, add.IsPrimary).Scan(&add.ID, &add.CreatedAt, &add.UpdatedAt)
	if err != nil {
		return core.Address{}, err
	}

	return add, nil
}

func (r *AddressRepository) Get(
	ctx context.Context,
	id string,
) (core.Address, error) {
	var add core.Address
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, street, city, zip_code, country, is_primary, created_at, updated_at
		FROM addresses 
		WHERE id = $1
	`, id).Scan(
		&add.ID, &add.UserID, &add.Street, &add.City, &add.ZipCode, &add.Country, &add.IsPrimary, &add.CreatedAt, &add.UpdatedAt,
	)
	if err != nil {
		return core.Address{}, err
	}
	return add, nil
}

func (r *AddressRepository) Update(
	ctx context.Context,
	id string,
	update core.AddressUpdate,
) (core.Address, error) {
	var add core.Address
	err := r.pool.QueryRow(ctx, `
		UPDATE addresses
		SET user_id = $1, street = $2, city = $3, zip_code = $4, country = $5, is_primary = true, updated_at = NOW()
		WHERE id = $6
		RETURNING id, user_id, street, city, zip_code, country, is_primary, created_at, updated_at
	`,
		update.UserID, update.Street, update.City, update.ZipCode, update.Country, id).Scan(
		&add.ID, &add.UserID, &add.Street, &add.City, &add.ZipCode,
		&add.Country, &add.IsPrimary, &add.CreatedAt, &add.UpdatedAt,
	)
	if err != nil {
		return core.Address{}, err
	}

	return add, nil
}

func (r *AddressRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}
