package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// SetupAdminRoutes registers all dashboard pages.
func SetupAdminRoutes(router fiber.Router) {
	router.Get("/", LandingPage)
	router.Get("/dashboard", LandingPage)
	// Add more admin routes here.
}

// LandingPage renders the home page
func LandingPage(c *fiber.Ctx) error {
	return c.Render("index", fiber.Map{
		"Title": "CargoZig - Smart Freight Brokerage Platform",
	}, "layouts/main")
}
