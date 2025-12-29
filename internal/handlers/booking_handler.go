package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nando/booking_lapangan_backend/internal/database"
	"github.com/nando/booking_lapangan_backend/internal/models"
)

// CreateBookingRequest
type CreateBookingRequest struct {
	CourtID     uint   `json:"court_id"`
	BookingDate string `json:"booking_date"` // YYYY-MM-DD
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Price       float64
}

func CheckAvailability(c *fiber.Ctx) error {
	courtID := c.Params("court_id")
	date := c.Query("date") // YYYY-MM-DD

	if date == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Date required"})
	}

	var bookings []models.Booking
	database.DB.Where("court_id = ? AND booking_date = ?", courtID, date).Find(&bookings)

	// In a real app, you would generate all slots (08:00 - 22:00) and mark them based on bookings.
	// For now, returning the raw bookings is enough for the frontend to map.

	return c.JSON(fiber.Map{"bookings": bookings})
}

func CreateBooking(c *fiber.Ctx) error {
	var req CreateBookingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userID := c.Locals("user_id").(float64)

	// Basic validation (skipped for brevity)

	// Parse date
	parsedDate, _ := time.Parse("2006-01-02", req.BookingDate)

	// Create Booking
	booking := models.Booking{
		ID:            fmt.Sprintf("BSK-%d-%d", int(userID), time.Now().Unix()), // Simple ID generation
		UserID:        uint(userID),
		CourtID:       req.CourtID,
		BookingDate:   parsedDate,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		TotalPrice:    req.Price, // Should be calculated on backend in prod
		Status:        models.BookingStatusPending,
		PaymentMethod: "XENDIT_VA",
	}

	// ---------------------------------------------------------
	// XENDIT INTEGRATION (Simulated)
	// ---------------------------------------------------------
	// In a real implementation:
	// xenditReq := xendit.InvoiceRequest{ ... }
	// resp, _ := xenditClient.CreateInvoice(xenditReq)
	// booking.PaymentURL = resp.InvoiceURL
	// ---------------------------------------------------------

	// For MVP/Demo:
	booking.PaymentURL = "https://checkout-staging.xendit.co/web/65123abc" // Mock URL

	result := database.DB.Create(&booking)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create booking"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Booking created. Please convert to payment.",
		"booking": booking,
		"payment_action": fiber.Map{
			"type": "redirect",
			"url":  booking.PaymentURL,
		},
	})
}
