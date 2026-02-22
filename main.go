package main

import (
	"PIZDEC/config"
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
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	storage := server.NewStorage(pool)
	err = storage.CreateTable(ctx)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
	s3Client := server.ConnectS3(ctx, cfg.S3Endpoint, cfg.S3Key, cfg.S3Secret, cfg.S3Bucket)
	rabbitClient := server.ConnectRabbit(cfg.RabbitURL, cfg.QueueName)
	defer rabbitClient.Close()
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
	httpServer := &http.Server{Addr: cfg.Port, Handler: srv.NewConnection()}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Stop work")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = httpServer.Shutdown(shutdownCtx)
	if err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}

	wg.Wait()
	log.Println("All workers and work!")
}
