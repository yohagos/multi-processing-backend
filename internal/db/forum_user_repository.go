package db

import "github.com/jackc/pgx/v5/pgxpool"

type ForumUserRepository struct {
	pool *pgxpool.Pool
}

func NewForumUserRepository(pool *pgxpool.Pool) *ForumUserRepository {
	return &ForumUserRepository{pool: pool}
}