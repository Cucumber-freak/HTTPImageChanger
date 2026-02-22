package server

import (
	"io"
	"log"
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
	r.Get("/download/{id}", s.DownloadHandler)

	return r
}

func (s *Server) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Incorrect Method", http.StatusMethodNotAllowed)
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Can`t read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	taskId := uuid.New().String()
	ctx := r.Context()

	err = s.S3.Upload(ctx, taskId, file, header.Size)
	if err != nil {
		http.Error(w, "S3 error", http.StatusInternalServerError)
		return
	}

	err = s.DB.SaveImg(ctx, taskId, header.Filename)
	if err != nil {
		http.Error(w, "Db error", http.StatusInternalServerError)
		return
	}

	err = s.Broker.Publish(ctx, taskId)
	if err != nil {
		http.Error(w, "RabbitMQ error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(taskId))
}

func (s *Server) DownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Incorrect Method", http.StatusMethodNotAllowed)
		return
	}
	imageID := chi.URLParam(r, "id")
	ctx := r.Context()
	if imageID == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}
	compressedName := imageID + "_compressed"
	reader, err := s.S3.Download(ctx, compressedName)
	if err != nil {
		log.Printf("File not found in S3: %v", err)
		http.Error(w, "Image not ready or not found", http.StatusNotFound)
		return
	}
	defer reader.Close()
	w.Header().Set("Content-Disposition", "attachment; filename="+imageID+".jpg")

	_, err = io.Copy(w, reader)
	if err != nil {
		log.Printf("Error streaming file: %v", err)
	}
}
