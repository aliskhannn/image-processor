package file

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Storage provides an S3-compatible storage backend using MinIO.
// It stores files in a specified bucket under different subdirectories.
type Storage struct {
	client     *minio.Client
	bucketName string
}

// NewStorage creates a new Storage instance connected to the specified MinIO server.
// If the bucket does not exist, it will be created automatically.
func NewStorage(ctx context.Context, endpoint, accessKey, secretKey, bucketName string, useSSL bool) (*Storage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize minio client: %w", err)
	}

	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exists {
		if err := client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &Storage{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// Save uploads the provided file reader to the specified subdirectory in the bucket.
// Returns the object path within the bucket.
func (s *Storage) Save(ctx context.Context, subdir, filename string, src io.Reader) (string, error) {
	objectName := filepath.Join(subdir, filename)

	_, err := s.client.PutObject(ctx, s.bucketName, objectName, src, -1, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return objectName, nil
}

// Load retrieves the file from the specified subdirectory in the bucket and returns a reader.
func (s *Storage) Load(ctx context.Context, path string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucketName, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to load file: %w", err)
	}

	return obj, nil
}

// Delete removes the specified file from the bucket.
func (s *Storage) Delete(ctx context.Context, path string) error {
	return s.client.RemoveObject(ctx, s.bucketName, path, minio.RemoveObjectOptions{})
}
