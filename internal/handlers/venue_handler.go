package handlers

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/nandocoeg2/backend-booking-lapangan/internal/database"
	"github.com/nandocoeg2/backend-booking-lapangan/internal/models"
	"github.com/nandocoeg2/backend-booking-lapangan/internal/storage"
)

// CreateVenueRequest
type CreateVenueRequest struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Address       string   `json:"address"`
	Location      string   `json:"location"`
	Latitude      float64  `json:"latitude"`
	Longitude     float64  `json:"longitude"`
	Rating        float64  `json:"rating"` // Initial fake rating
	StartingPrice float64  `json:"starting_price"`
	Facilities    []string `json:"facilities"`
	ImageURLs     []string `json:"image_urls"`
	Category      string   `json:"category"`
}

func CreateVenue(c *fiber.Ctx) error {
	// Parse Multipart Form
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request (expecting multipart/form-data)"})
	}

	// Helper to get form value
	getValue := func(key string) string {
		if v, ok := form.Value[key]; ok && len(v) > 0 {
			return v[0]
		}
		return ""
	}

	// Parse fields manually from form strings
	// In a real app, use a better binder or validation
	name := getValue("name")
	description := getValue("description")
	address := getValue("address")
	location := getValue("location")
	category := getValue("category")

	lat, _ := strconv.ParseFloat(getValue("latitude"), 64)
	long, _ := strconv.ParseFloat(getValue("longitude"), 64)
	price, _ := strconv.ParseFloat(getValue("starting_price"), 64)

	// Facilities (comma separated or multiple values)
	facilities := form.Value["facilities"]
	// If passed as facilities[]=A&facilities[]=B, multipart parser handles it.
	// If json string, need unmarshal. Assuming simple multiple keys for now.

	userID := c.Locals("user_id").(float64)

	// Handle Image Uploads
	var imageURLs []string
	files := form.File["images"] // Expecting images sent as 'images' key

	for _, file := range files {
		filename, err := storage.UploadFile(file)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to upload image: " + err.Error()})
		}
		// Construct Proxy URL: /api/v1/media/:filename
		// Since we don't know the full host here easily without config, we'll store the relative proxy path
		// OR store full URL if we knew the domain.
		// Best practice: Store relative path or key, and construct URL on Get.
		// For simplicity/compatibility with active frontend which expects full URLs or assets:
		// We'll store the API Proxy URL.
		// Note: Frontend needs to handle this.
		// Let's assume we store: "http(s)://backend-host/api/v1/media/filename"
		// But "backend-host" varies (localhost vs tunnel).
		// Storing just the filename or "/api/v1/media/filename" is safer.
		// Let's store "/api/v1/media/filename" and let frontend prepend host if needed,
		// OR frontend treats it as relative?

		// Wait, the frontend code we wrote handles `displayImage.startsWith('http')`.
		// If we store "/api/v1/media/...", it won't start with http.
		// We should probably rely on the backend to return full URLs in the API response
		// (by prepending the request Host).
		// But simply storing "http://localhost:8080/api/v1/media/..." breaks in tunnel.
		// So we will store just the PROXY PATH "/api/v1/media/filename".
		// And ensure frontend (or API response wrapper) handles it.
		// Actually, `storage.UploadFile` returns just filename.

		imageURLs = append(imageURLs, fmt.Sprintf("/api/v1/media/%s", filename))
	}

	// If no images uploaded but image_urls provided (e.g. from existing external links), handle that?
	// Skipping for now.

	venue := models.Venue{
		Name:          name,
		Description:   description,
		Address:       address,
		Location:      location,
		Latitude:      lat,
		Longitude:     long,
		Rating:        4.5, // Default new
		StartingPrice: price,
		Facilities:    facilities,
		ImageURLs:     imageURLs,
		Category:      category,
		OwnerID:       uint(userID),
	}

	result := database.DB.Create(&venue)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create venue"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Venue created successfully",
		"venue":   venue,
	})
}

func GetVenues(c *fiber.Ctx) error {
	var venues []models.Venue
	query := database.DB

	// Filters
	if location := c.Query("location"); location != "" {
		query = query.Where("location ILIKE ?", "%"+location+"%")
	}
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}
	if search := c.Query("search"); search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	query.Find(&venues)
	return c.JSON(fiber.Map{"data": venues})
}

func GetVenue(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var venue models.Venue
	result := database.DB.Preload("Courts").First(&venue, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Venue not found"})
	}

	return c.JSON(fiber.Map{"data": venue})
}
