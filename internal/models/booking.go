package models

import (
	"time"

	"gorm.io/gorm"
)

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "PENDING"
	BookingStatusPaid      BookingStatus = "PAID"
	BookingStatusCancelled BookingStatus = "CANCELLED"
	BookingStatusCompleted BookingStatus = "COMPLETED"
)

type Booking struct {
	ID            string         `gorm:"primaryKey" json:"id"` // Custom ID e.g. BSK-88291
	UserID        uint           `json:"user_id"`
	CourtID       uint           `json:"court_id"`
	Court         Court          `json:"court,omitempty"`
	BookingDate   time.Time      `json:"booking_date"` // YYYY-MM-DD
	StartTime     string         `json:"start_time"`
	EndTime       string         `json:"end_time"`
	TotalPrice    float64        `json:"total_price"`
	Status        BookingStatus  `json:"status"`
	PaymentMethod string         `json:"payment_method"`
	PaymentURL    string         `json:"payment_url"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
