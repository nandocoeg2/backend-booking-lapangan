package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/nando/booking_lapangan_backend/internal/config"
	"github.com/nando/booking_lapangan_backend/internal/database"
	"github.com/nando/booking_lapangan_backend/internal/models"
	"github.com/nando/booking_lapangan_backend/internal/routes"
	"github.com/nando/booking_lapangan_backend/internal/storage"
)

func main() {
	// 1. Load Config
	cfg := config.LoadConfig()

	// 2. Connect Database & Storage
	database.ConnectDB(cfg)
	storage.InitMinio(cfg)

	// 3. Auto Migrate (For initialization)
	// In production, use standard migration scripts
	err := database.DB.AutoMigrate(
		&models.User{},
		&models.Venue{},
		&models.Court{},
		&models.Schedule{},
		&models.Booking{},
	)
	if err != nil {
		log.Fatal("Migration failed: ", err)
	}
	log.Println("Database migrated successfully")

	// 4. Init Fiber
	app := fiber.New(fiber.Config{
		AppName: "CourtBooker API",
	})

	// 5. Middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())

	// 6. Routes
	routes.SetupRoutes(app)

	// 7. Start Server
	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
