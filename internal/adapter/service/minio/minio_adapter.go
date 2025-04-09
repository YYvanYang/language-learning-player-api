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
	// Using a short context timeout for the check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Check if the default bucket exists. Create if not? (Consider permissions)
	exists, err := minioClient.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		log.Error("Failed to check if MinIO bucket exists", "error", err, "bucket", cfg.BucketName)
		// Decide if this is fatal or not. Maybe log warning and proceed?
		// return nil, fmt.Errorf("failed to check minio bucket %s: %w", cfg.BucketName, err)
	}
	if !exists {
		log.Warn("Default MinIO bucket does not exist. Consider creating it.", "bucket", cfg.BucketName)
		// Optionally create the bucket here if desired and permissions allow:
		// err = minioClient.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
		// if err != nil { ... handle error ... }
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

	// Set request parameters (empty for GET)
	reqParams := make(url.Values)
	// Example: Force download with filename
	// reqParams.Set("response-content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", objectKey))

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

	opts := minio.RemoveObjectOptions{
		// GovernanceBypass: true, // Set to true to bypass object retention locks if configured
	}

	err := s.client.RemoveObject(ctx, bucket, objectKey, opts)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to delete object from MinIO", "error", err, "bucket", bucket, "key", objectKey)
		// Check if the error indicates the object wasn't found - might not be a fatal error depending on use case
		// errResp := minio.ToErrorResponse(err)
		// if errResp.Code == "NoSuchKey" { return nil } // Example: treat not found as success
		return fmt.Errorf("failed to delete object %s/%s: %w", bucket, objectKey, err)
	}

	s.logger.InfoContext(ctx, "Deleted object from MinIO", "bucket", bucket, "key", objectKey)
	return nil
}

// GetPresignedPutURL generates a temporary URL for uploading an object.
func (s *MinioStorageService) GetPresignedPutURL(ctx context.Context, bucket, objectKey string, expiry time.Duration /*, opts port.PutObjectOptions? */) (string, error) {
	if bucket == "" {
		bucket = s.defaultBucket
	}
	if expiry <= 0 {
		expiry = s.defaultExpiry
	}

	// Removed unused putOpts variable
	// putOpts := minio.PutObjectOptions{}
	// if opts != nil && opts.ContentType != "" {
	//     putOpts.ContentType = opts.ContentType
	// }

	// Note: PresignedPutObject might require url.Values for certain headers like content-type constraints
	// Check minio-go SDK documentation for the exact way to enforce Content-Type if needed.
	// Example (conceptual, check SDK):
	policy := minio.NewPostPolicy()
	policy.SetBucket(bucket)
	policy.SetKey(objectKey)
	policy.SetExpires(time.Now().UTC().Add(expiry))
	// if opts != nil && opts.ContentType != "" {
	//    policy.SetContentType(opts.ContentType)
	// }
	// presignedURL, err := s.client.PresignedPostPolicy(ctx, policy) // For POST uploads
	// OR use PresignedPutObject directly, potentially setting headers via request parameters

	// Simpler version using PresignedPutObject without strict header enforcement in signature itself:
	presignedURL, err := s.client.PresignedPutObject(ctx, bucket, objectKey, expiry)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to generate presigned PUT URL", "error", err, "bucket", bucket, "key", objectKey)
		return "", fmt.Errorf("failed to get presigned PUT URL for %s/%s: %w", bucket, objectKey, err)
	}

	s.logger.DebugContext(ctx, "Generated presigned PUT URL", "bucket", bucket, "key", objectKey, "expiry", expiry)
	return presignedURL.String(), nil
}

// Compile-time check to ensure MinioStorageService satisfies the port.FileStorageService interface
var _ port.FileStorageService = (*MinioStorageService)(nil)
