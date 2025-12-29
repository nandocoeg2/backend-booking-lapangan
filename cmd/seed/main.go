package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/nando/booking_lapangan_backend/internal/config"
	"github.com/nando/booking_lapangan_backend/internal/database"
	"github.com/nando/booking_lapangan_backend/internal/models"
	"github.com/nando/booking_lapangan_backend/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 1. Load Config
	cfg := config.LoadConfig()
	log.Println("Config loaded")

	// 2. Initialize DB
	database.ConnectDB(cfg)
	log.Println("Database connected")

	// 3. Initialize MinIO
	storage.InitMinio(cfg)
	log.Println("MinIO initialized")

	// 4. AutoMigrate (Optional, usually done by main app, but good for safety)
	err := database.DB.AutoMigrate(&models.User{}, &models.Venue{}, &models.Court{}, &models.Schedule{}, &models.Booking{})
	if err != nil {
		log.Fatalf("Failed to auto migrate: %v", err)
	}

	// 5. Seed Data
	seedUsers()
	uploadedImages := seedImages(cfg)
	seedVenues(uploadedImages)

	log.Println("Seeding completed successfully")
}

func seedUsers() {
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	password := string(passwordHash)

	users := []models.User{
		{
			Name:         "Admin User",
			Email:        "admin@example.com",
			PasswordHash: password,
			Role:         models.RoleAdmin,
			PhoneNumber:  "081234567890",
		},
		{
			Name:         "Venue Owner 1",
			Email:        "owner1@example.com",
			PasswordHash: password,
			Role:         models.RoleVenueOwner,
			PhoneNumber:  "081234567891",
		},
		{
			Name:         "Regular User",
			Email:        "user@example.com",
			PasswordHash: password,
			Role:         models.RoleUser,
			PhoneNumber:  "081234567892",
		},
	}

	for _, u := range users {
		var existingUser models.User
		if err := database.DB.Where("email = ?", u.Email).First(&existingUser).Error; err == nil {
			log.Printf("User %s already exists, skipping...", u.Email)
			continue
		}
		if err := database.DB.Create(&u).Error; err != nil {
			log.Printf("Failed to create user %s: %v", u.Email, err)
		} else {
			log.Printf("Created user %s", u.Email)
		}
	}
}

func seedImages(cfg config.Config) map[string]string {
	// Map of original filename -> MinIO object name
	uploadedImages := make(map[string]string)

	// Path to assets using relative path from backend root (assuming running from backend dir)
	assetDir := "../assets/images"

	files, err := os.ReadDir(assetDir)
	if err != nil {
		log.Printf("Failed to read asset directory (check if running from backend dir): %v", err)
		// Fallback to absolute path or ignore if fails
		return uploadedImages
	}

	ctx := context.Background()

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		filename := f.Name()
		ext := filepath.Ext(filename)
		if !strings.EqualFold(ext, ".jpg") && !strings.EqualFold(ext, ".png") && !strings.EqualFold(ext, ".jpeg") {
			continue
		}

		filePath := filepath.Join(assetDir, filename)
		file, err := os.Open(filePath)
		if err != nil {
			log.Printf("Failed to open file %s: %v", filename, err)
			continue
		}

		fileInfo, _ := file.Stat()
		size := fileInfo.Size()

		// Use original filename for simplicity in seed to easier mapping, or prefix
		objectName := fmt.Sprintf("venues/seed_%s", filename)
		contentType := "image/jpeg"
		if strings.HasSuffix(strings.ToLower(filename), ".png") {
			contentType = "image/png"
		}

		_, err = storage.MinioClient.PutObject(ctx, cfg.MinioBucket, objectName, file, size, minio.PutObjectOptions{
			ContentType: contentType,
		})
		if err != nil {
			log.Printf("Failed to upload %s to MinIO: %v", filename, err)
		} else {
			log.Printf("Uploaded %s to MinIO as %s", filename, objectName)
			uploadedImages[filename] = objectName
		}
		file.Close()
	}

	return uploadedImages
}

func seedVenues(images map[string]string) {
	// Find Owner
	var owner models.User
	if err := database.DB.Where("email = ?", "owner1@example.com").First(&owner).Error; err != nil {
		log.Printf("Owner not found, skipping venue creation")
		return
	}

	// Helper to get image URL (just the filename as stored in DB)
	getImage := func(name string) string {
		if val, ok := images[name]; ok {
			return val
		}
		return ""
	}

	venues := []models.Venue{
		{
			Name:          "Gor Bulutangkis Sejahtera",
			Description:   "Lapangan bulutangkis terbaik di Jakarta Selatan dengan fasilitas lengkap.",
			Address:       "Jl. Fatmawati No. 10",
			Location:      "Jakarta Selatan",
			Latitude:      -6.261493,
			Longitude:     106.810600,
			Rating:        4.8,
			StartingPrice: 50000,
			Category:      "Badminton",
			Facilities:    []string{"Parking", "Toilet", "Wifi", "Canteen"},
			ImageURLs:     filterEmpty([]string{getImage("badminton_court_1.jpg"), getImage("badminton_court_2.jpg")}),
			OwnerID:       owner.ID,
			Courts: []models.Court{
				{Name: "Court A", Description: "Lapangan Karpet", IsActive: true},
				{Name: "Court B", Description: "Lapangan Kayu", IsActive: true},
				{Name: "Court C", Description: "Lapangan Semen", IsActive: true},
			},
		},
		{
			Name:          "Futsal Center 99",
			Description:   "Lapangan futsal rumput sintetis standar internasional.",
			Address:       "Jl. Kemang Raya No. 99",
			Location:      "Jakarta Selatan",
			Latitude:      -6.280000,
			Longitude:     106.820000,
			Rating:        4.5,
			StartingPrice: 120000,
			Category:      "Futsal",
			Facilities:    []string{"Parking", "Toilet", "Shower", "Locker"},
			ImageURLs:     filterEmpty([]string{getImage("futsal_court_1.jpg"), getImage("futsal_court_2.jpg")}),
			OwnerID:       owner.ID,
			Courts: []models.Court{
				{Name: "Lapangan 1", Description: "Rumput Sintetis", IsActive: true},
				{Name: "Lapangan 2", Description: "Interlock", IsActive: true},
			},
		},
		{
			Name:          "Tennis Club Elite",
			Description:   "Klub tenis eksklusif dengan pelatih profesional.",
			Address:       "Jl. Sudirman Kav. 50",
			Location:      "Jakarta Pusat",
			Latitude:      -6.200000,
			Longitude:     106.800000,
			Rating:        4.9,
			StartingPrice: 150000,
			Category:      "Tennis",
			Facilities:    []string{"Parking", "Cafe", "Pro Shop", "Coach"},
			ImageURLs:     filterEmpty([]string{getImage("tennis_court_1.jpg"), getImage("tennis_court_2.jpg")}),
			OwnerID:       owner.ID,
			Courts: []models.Court{
				{Name: "Center Court", Description: "Hard Court", IsActive: true},
				{Name: "Court 1", Description: "Clay Court", IsActive: true},
			},
		},
		{
			Name:          "Basket Hall 23",
			Description:   "Lapangan basket indoor full AC.",
			Address:       "Jl. Asia Afrika",
			Location:      "Jakarta Pusat",
			Latitude:      -6.220000,
			Longitude:     106.800000,
			Rating:        4.7,
			StartingPrice: 200000,
			Category:      "Basket",
			Facilities:    []string{"Parking", "Toilet", "Changing Room", "Scoreboard"},
			ImageURLs:     filterEmpty([]string{getImage("basket_court_1.jpg"), getImage("basket_court_2.jpg")}),
			OwnerID:       owner.ID,
			Courts: []models.Court{
				{Name: "Main Court", Description: "Parquet Flooring", IsActive: true},
			},
		},
	}

	for _, v := range venues {
		var existingVenue models.Venue
		if err := database.DB.Where("name = ?", v.Name).First(&existingVenue).Error; err == nil {
			log.Printf("Venue %s already exists, skipping...", v.Name)
			continue
		}

		if err := database.DB.Create(&v).Error; err != nil {
			log.Printf("Failed to create venue %s: %v", v.Name, err)
		} else {
			log.Printf("Created venue %s with %d courts", v.Name, len(v.Courts))

			// Add schedules for each court
			for _, court := range v.Courts {
				seedSchedules(court.ID, v.StartingPrice)
			}
		}
	}
}

func seedSchedules(courtID uint, basePrice float64) {
	// Create schedules for Monday (1) to Sunday (0/7)
	// Simple schedule: 08:00 - 22:00
	days := []int{0, 1, 2, 3, 4, 5, 6}
	startHour := 8
	endHour := 22

	var schedules []models.Schedule
	for _, day := range days {
		for h := startHour; h < endHour; h++ {
			startTime := fmt.Sprintf("%02d:00", h)
			endTime := fmt.Sprintf("%02d:00", h+1)

			price := basePrice
			if h >= 18 { // Evening price
				price += 20000
			}

			schedules = append(schedules, models.Schedule{
				CourtID:   courtID,
				DayOfWeek: day,
				StartTime: startTime,
				EndTime:   endTime,
				Price:     price,
				IsPromo:   false,
			})
		}
	}

	database.DB.Create(&schedules)
}

func filterEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}
