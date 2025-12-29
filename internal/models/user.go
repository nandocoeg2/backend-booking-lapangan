package models

import (
	"time"

	"gorm.io/gorm"
)

type Role string

const (
	RoleUser       Role = "USER"
	RoleAdmin      Role = "ADMIN"
	RoleVenueOwner Role = "VENUE_OWNER"
)

type User struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	Name              string         `json:"name"`
	Email             string         `gorm:"uniqueIndex" json:"email"`
	PasswordHash      string         `json:"-"`
	PhoneNumber       string         `json:"phone_number"`
	ProfilePictureURL string         `json:"profile_picture_url"`
	Role              Role           `gorm:"default:'USER'" json:"role"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}
