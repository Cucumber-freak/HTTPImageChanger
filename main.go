package main

import (
	"PIZDEC/server"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, "postgres://user:password@localhost:5432/photo_service")
	if err != nil {
		log.Fatal(err)
	}
	storage := server.NewStorage(pool)
	s3Client := server.ConnectS3("localhost:9000", "admin", "admin", "images")
	rabbitClient := server.ConnectRabbit("amqp://guest:guest@localhost:5672/", "task_queue")
	srv := &server.Server{
		DB:     storage,
		S3:     s3Client,
		Broker: rabbitClient,
	}

	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			srv.StartWorker(ctx)
		}()
	}
	httpServer := &http.Server{Addr: ":8080", Handler: srv.NewConnection()}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Stop work")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	httpServer.Shutdown(shutdownCtx)

	wg.Wait()
	log.Println("All workers and work!")
}
