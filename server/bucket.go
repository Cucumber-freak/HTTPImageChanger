package server

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Client struct {
	Client     *minio.Client
	BucketName string
}

func ConnectS3(endpoint, accessKey, secretKey, bucket string) *S3Client {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		panic(fmt.Sprintf("Ошибка подключения к MinIO: %v", err))
	}
	return &S3Client{Client: client, BucketName: bucket}
}

func (s *S3Client) Upload(objectName string, reader io.Reader, size int64) error {
	_, err := s.Client.PutObject(context.Background(), s.BucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	return err
}
