// internal/adapter/service/minio/minio_adapter.go
package minioadapter // Use a distinct name like minioadapter or minioservice

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/yvanyang/language-learning-player-api/internal/config" // Adjust import path
	"github.com/yvanyang/language-learning-player-api/internal/port"   // Adjust import path
)

// MinioStorageService implements the port.FileStorageService interface using MinIO.
type MinioStorageService struct {
	client        *minio.Client
	defaultBucket string
	defaultExpiry time.Duration
	logger        *slog.Logger
}

// NewMinioStorageService creates a new MinioStorageService.
func NewMinioStorageService(cfg config.MinioConfig, logger *slog.Logger) (*MinioStorageService, error) {
	if cfg.Endpoint == "" || cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" || cfg.BucketName == "" {
		return nil, fmt.Errorf("minio configuration (Endpoint, AccessKeyID, SecretAccessKey, BucketName) cannot be empty")
	}

	log := logger.With("service", "MinioStorageService")
	log.Info("Initializing MinIO client", "endpoint", cfg.Endpoint, "ssl", cfg.UseSSL, "bucket", cfg.BucketName)

	// Initialize minio client object.
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		log.Error("Failed to initialize MinIO client", "error", err)
		return nil, fmt.Errorf("minio client initialization failed: %w", err)
	}

	// Optional: Ping MinIO to check connectivity (MinIO server needs to be running)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	exists, err := minioClient.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		log.Error("Failed to check if MinIO bucket exists", "error", err, "bucket", cfg.BucketName)
		// Decide if this is fatal or not. Log warning and proceed.
	}
	if !exists {
		log.Warn("Default MinIO bucket does not exist. Consider creating it.", "bucket", cfg.BucketName)
	} else {
		log.Info("MinIO bucket found", "bucket", cfg.BucketName)
	}

	return &MinioStorageService{
		client:        minioClient,
		defaultBucket: cfg.BucketName,
		defaultExpiry: cfg.PresignExpiry, // Use expiry from config
		logger:        log,
	}, nil
}

// GetPresignedGetURL generates a temporary URL for downloading an object.
func (s *MinioStorageService) GetPresignedGetURL(ctx context.Context, bucket, objectKey string, expiry time.Duration) (string, error) {
	if bucket == "" {
		bucket = s.defaultBucket
	}
	if expiry <= 0 {
		expiry = s.defaultExpiry
	}
	reqParams := make(url.Values)

	presignedURL, err := s.client.PresignedGetObject(ctx, bucket, objectKey, expiry, reqParams)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to generate presigned GET URL", "error", err, "bucket", bucket, "key", objectKey)
		return "", fmt.Errorf("failed to get presigned URL for %s/%s: %w", bucket, objectKey, err)
	}

	s.logger.DebugContext(ctx, "Generated presigned GET URL", "bucket", bucket, "key", objectKey, "expiry", expiry)
	return presignedURL.String(), nil
}

// DeleteObject removes an object from MinIO storage.
func (s *MinioStorageService) DeleteObject(ctx context.Context, bucket, objectKey string) error {
	if bucket == "" {
		bucket = s.defaultBucket
	}
	opts := minio.RemoveObjectOptions{}

	err := s.client.RemoveObject(ctx, bucket, objectKey, opts)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to delete object from MinIO", "error", err, "bucket", bucket, "key", objectKey)
		return fmt.Errorf("failed to delete object %s/%s: %w", bucket, objectKey, err)
	}

	s.logger.InfoContext(ctx, "Deleted object from MinIO", "bucket", bucket, "key", objectKey)
	return nil
}

// GetPresignedPutURL generates a temporary URL for uploading an object.
func (s *MinioStorageService) GetPresignedPutURL(ctx context.Context, bucket, objectKey string, expiry time.Duration) (string, error) {
	if bucket == "" {
		bucket = s.defaultBucket
	}
	if expiry <= 0 {
		expiry = s.defaultExpiry
	}

	// policy := minio.NewPostPolicy() // Note: Using PostPolicy for Put? Check SDK if PresignedPutObject directly is better. Let's stick to PresignedPutObject for simplicity.
	// policy.SetBucket(bucket)
	// policy.SetKey(objectKey)
	// policy.SetExpires(time.Now().UTC().Add(expiry))

	presignedURL, err := s.client.PresignedPutObject(ctx, bucket, objectKey, expiry)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to generate presigned PUT URL", "error", err, "bucket", bucket, "key", objectKey)
		return "", fmt.Errorf("failed to get presigned PUT URL for %s/%s: %w", bucket, objectKey, err)
	}

	s.logger.DebugContext(ctx, "Generated presigned PUT URL", "bucket", bucket, "key", objectKey, "expiry", expiry)
	return presignedURL.String(), nil
}

// ObjectExists checks if an object exists in the specified bucket using StatObject.
// ADDED: Implementation for ObjectExists
func (s *MinioStorageService) ObjectExists(ctx context.Context, bucket, objectKey string) (bool, error) {
	if bucket == "" {
		bucket = s.defaultBucket
	}

	_, err := s.client.StatObject(ctx, bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		// Check if the error code indicates that the object was not found.
		if errResponse.Code == "NoSuchKey" {
			s.logger.DebugContext(ctx, "Object check: Object not found", "bucket", bucket, "key", objectKey)
			return false, nil // Object does not exist, but this is not an error in checking
		}
		// For any other error (network issue, permissions, etc.), log and return the error.
		s.logger.ErrorContext(ctx, "Failed to stat object", "error", err, "bucket", bucket, "key", objectKey)
		return false, fmt.Errorf("failed to check object existence for %s/%s: %w", bucket, objectKey, err)
	}
	// If StatObject returns no error, the object exists.
	s.logger.DebugContext(ctx, "Object check: Object found", "bucket", bucket, "key", objectKey)
	return true, nil
}

// Compile-time check to ensure MinioStorageService satisfies the port.FileStorageService interface
var _ port.FileStorageService = (*MinioStorageService)(nil)
