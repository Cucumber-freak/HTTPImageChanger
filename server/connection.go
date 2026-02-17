package server

import (
	"context"
	"net/http"

	chi "github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/google/uuid"
)

type Server struct {
	DB     *Storage
	S3     *S3Client
	Broker *RabbitClient
}

func (s *Server) NewConnection() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/upload", s.UploadHandler)

	return r
}

func (s *Server) UploadHandler(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Can`t read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	taskId := uuid.New().String()
	var ctx context.Context
	s.DB.SaveImg(ctx, taskId, header.Filename)
	if err != nil {
		http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		return
	}
}
