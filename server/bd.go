package server

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type Storage struct {
	conn *pgx.Conn
}

func NewStorage(conn *pgx.Conn) *Storage {
	return &Storage{conn: conn}
}

func (s *Storage) CreateTable(ctx context.Context) error {
	sqlQuery :=
		`
CREATE TABLE IF NOT EXISTS orders (
        id VARCHAR(15) NOT NULL UNIQUE,
        file_Name VARCHAR(255) NOT NULL UNIQUE,
        status VARCHAR(50) NOT NULL,
		created_Time TIMESTAMPTZ DEFAULT NOW(),
		error_Message TEXT 
    );
`
	_, err := s.conn.Exec(ctx, sqlQuery)
	return err
}

func (s *Storage) SaveImg(ctx context.Context, id, file_Name string) error {
	sqlQuery := `INSERT INTO links(id, file_Name) VALUES ($1, $2);`
	_, err := s.conn.Exec(ctx, sqlQuery, id, file_Name)
	return err
}

func (s *Storage) GetImg(ctx context.Context, id string) (string, error) {
	var file_Name string
	sqlQuery := `SELECT file_Name FROM orders WHERE id = $1;`
	err := s.conn.QueryRow(ctx, sqlQuery, id).Scan(&file_Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.New("link dont founded")
		}
		return "", err
	}
	return file_Name, err
}
