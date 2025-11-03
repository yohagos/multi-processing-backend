package db

import (
	"bufio"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDatabase(ctx context.Context, url string) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		panic(err)
	}
	if err = pool.Ping(ctx); err != nil {
		panic(err)
	}
	return pool
}

func ApplyMigrations(
	ctx context.Context,
	pool *pgxpool.Pool,
) error {
	files, err := listFiles("migrations/sql")
	if err != nil {
		return err
	}

	for _, file := range files {
		normalized := filepath.ToSlash(file)
		if err := applyMigrationFile(ctx, pool, normalized); err != nil {
			return err
		}
	}

	return nil
}

func listFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".sql") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func applyMigrationFile(ctx context.Context, pool *pgxpool.Pool, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var stmt strings.Builder

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
		stmt.WriteString(line + " ")
		if strings.HasSuffix(line, ";") {
			sql := strings.TrimSpace(stmt.String())
			if _, err := pool.Exec(ctx, sql); err != nil {
				return err
			}
			stmt.Reset()
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
