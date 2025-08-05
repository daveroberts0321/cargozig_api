package handlers

import (
	"cargozig_api/config"
	"cargozig_api/middleware"
	"cargozig_api/models"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// SetupAdminRoutes sets up the admin page routes
// /toc/login, /toc/dashboard, /toc/registernewuser, /toc/setup
func SetupAdminRoutes(router fiber.Router) {
	// Admin page routes (HTML pages)
	router.Post("/admin/setup", AdminSetup) // Admin setup API endpoint (no auth required) - must come first
	router.Get("/setup", AdminSetupPage)    // Initial admin setup (no auth required)
	router.Get("/login", AdminLoginPage)
	router.Get("/dashboard", middleware.AdminAuthMiddleware(), AdminDashboardPage)
	router.Get("/registernewuser", middleware.AdminAuthMiddleware(), AdminRegisterNewUserPage)
	router.Get("/", LandingPage) // Root route should come last

	// Debug route to test if routing is working
	router.Post("/debug", func(c *fiber.Ctx) error {
		fmt.Println("=== Debug route hit ===")
		return c.JSON(fiber.Map{"message": "Debug route working"})
	})
}

// LandingPage renders the home page
func LandingPage(c *fiber.Ctx) error {
	return c.Render("index", fiber.Map{
		"Title": "CargoZig - Smart Freight Brokerage Platform",
	}, "layouts/main")
}

// AdminLoginPage serves the admin login HTML page
func AdminLoginPage(c *fiber.Ctx) error {
	fmt.Println("Serving admin login page")
	return c.Render("admin/login", fiber.Map{
		"Title": "Admin Login",
	}, "")
}

// AdminDashboardPage serves the admin dashboard HTML page
func AdminDashboardPage(c *fiber.Ctx) error {
	fmt.Println("Serving admin dashboard page")
	return c.Render("admin/dashboard", fiber.Map{
		"Title":      "Admin Dashboard",
		"PageTitle":  "Admin Dashboard",
		"ActivePage": "dashboard",
	}, "layouts/admin")
}

// AdminRegisterNewUserPage serves the admin register new user HTML page
func AdminRegisterNewUserPage(c *fiber.Ctx) error {
	fmt.Println("Serving admin register new user page")
	return c.Render("admin/register-new-user", fiber.Map{
		"Title":      "Register New User",
		"PageTitle":  "Register New User",
		"ActivePage": "register-user",
	}, "layouts/admin")
}

// AdminSetupPage serves the initial admin setup page (no auth required)
func AdminSetupPage(c *fiber.Ctx) error {
	fmt.Println("Serving admin setup page")
	return c.Render("admin/setup", fiber.Map{
		"Title": "Admin Setup",
	}, "")
}

// AdminSetup creates the first admin user (no authentication required)
func AdminSetup(c *fiber.Ctx) error {
	fmt.Println("=== AdminSetup called ===")

	// Get database instance
	db := config.GetDB()

	// Check if any admin users exist
	var adminCount int64
	countErr := db.Model(&models.User{}).Where("roles @> ?", []string{"admin"}).Count(&adminCount)
	if countErr != nil {
		fmt.Printf("Error checking admin count: %v\n", countErr)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error checking admin count"})
	}

	fmt.Printf("Found %d existing admin users\n", adminCount)

	// Only allow setup if no admin users exist
	if adminCount > 0 {
		fmt.Println("Admin users already exist. Setup not allowed.")
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Admin setup not allowed. Admin users already exist."})
	}

	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		fmt.Println("Error parsing admin setup request:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate input
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing required fields"})
	}

	// Trim spaces
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	// Check if user already exists
	var existingUser models.User
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "User already exists"})
	}

	// Hash password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error hashing admin password:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	// Create new admin user
	newAdmin := models.User{
		Username:    req.Username,
		Email:       req.Email,
		Password:    string(hashedPassword),
		Roles:       []models.Role{models.RoleAdmin}, // Admin role only
		Permissions: []models.Permission{},           // No custom permissions initially
		Active:      true,
	}

	// Save user to the database
	if err := db.Create(&newAdmin).Error; err != nil {
		fmt.Println("Error creating admin user:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create admin user"})
	}

	fmt.Printf("First admin user created successfully: %s (ID: %s)\n", newAdmin.Username, newAdmin.ID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "First admin user created successfully",
		"user_id": newAdmin.ID,
		"user": fiber.Map{
			"id":       newAdmin.ID,
			"username": newAdmin.Username,
			"email":    newAdmin.Email,
			"roles":    newAdmin.Roles,
		},
	})
}
