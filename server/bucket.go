package server

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Client struct {
	Client     *minio.Client
	BucketName string
}

func ConnectS3(ctx context.Context, endpoint, accessKey, secretKey, bucket string) *S3Client {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		panic(fmt.Sprintf("Can't connect to MinIO: %v", err))
	}

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		panic(fmt.Sprintf("Error checking bucket %s: %v", bucket, err))
	}

	if !exists {
		log.Printf("Bucket %s not found, creating...", bucket)
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			panic(fmt.Sprintf("Failed to create bucket %s: %v", bucket, err))
		}
		log.Printf("Bucket %s created successfully", bucket)
	}

	return &S3Client{Client: client, BucketName: bucket}
}

func (s *S3Client) Upload(ctx context.Context, objectName string, reader io.Reader, size int64) error {
	_, err := s.Client.PutObject(ctx, s.BucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	return err
}

func (s *S3Client) Download(ctx context.Context, n string) (io.ReadCloser, error) {
	object, err := s.Client.GetObject(
		ctx,
		s.BucketName,
		n,
		minio.GetObjectOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("can't get object")
	}
	_, err = object.Stat()
	if err != nil {
		return nil, fmt.Errorf("can,t find file in S3: %w", err)
	}

	return object, nil
}
