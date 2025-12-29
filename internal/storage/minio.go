package storage

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nando/booking_lapangan_backend/internal/config"
)

var MinioClient *minio.Client

func InitMinio(cfg config.Config) {
	var err error
	MinioClient, err = minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		log.Fatalf("Failed to initialize MinIO client: %v", err)
	}

	// Ensure bucket exists
	ctx := context.Background()
	exists, err := MinioClient.BucketExists(ctx, cfg.MinioBucket)
	if err != nil {
		log.Fatalf("Failed to check bucket existence: %v", err)
	}
	if !exists {
		err = MinioClient.MakeBucket(ctx, cfg.MinioBucket, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatalf("Failed to create bucket: %v", err)
		}
		log.Printf("Successfully created bucket %s", cfg.MinioBucket)
	}

	// Set bucket policy to public (read-only)
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}
		]
	}`, cfg.MinioBucket)

	err = MinioClient.SetBucketPolicy(ctx, cfg.MinioBucket, policy)
	if err != nil {
		log.Printf("Warning: Failed to set bucket policy: %v", err)
	}
}

func UploadFile(file *multipart.FileHeader) (string, error) {
	ctx := context.Background()
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Generate unique filename with 'venues/' path prefix
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("venues/%s%s", uuid.New().String(), ext)
	contentType := file.Header.Get("Content-Type")

	// Upload to MinIO
	// Assuming config.LoadConfig() is accessible or we pass bucket name.
	// For simplicity, re-loading config or using global ref if needed.
	// Best practice: Pass config or use initialized global.
	// We'll use the one from Init for now, but better to store bucket name in a var.
	// Let's modify InitMinio to store bucket name globally for this package.

	// Quick fix: Load config again or assume standard
	cfg := config.LoadConfig()

	info, err := MinioClient.PutObject(ctx, cfg.MinioBucket, filename, src, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	log.Printf("Successfully uploaded %s of size %d\n", filename, info.Size)

	// Construct public URL
	// If running in tunnel, we might need a different base URL logic.
	// For now, return relative path or full internal URL.
	// Given the user wants to test on mobile, we should proabably return the tunnel URL + bucket + filename?
	// Or just the filename and let frontend construct it?
	// Standard S3 style: http://endpoint/bucket/filename

	// For localhost testing with tunnel:
	// If endpoint is localhost:9000, that's internal.
	// External might be handled by cloudflare if routing localhost:9000.
	// But usually cloudflare routes port 8080.
	// If user wants images to be accessible, they might need to tunnel 9000 too.
	// OR we proxy images through our backend API.

	// Let's assume for now we return the filename and let the client decide,
	// OR we return a fully qualified URL if we know the public gateway.

	// Since we didn't ask for a public gateway URL for MinIO,
	// and Cloudflare is tunneling `court-booker` -> `localhost:8080` (presumably),
	// we should PROXY the image requests via Go Fiber to avoid needing a second tunnel.

	return filename, nil
}
