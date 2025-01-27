package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// LandingPage renders the home page
func LandingPage(c *fiber.Ctx) error {
	return c.Render("index", fiber.Map{
		"Title": "CargoZig - Smart Freight Brokerage Platform",
	}, "layouts/main")
}
