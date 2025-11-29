package handlers

import (
	"cargozig_api/config"
	"cargozig_api/middleware"
	"cargozig_api/models"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// SetupSuperAdminRoutes sets up the super admin routes
// Super admins can manage all companies, users, and system settings
func SetupSuperAdminRoutes(router fiber.Router) {
	// All super admin routes require authentication and system_admin permission
	router.Use(middleware.AuthenticateUser())
	router.Use(middleware.RequirePermission(models.SystemAdmin))

	// Dashboard
	router.Get("/dashboard", SuperAdminDashboard)

	// Companies Management
	router.Get("/companies", SuperAdminCompanies)
	router.Get("/companies/new", SuperAdminNewCompany)
	router.Get("/companies/:id", SuperAdminViewCompany)
	router.Get("/companies/:id/edit", SuperAdminEditCompany)

	// Users Management
	router.Get("/users", SuperAdminUsers)
	router.Get("/users/new", SuperAdminNewUser)
	router.Get("/users/:id", SuperAdminViewUser)
	router.Get("/users/:id/edit", SuperAdminEditUser)

	// Shippers Management
	router.Get("/shippers", SuperAdminShippers)
	router.Get("/shippers/:id", SuperAdminViewShipper)
	router.Get("/shippers/:id/edit", SuperAdminEditShipper)

	// Carriers Management
	router.Get("/carriers", SuperAdminCarriers)
	router.Get("/carriers/:id", SuperAdminViewCarrier)
	router.Get("/carriers/:id/edit", SuperAdminEditCarrier)

	// Brokers Management
	router.Get("/brokers", SuperAdminBrokers)
	router.Get("/brokers/:id", SuperAdminViewBroker)
	router.Get("/brokers/:id/edit", SuperAdminEditBroker)

	// System Settings
	router.Get("/settings", SuperAdminSettings)
	router.Get("/analytics", SuperAdminAnalytics)

	// API endpoints for CRUD operations
	router.Delete("/api/shippers/:id", SuperAdminDeleteShipper)
	router.Delete("/api/carriers/:id", SuperAdminDeleteCarrier)
	router.Delete("/api/brokers/:id", SuperAdminDeleteBroker)
	router.Delete("/api/companies/:id", SuperAdminDeleteCompany)
	router.Delete("/api/users/:id", SuperAdminDeleteUser)
}

// SuperAdminDashboard renders the super admin dashboard
func SuperAdminDashboard(c *fiber.Ctx) error {
	fmt.Println("SuperAdminDashboard called")

	db := config.GetDB()

	// Get statistics
	var totalCompanies, totalUsers, totalShippers, totalCarriers int64
	db.Model(&models.Company{}).Count(&totalCompanies)
	db.Model(&models.User{}).Count(&totalUsers)
	db.Model(&models.User{}).Where("? = ANY(roles)", models.RoleShipper).Count(&totalShippers)
	db.Model(&models.User{}).Where("? = ANY(roles)", models.RoleCarrier).Count(&totalCarriers)

	return c.Render("superadmin/dashboard", fiber.Map{
		"Title":      "Super Admin Dashboard",
		"ActivePage": "dashboard",
		"Username":   c.Locals("username"),
		"Stats": fiber.Map{
			"TotalCompanies": totalCompanies,
			"TotalUsers":     totalUsers,
			"TotalShippers":  totalShippers,
			"TotalCarriers":  totalCarriers,
		},
		"RecentActivity": []fiber.Map{}, // TODO: Implement activity tracking
	}, "layouts/superadmin")
}

// SuperAdminShippers renders the shippers management page
func SuperAdminShippers(c *fiber.Ctx) error {
	fmt.Println("SuperAdminShippers called")

	db := config.GetDB()

	// Get all shippers (users with shipper role)
	var shippers []models.User
	db.Preload("Company").Where("? = ANY(roles)", models.RoleShipper).Find(&shippers)

	// Get all companies for filter
	var companies []models.Company
	db.Find(&companies)

	return c.Render("superadmin/shippers", fiber.Map{
		"Title":      "Shippers Management",
		"ActivePage": "shippers",
		"Username":   c.Locals("username"),
		"Shippers":   shippers,
		"Companies":  companies,
		"Pagination": fiber.Map{
			"Start":       1,
			"End":         len(shippers),
			"Total":       len(shippers),
			"CurrentPage": 1,
			"TotalPages":  1,
		},
	}, "layouts/superadmin")
}

// SuperAdminCarriers renders the carriers management page
func SuperAdminCarriers(c *fiber.Ctx) error {
	fmt.Println("SuperAdminCarriers called")

	db := config.GetDB()

	// Get all carriers (users with carrier role)
	var carriers []models.User
	db.Preload("Company").Where("? = ANY(roles)", models.RoleCarrier).Find(&carriers)

	// Get all companies for filter
	var companies []models.Company
	db.Find(&companies)

	return c.Render("superadmin/carriers", fiber.Map{
		"Title":      "Carriers Management",
		"ActivePage": "carriers",
		"Username":   c.Locals("username"),
		"Carriers":   carriers,
		"Companies":  companies,
		"Pagination": fiber.Map{
			"Start":       1,
			"End":         len(carriers),
			"Total":       len(carriers),
			"CurrentPage": 1,
			"TotalPages":  1,
		},
	}, "layouts/superadmin")
}

// SuperAdminBrokers renders the brokers management page
func SuperAdminBrokers(c *fiber.Ctx) error {
	fmt.Println("SuperAdminBrokers called")

	db := config.GetDB()

	// Get all brokers/admins (users with admin role)
	var brokers []models.User
	db.Preload("Company").Where("? = ANY(roles)", models.RoleAdmin).Find(&brokers)

	// Get all companies for filter
	var companies []models.Company
	db.Find(&companies)

	return c.Render("superadmin/brokers", fiber.Map{
		"Title":      "Brokers Management",
		"ActivePage": "brokers",
		"Username":   c.Locals("username"),
		"Brokers":    brokers,
		"Companies":  companies,
		"Pagination": fiber.Map{
			"Start":       1,
			"End":         len(brokers),
			"Total":       len(brokers),
			"CurrentPage": 1,
			"TotalPages":  1,
		},
	}, "layouts/superadmin")
}

// SuperAdminCompanies renders the companies management page
func SuperAdminCompanies(c *fiber.Ctx) error {
	fmt.Println("SuperAdminCompanies called")

	db := config.GetDB()

	// Get all companies
	var companies []models.Company
	db.Find(&companies)

	return c.Render("superadmin/companies", fiber.Map{
		"Title":      "Companies Management",
		"ActivePage": "companies",
		"Username":   c.Locals("username"),
		"Companies":  companies,
	}, "layouts/superadmin")
}

// SuperAdminUsers renders all users management page
func SuperAdminUsers(c *fiber.Ctx) error {
	fmt.Println("SuperAdminUsers called")

	db := config.GetDB()

	// Get all users
	var users []models.User
	db.Preload("Company").Find(&users)

	return c.Render("superadmin/users", fiber.Map{
		"Title":      "Users Management",
		"ActivePage": "users",
		"Username":   c.Locals("username"),
		"Users":      users,
	}, "layouts/superadmin")
}

// SuperAdminSettings renders the system settings page
func SuperAdminSettings(c *fiber.Ctx) error {
	return c.Render("superadmin/settings", fiber.Map{
		"Title":      "System Settings",
		"ActivePage": "settings",
		"Username":   c.Locals("username"),
	}, "layouts/superadmin")
}

// SuperAdminAnalytics renders the analytics page
func SuperAdminAnalytics(c *fiber.Ctx) error {
	return c.Render("superadmin/analytics", fiber.Map{
		"Title":      "Platform Analytics",
		"ActivePage": "analytics",
		"Username":   c.Locals("username"),
	}, "layouts/superadmin")
}

// Placeholder view/edit/new handlers (implement as needed)
func SuperAdminViewCompany(c *fiber.Ctx) error   { return c.SendString("View Company - TODO") }
func SuperAdminEditCompany(c *fiber.Ctx) error   { return c.SendString("Edit Company - TODO") }
func SuperAdminNewCompany(c *fiber.Ctx) error    { return c.SendString("New Company - TODO") }
func SuperAdminViewUser(c *fiber.Ctx) error      { return c.SendString("View User - TODO") }
func SuperAdminEditUser(c *fiber.Ctx) error      { return c.SendString("Edit User - TODO") }
func SuperAdminNewUser(c *fiber.Ctx) error       { return c.SendString("New User - TODO") }
func SuperAdminViewShipper(c *fiber.Ctx) error   { return c.SendString("View Shipper - TODO") }
func SuperAdminEditShipper(c *fiber.Ctx) error   { return c.SendString("Edit Shipper - TODO") }
func SuperAdminViewCarrier(c *fiber.Ctx) error   { return c.SendString("View Carrier - TODO") }
func SuperAdminEditCarrier(c *fiber.Ctx) error   { return c.SendString("Edit Carrier - TODO") }
func SuperAdminViewBroker(c *fiber.Ctx) error    { return c.SendString("View Broker - TODO") }
func SuperAdminEditBroker(c *fiber.Ctx) error    { return c.SendString("Edit Broker - TODO") }

// Delete handlers
func SuperAdminDeleteShipper(c *fiber.Ctx) error {
	id := c.Params("id")
	db := config.GetDB()
	
	if err := db.Delete(&models.User{}, "id = ?", id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete shipper"})
	}
	
	return c.JSON(fiber.Map{"status": "success", "message": "Shipper deleted"})
}

func SuperAdminDeleteCarrier(c *fiber.Ctx) error {
	id := c.Params("id")
	db := config.GetDB()
	
	if err := db.Delete(&models.User{}, "id = ?", id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete carrier"})
	}
	
	return c.JSON(fiber.Map{"status": "success", "message": "Carrier deleted"})
}

func SuperAdminDeleteBroker(c *fiber.Ctx) error {
	id := c.Params("id")
	db := config.GetDB()
	
	if err := db.Delete(&models.User{}, "id = ?", id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete broker"})
	}
	
	return c.JSON(fiber.Map{"status": "success", "message": "Broker deleted"})
}

func SuperAdminDeleteCompany(c *fiber.Ctx) error {
	id := c.Params("id")
	db := config.GetDB()
	
	if err := db.Delete(&models.Company{}, "id = ?", id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete company"})
	}
	
	return c.JSON(fiber.Map{"status": "success", "message": "Company deleted"})
}

func SuperAdminDeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	db := config.GetDB()
	
	if err := db.Delete(&models.User{}, "id = ?", id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete user"})
	}
	
	return c.JSON(fiber.Map{"status": "success", "message": "User deleted"})
}

