package handlers

import (
	"cargozig_api/config"
	"cargozig_api/models"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// SetupPublicRoutes sets up the public page routes
func SetupPublicRoutes(router fiber.Router) {
	// Public pages (no auth required)
	router.Get("/", HomePage)
	router.Get("/about", AboutPage)
	router.Get("/features", FeaturesPage)
	router.Get("/contact", ContactPage)
	router.Get("/login", PublicLoginPage)

	// Public API endpoints
	router.Post("/api/contact", ContactFormHandler)
}

// HomePage renders the public home page
func HomePage(c *fiber.Ctx) error {
	return c.Render("pages/home", fiber.Map{
		"Title":       "CargoZig - Smart Freight Brokerage Platform",
		"Description": "Revolutionary freight brokerage platform with intelligent automation, real-time tracking, and seamless carrier-shipper connections.",
	}, "layouts/public")
}

// AboutPage renders the about page
func AboutPage(c *fiber.Ctx) error {
	return c.Render("pages/about", fiber.Map{
		"Title":       "About CargoZig - Our Mission & Story",
		"Description": "Learn about CargoZig's mission to revolutionize freight brokerage through intelligent technology and seamless connectivity.",
	}, "layouts/public")
}

// FeaturesPage renders the features page
func FeaturesPage(c *fiber.Ctx) error {
	return c.Render("pages/features", fiber.Map{
		"Title":       "CargoZig Features - Advanced Freight Brokerage Tools",
		"Description": "Discover powerful features designed for modern freight brokerage operations including real-time tracking, AI matching, and analytics.",
	}, "layouts/public")
}

// ContactPage renders the contact page
func ContactPage(c *fiber.Ctx) error {
	return c.Render("pages/contact", fiber.Map{
		"Title":       "Contact CargoZig - Get in Touch",
		"Description": "Ready to transform your freight operations? Contact our team to learn how CargoZig can help streamline your brokerage business.",
	}, "layouts/public")
}

// PublicLoginPage renders the public login page
func PublicLoginPage(c *fiber.Ctx) error {
	return c.Render("pages/login", fiber.Map{
		"Title": "Login - CargoZig",
	}, "layouts/public")
}

// ContactFormHandler handles contact form submissions
func ContactFormHandler(c *fiber.Ctx) error {
	var req struct {
		FirstName  string `json:"firstName"`
		LastName   string `json:"lastName"`
		Email      string `json:"email"`
		Company    string `json:"company"`
		Phone      string `json:"phone"`
		Subject    string `json:"subject"`
		Message    string `json:"message"`
		Newsletter bool   `json:"newsletter"`
	}

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		fmt.Println("Error parsing contact form:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate required fields
	if req.FirstName == "" || req.LastName == "" || req.Email == "" || req.Subject == "" || req.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing required fields"})
	}

	// Get database instance
	db := config.GetDB()

	// Create contact record
	contact := models.Contact{
		Name:    req.FirstName + " " + req.LastName,
		Email:   req.Email,
		Phone:   req.Phone,
		Company: req.Company,
		Subject: req.Subject,
		Message: req.Message,
		Status:  "new",
	}

	// Save to database
	if err := db.Create(&contact).Error; err != nil {
		fmt.Println("Error saving contact:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save contact"})
	}

	// Handle newsletter subscription if requested
	if req.Newsletter {
		mailingList := models.MailingList{
			Email:  req.Email,
			Name:   req.FirstName + " " + req.LastName,
			Active: true,
			Source: "contact_form",
		}

		// Check if email already exists
		var existing models.MailingList
		result := db.Where("email = ?", req.Email).First(&existing)
		if result.Error != nil {
			// Email doesn't exist, create new entry
			if err := db.Create(&mailingList).Error; err != nil {
				fmt.Println("Error saving to mailing list:", err)
				// Don't fail the whole request for newsletter signup
			}
		}
	}

	fmt.Printf("Contact form submitted by %s (%s)\n", contact.Name, contact.Email)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Thank you for your message. We'll get back to you soon!",
	})
}
