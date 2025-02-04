package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// SetupApiRoutes registers all API routes.
func SetupApiRoutes(router fiber.Router) {
	router.Get("/ping", Ping)
	// Add more API routes here.
}

// @Summary Health check endpoint
// @Description Simple ping-pong no-auth response to verify API is running
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Returns pong message"
// @Router api/ping [get]
func Ping(c *fiber.Ctx) error {
	// Send simple JSON response with "pong" message
	// This endpoint is used for health checks and API verification
	return c.JSON(fiber.Map{
		"status":  "success",
		"code":    200,
		"message": "pong",
	})
}
