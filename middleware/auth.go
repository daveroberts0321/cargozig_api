package middleware

import (
	"cargozig_api/config" // Import config to access the initialized database
	"cargozig_api/models"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// AuthenticateUser verifies the JWT token and loads the user information
func AuthenticateUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get JWT token from cookie - check both auth_token and admin_auth_token
		tokenString := c.Cookies("auth_token")
		if tokenString == "" {
			tokenString = c.Cookies("admin_auth_token")
		}

		// If no token in cookie, check Authorization header
		if tokenString == "" {
			auth := c.Get("Authorization")
			if auth != "" && strings.HasPrefix(auth, "Bearer ") {
				tokenString = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		// If still no token, return unauthorized
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Authentication required",
			})
		}

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return GetJWTSecret(), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid authentication token",
			})
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid token claims",
			})
		}

		// Extract user ID from claims
		userID, ok := claims["user_id"].(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid user identification",
			})
		}

		// Extract roles from claims
		var roles []models.Role
		if rolesInterface, exists := claims["roles"]; exists {
			if rolesArray, ok := rolesInterface.([]interface{}); ok {
				for _, roleInterface := range rolesArray {
					if roleStr, ok := roleInterface.(string); ok {
						roles = append(roles, models.Role(roleStr))
					}
				}
			}
		}

		// Store user info in context for downstream handlers
		c.Locals("user_id", userID)
		c.Locals("roles", roles)

		return c.Next()
	}
}

// RequireRole checks if the authenticated user has any of the required roles
func RequireRole(requiredRoles ...models.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// This middleware should be used after AuthenticateUser
		userRolesInterface := c.Locals("roles")
		if userRolesInterface == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Authentication required",
			})
		}

		// Convert roles to the correct type
		userRoles, ok := userRolesInterface.([]models.Role)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid role format",
			})
		}

		// Check if user has any of the required roles
		for _, userRole := range userRoles {
			for _, requiredRole := range requiredRoles {
				if userRole == requiredRole {
					return c.Next() // User has one of the required roles
				}
			}
		}

		// Check for admin role which overrides other role requirements
		for _, userRole := range userRoles {
			if userRole == models.RoleAdmin {
				return c.Next() // Admin role has access to everything
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "error",
			"message": "Access denied: insufficient privileges",
		})
	}
}

// RequirePermission checks if the authenticated user has the required permission
func RequirePermission(requiredPermission models.Permission) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// This middleware should be used after AuthenticateUser
		userID := c.Locals("user_id")
		if userID == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Authentication required",
			})
		}

		// Get user roles from context
		userRoles, ok := c.Locals("roles").([]models.Role)
		if !ok {
			// If roles aren't in context, we need to fetch the user from DB
			db := config.GetDB() // Get the database from config
			var user models.User
			if err := db.First(&user, "id = ?", userID).Error; err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"status":  "error",
					"message": "User not found",
				})
			}
			userRoles = user.Roles
		}

		// Check if any of the user's roles have the required permission
		for _, role := range userRoles {
			if role.HasPermission(requiredPermission) {
				return c.Next()
			}
		}

		// If admin role, allow access
		for _, role := range userRoles {
			if role == models.RoleAdmin {
				return c.Next()
			}
		}

		// Check for System Admin permission which overrides other permissions
		for _, role := range userRoles {
			permissions := models.DefaultRolePermissions[role]
			for _, perm := range permissions {
				if perm == models.SystemAdmin {
					return c.Next()
				}
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "error",
			"message": "Access denied: required permission not granted",
		})
	}
}

// LoadUser fetches the full user record and adds it to the request context
func LoadUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		if userID == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Authentication required",
			})
		}

		// Fetch the user from the database
		db := config.GetDB() // Get the database from config
		var user models.User
		if err := db.First(&user, "id = ?", userID).Error; err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "User not found",
			})
		}

		// Store the full user object in context
		c.Locals("user", &user)

		return c.Next()
	}
}

// RequireSuperAdmin ensures the user has super admin permissions and redirects if not
func RequireSuperAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(*models.User)
		if !ok || user == nil {
			return c.Redirect("/toc/login")
		}

		if !user.HasPermission(models.SystemAdmin) {
			// Redirect to their appropriate dashboard
			userType := c.Cookies("user_type")
			switch userType {
			case "broker":
				return c.Redirect("/toc/dashboard")
			case "shipper":
				return c.Redirect("/shipper/dashboard")
			case "carrier":
				return c.Redirect("/carrier/dashboard")
			default:
				return c.Redirect("/toc/login")
			}
		}

		return c.Next()
	}
}

// RequireBroker ensures the user has broker/admin role
func RequireBroker() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(*models.User)
		if !ok || user == nil {
			return c.Redirect("/toc/login")
		}

		// Check if user is a broker/admin
		hasAdminRole := false
		for _, role := range user.Roles {
			if role == models.RoleAdmin {
				hasAdminRole = true
				break
			}
		}

		if !hasAdminRole {
			// Redirect to their appropriate dashboard
			userType := c.Cookies("user_type")
			switch userType {
			case "superadmin":
				return c.Redirect("/superadmin/dashboard")
			case "shipper":
				return c.Redirect("/shipper/dashboard")
			case "carrier":
				return c.Redirect("/carrier/dashboard")
			default:
				return c.Redirect("/toc/login")
			}
		}

		return c.Next()
	}
}

// RequireShipper ensures the user has shipper role
func RequireShipper() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(*models.User)
		if !ok || user == nil {
			return c.Redirect("/toc/login")
		}

		// Check if user is a shipper
		hasShipperRole := false
		for _, role := range user.Roles {
			if role == models.RoleShipper {
				hasShipperRole = true
				break
			}
		}

		if !hasShipperRole {
			// Redirect to their appropriate dashboard
			userType := c.Cookies("user_type")
			switch userType {
			case "superadmin":
				return c.Redirect("/superadmin/dashboard")
			case "broker":
				return c.Redirect("/toc/dashboard")
			case "carrier":
				return c.Redirect("/carrier/dashboard")
			default:
				return c.Redirect("/toc/login")
			}
		}

		return c.Next()
	}
}

// RequireCarrier ensures the user has carrier role
func RequireCarrier() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(*models.User)
		if !ok || user == nil {
			return c.Redirect("/toc/login")
		}

		// Check if user is a carrier
		hasCarrierRole := false
		for _, role := range user.Roles {
			if role == models.RoleCarrier {
				hasCarrierRole = true
				break
			}
		}

		if !hasCarrierRole {
			// Redirect to their appropriate dashboard
			userType := c.Cookies("user_type")
			switch userType {
			case "superadmin":
				return c.Redirect("/superadmin/dashboard")
			case "broker":
				return c.Redirect("/toc/dashboard")
			case "shipper":
				return c.Redirect("/shipper/dashboard")
			default:
				return c.Redirect("/toc/login")
			}
		}

		return c.Next()
	}
}