package main

import (
	"PIZDEC/server"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	connStr := "postgres://user:password@localhost:5432/dbname"
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("Не удалось создать пул соединений: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("База данных недоступна: %v", err)
	}

	storage := server.NewStorage(pool)

	if err := storage.CreateTable(ctx); err != nil {
		log.Fatalf("Ошибка при создании таблицы: %v", err)
	}

	s3Client := server.ConnectS3("localhost:9000", "minioadmin", "minioadmin", "images")
	rabbitClient := server.ConnectRabbit("amqp://guest:guest@localhost:5672/", "task_queue")

	srv := &server.Server{
		DB:     storage,
		S3:     s3Client,
		Broker: rabbitClient,
	}

	log.Println("Сервер запускается на порту :8080...")
	err = http.ListenAndServe(":8080", srv.NewConnection())
	if err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
