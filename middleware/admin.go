package middleware

import (
	"cargozig_api/config"
	"cargozig_api/models"

	"github.com/gofiber/fiber/v2"
)

// AdminAuthMiddleware protects routes that require admin authentication
func AdminAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get JWT token from cookie
		token := c.Cookies("admin_auth_token")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Admin authentication required",
			})
		}

		// Parse and validate the token
		claims, err := ParseJWT(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Check if user has admin role
		userID := claims["user_id"].(string)
		roles := claims["roles"].([]interface{})

		hasAdminRole := false
		for _, role := range roles {
			if role.(string) == string(models.RoleAdmin) {
				hasAdminRole = true
				break
			}
		}

		if !hasAdminRole {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Admin privileges required",
			})
		}

		// Get user from database to ensure they're still active
		db := config.GetDB()
		var user models.User
		if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User not found",
			})
		}

		// Check if user is active
		if !user.Active {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Account is disabled",
			})
		}

		// Store user info in context for later use
		c.Locals("admin_user", &user)
		c.Locals("admin_user_id", userID)

		return c.Next()
	}
}
