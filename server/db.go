package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func NewStorage(pool *pgxpool.Pool) *Storage {
	return &Storage{pool: pool}
}

func (s *Storage) CreateTable(ctx context.Context) error {
	sqlQuery := `
    CREATE TABLE IF NOT EXISTS images (
        id UUID PRIMARY KEY,
        file_name VARCHAR(255) NOT NULL,
        status VARCHAR(50) NOT NULL DEFAULT 'pending',
        created_at TIMESTAMPTZ DEFAULT NOW(),
        error_message TEXT 
    );`
	_, err := s.pool.Exec(ctx, sqlQuery)
	return err
}

func (s *Storage) SaveImg(ctx context.Context, id, fileName string) error {
	sqlQuery := `INSERT INTO images (id, file_name, status) VALUES ($1, $2, $3);`
	_, err := s.pool.Exec(ctx, sqlQuery, id, fileName, "pending")
	return err
}

func (s *Storage) GetImg(ctx context.Context, id string) (string, error) {
	var fileName string
	sqlQuery := `SELECT file_name FROM images WHERE id = $1;`
	err := s.pool.QueryRow(ctx, sqlQuery, id).Scan(&fileName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("image with id %s not found", id)
		}
		return "", err
	}
	return fileName, nil
}
