package models

import (
	"time"

	"gorm.io/gorm"
)

type Venue struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Address       string         `json:"address"`
	Location      string         `json:"location"` // City/District
	Latitude      float64        `json:"latitude"`
	Longitude     float64        `json:"longitude"`
	Rating        float64        `json:"rating"`
	StartingPrice float64        `json:"starting_price"`
	Facilities    []string       `gorm:"serializer:json" json:"facilities"` // Requires GORM JSON serializer usually
	ImageURLs     []string       `gorm:"serializer:json" json:"image_urls"`
	Category      string         `json:"category"` // Badminton, Futsal
	OwnerID       uint           `json:"owner_id"` // FK to User
	Courts        []Court        `json:"courts,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type Court struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	VenueID     uint           `json:"venue_id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type Schedule struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	CourtID   uint    `json:"court_id"`
	DayOfWeek int     `json:"day_of_week"` // 0=Sunday, 1=Monday...
	StartTime string  `json:"start_time"`  // "08:00"
	EndTime   string  `json:"end_time"`    // "09:00"
	Price     float64 `json:"price"`
	IsPromo   bool    `json:"is_promo"`
}
