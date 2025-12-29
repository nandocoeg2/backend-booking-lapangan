package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/minio/minio-go/v7"
	"github.com/nandocoeg2/backend-booking-lapangan/internal/config"
	"github.com/nandocoeg2/backend-booking-lapangan/internal/storage"
)

// GetImage proxies the image from MinIO to the client.
// Route: GET /api/v1/media/:filename
func GetImage(c *fiber.Ctx) error {
	filename := c.Params("filename")
	if filename == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Filename is required",
		})
	}

	cfg := config.LoadConfig()
	ctx := context.Background()

	// Get object from MinIO
	object, err := storage.MinioClient.GetObject(ctx, cfg.MinioBucket, filename, minio.GetObjectOptions{})
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Image not found",
		})
	}
	defer object.Close()

	// Check if object exists and get info
	info, err := object.Stat()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Image not found",
		})
	}

	// Set headers
	c.Set("Content-Type", info.ContentType)
	c.Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
	c.Set("Content-Length", fmt.Sprintf("%d", info.Size))
	c.Set("Last-Modified", info.LastModified.Format(time.RFC1123))

	// Stream object to response
	// Fiber's c.Response().BodyWriter() isn't directly compatible with io.Copy easily in all versions,
	// but c.SendStream is available in newer Fiber versions or we can use standard readAll (less efficient).
	// Best approach for Fiber v2 is usually c.SendStream or just writing to the body if small.
	// Let's use io.Copy to c.Response().BodyWriter() if possible, or read to buffer.

	// Efficient streaming in Fiber:
	return c.SendStream(object, int(info.Size))
}
