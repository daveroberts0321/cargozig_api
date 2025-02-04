package middleware

import (
	"cargozig_api/models"

	"github.com/gofiber/fiber/v2"
)

func RequirePermission(requiredPermission models.Permission) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.User) // Assuming you set this in auth middleware

		// Check if user has required permission
		for _, permission := range user.Permissions {
			if permission == requiredPermission || permission == models.SystemAdmin {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Insufficient permissions for this user to access this resource",
		})
	}
}
