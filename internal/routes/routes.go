package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nando/booking_lapangan_backend/internal/handlers"
	"github.com/nando/booking_lapangan_backend/internal/middleware"
	"github.com/nando/booking_lapangan_backend/internal/models"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	// Auth
	auth := api.Group("/auth")
	auth.Post("/register", handlers.Register)
	auth.Post("/login", handlers.Login)

	// User Routes
	user := api.Group("/user", middleware.Protected)
	user.Get("/me", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"user_id": c.Locals("user_id"),
			"role":    c.Locals("role"),
			"msg":     "Hello User",
		})
	})
	user.Get("/courts/:court_id/availability", handlers.CheckAvailability)
	user.Post("/bookings", handlers.CreateBooking)

	// Public Venue Routes
	api.Get("/venues", handlers.GetVenues)
	api.Get("/venues/:id", handlers.GetVenue)
	api.Get("/media/:filename", handlers.GetImage) // Image Proxy

	// Owner Routes
	owner := api.Group("/owner", middleware.Protected, middleware.RoleMiddleware(models.RoleVenueOwner))
	owner.Get("/dashboard", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"msg": "Owner Dashboard"})
	})
	owner.Post("/venues", handlers.CreateVenue)

	// Admin Routes
	admin := api.Group("/admin", middleware.Protected, middleware.RoleMiddleware(models.RoleAdmin))
	admin.Get("/dashboard", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"msg": "Admin Dashboard"})
	})
	admin.Post("/venues", handlers.CreateVenue)
}
