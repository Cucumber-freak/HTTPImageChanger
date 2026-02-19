package server

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
)

func (s *Server) StartWorker(ctx context.Context) {
	msgs, _ := s.Broker.Channel.Consume(s.Broker.Queue, "", false, false, false, false, nil)
	for {
		select {
		case <-ctx.Done():

			return
		case f, ok := <-msgs:
			if !ok {
				return
			}

			imageID := string(f.Body)
			s.DB.UpdateStatus(ctx, imageID, "processing")
			reader, err := s.S3.Download(imageID)
			if err != nil {
				log.Printf("Error downloading %s: %v", imageID, err)
				f.Nack(false, true)
				continue
			}

			changedImage, size, err := s.ChangeImage(reader)
			reader.Close()
			if err != nil {
				log.Printf("Error processing %s: %v", imageID, err)
				s.DB.UpdateStatus(ctx, imageID, "error")
				f.Ack(false)
				continue
			}
			err = s.S3.Upload(imageID+"_compressed", changedImage, size)
			if err != nil {
				log.Printf("Error uploading %s: %v", imageID, err)
				f.Nack(false, true)
				continue
			}
			s.DB.UpdateStatus(ctx, imageID, "completed")
			f.Ack(false)
			log.Printf("Task %s completed!", imageID)
		}
	}
}

func (s *Server) ChangeImage(imgFile io.ReadCloser) (io.Reader, int64, error) {
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return nil, 0, fmt.Errorf("decode error: %w", err)
	}
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 60})
	if err != nil {
		return nil, 0, fmt.Errorf("encode error: %w", err)
	}
	return buf, int64(buf.Len()), nil
}
