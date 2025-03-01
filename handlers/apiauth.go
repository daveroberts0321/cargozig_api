package handlers

import (
	"cargozig_api/config"
	"cargozig_api/models"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// SetupApiAuthRoutes sets up the authentication routes
func SetupApiAuthRoutes(router fiber.Router) {
	router.Post("/login", UserLogin)
	router.Post("/logout", UserLogout)
	router.Get("/protected", ProtectedRoute)
	router.Post("/register", RegisterNewUser)
	router.Post("/newuserregistration", NewUserRegistration) // New endpoint
}

// GenerateJWT creates a JWT token
func GenerateJWT(userID string, roles []models.Role) (string, error) {
	// Convert roles to strings for JWT
	roleStrings := make([]string, len(roles))
	for i, role := range roles {
		roleStrings[i] = string(role)
	}

	// Define claims for the token
	claims := jwt.MapClaims{
		"user_id": userID,
		"roles":   roleStrings,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Expires in 24 hours
	}

	// Create a new token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Sign the token with the secret
	return token.SignedString(jwtSecret)
}

// RegisterNewUser handles user registration and sets a JWT cookie
func RegisterNewUser(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		fmt.Println("Error parsing request body:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate input
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing required fields"})
	}

	// Trim spaces
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	// Get database instance
	db := config.GetDB()

	// Check if user already exists
	var existingUser models.User
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "User already exists"})
	}

	// Hash password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error hashing password:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	// Create new user with default role
	newUser := models.User{
		Username:    req.Username,
		Email:       req.Email,
		Password:    string(hashedPassword),
		Roles:       []models.Role{models.RoleShipper}, // Default role
		Permissions: []models.Permission{},             // No custom permissions initially
		Active:      true,
	}

	// Save user to the database
	if err := db.Create(&newUser).Error; err != nil {
		fmt.Println("Error creating user:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create user"})
	}

	// Generate JWT token
	token, err := GenerateJWT(newUser.ID.String(), newUser.Roles)
	if err != nil {
		fmt.Println("Error generating token:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not generate token"})
	}

	// Set JWT as an HTTP-only cookie
	var secure bool = true
	var sameSite string = "Strict"
	if os.Getenv("ENVIRONMENT") == "development" {
		secure = false
		sameSite = "None"
	}

	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,     // Prevents JavaScript access
		Secure:   secure,   // Works only on HTTPS by default
		SameSite: sameSite, // Strict, Lax, None
	})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "User registered successfully",
		"user_id": newUser.ID,
		"token":   token,
		"user": fiber.Map{
			"id":       newUser.ID,
			"username": newUser.Username,
			"email":    newUser.Email,
			"roles":    newUser.Roles,
		},
	})
}

// NewUserRegistration handles the new comprehensive user registration endpoint
func NewUserRegistration(c *fiber.Ctx) error {
	var req struct {
		// User details
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"` // "admin", "shipper", or "carrier"

		// Company details
		CompanyName    string `json:"company_name"`
		CompanyEmail   string `json:"company_email"`
		CompanyPhone   string `json:"company_phone"`
		CompanyAddress string `json:"company_address"`
		CompanyCity    string `json:"company_city"`
		CompanyState   string `json:"company_state"`
		CompanyZipCode string `json:"company_zip_code"`
		CompanyCountry string `json:"company_country"`
		CompanyWebsite string `json:"company_website"`
		CompanyTaxID   string `json:"company_tax_id"`
		CompanyType    string `json:"company_type"` // "shipper", "carrier", "both"
	}

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate required fields
	if req.Username == "" || req.Email == "" || req.Password == "" ||
		req.CompanyName == "" || req.CompanyEmail == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing required fields"})
	}

	// Validate role
	validRole := false
	selectedRole := models.Role(req.Role)
	for _, role := range []models.Role{models.RoleAdmin, models.RoleShipper, models.RoleCarrier} {
		if selectedRole == role {
			validRole = true
			break
		}
	}
	if !validRole {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid role"})
	}

	// Trim spaces
	req.Email = strings.TrimSpace(req.Email)
	req.CompanyEmail = strings.TrimSpace(req.CompanyEmail)

	// Get database instance
	db := config.GetDB()

	// Check if user already exists
	var existingUser models.User
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "User with this email already exists"})
	}

	// Check if company already exists
	var existingCompany models.Company
	if err := db.Where("email = ?", req.CompanyEmail).First(&existingCompany).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Company with this email already exists"})
	}

	// Start a transaction
	tx := db.Begin()
	if tx.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	// Create new company
	company := models.Company{
		Name:        req.CompanyName,
		Email:       req.CompanyEmail,
		Phone:       req.CompanyPhone,
		Address:     req.CompanyAddress,
		City:        req.CompanyCity,
		State:       req.CompanyState,
		ZipCode:     req.CompanyZipCode,
		Country:     req.CompanyCountry,
		Website:     req.CompanyWebsite,
		TaxID:       req.CompanyTaxID,
		CompanyType: req.CompanyType,
		Active:      true,
		Verified:    false, // Require verification process
	}

	if err := tx.Create(&company).Error; err != nil {
		tx.Rollback()
		fmt.Println("Error creating company:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create company"})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		fmt.Println("Error hashing password:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	// Create new user with role
	now := time.Now()
	user := models.User{
		Username:    req.Username,
		Email:       req.Email,
		Password:    string(hashedPassword),
		CompanyID:   company.ID,
		Roles:       []models.Role{selectedRole},
		Permissions: []models.Permission{}, // No custom permissions initially
		Active:      true,
		LastLogin:   &now,
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		fmt.Println("Error creating user:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create user"})
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		fmt.Println("Error committing transaction:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not complete registration"})
	}

	// Generate JWT token
	token, err := GenerateJWT(user.ID.String(), user.Roles)
	if err != nil {
		fmt.Println("Error generating token:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not generate token"})
	}

	// Set JWT as an HTTP-only cookie
	var secure bool = true
	var sameSite string = "Strict"
	if os.Getenv("ENVIRONMENT") == "development" {
		secure = false
		sameSite = "None"
	}

	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Registration successful",
		"user_id": user.ID,
		"token":   token,
		"user": fiber.Map{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"roles":    user.Roles,
		},
		"company": fiber.Map{
			"id":   company.ID,
			"name": company.Name,
			"type": company.CompanyType,
		},
	})
}

// UserLogin authenticates the user and returns a JWT in a secure cookie
func UserLogin(c *fiber.Ctx) error {
	var loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Parse the request body
	if err := c.BodyParser(&loginRequest); err != nil {
		fmt.Println("Error parsing login request:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Get database instance
	db := config.GetDB()

	// Find user in the database
	var user models.User
	if err := db.Where("email = ?", loginRequest.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	// Check if user is active
	if !user.Active {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Account is disabled. Please contact support."})
	}

	// Compare the hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	// Update last login time
	now := time.Now()
	user.LastLogin = &now
	db.Save(&user)

	// Generate JWT token
	token, err := GenerateJWT(user.ID.String(), user.Roles)
	if err != nil {
		fmt.Println("Error generating token:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	// Set environment-specific cookie settings
	var secure bool = true
	var sameSite string = "Strict"
	if os.Getenv("ENVIRONMENT") == "development" {
		secure = false
		sameSite = "None"
	}

	// Set JWT as an HTTP-only cookie
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Login successful",
		"user": fiber.Map{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"roles":    user.Roles,
		},
	})
}

// UserLogout clears the authentication cookie
func UserLogout(c *fiber.Ctx) error {
	// Clear the auth token cookie
	var secure bool = true
	var sameSite string = "Strict"
	if os.Getenv("ENVIRONMENT") == "development" {
		secure = false
		sameSite = "None"
	}

	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour), // Expire immediately
		HTTPOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})

	return c.JSON(fiber.Map{"status": "success", "message": "Logged out"})
}

// ProtectedRoute example (require authentication)
func ProtectedRoute(c *fiber.Ctx) error {
	// Get JWT token from cookie
	tokenString := c.Cookies("auth_token")
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Validate JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
	}

	// Extract claims from token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token claims"})
	}

	// Extract user ID and roles
	userID, ok := claims["user_id"].(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user ID in token"})
	}

	// Convert roles from interface{} array to string array
	roleInterfaces, ok := claims["roles"].([]interface{})
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid roles in token"})
	}

	roles := make([]string, len(roleInterfaces))
	for i, role := range roleInterfaces {
		roleStr, ok := role.(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid role format in token"})
		}
		roles[i] = roleStr
	}

	// Set user info in request context for use in subsequent handlers
	c.Locals("user_id", userID)
	c.Locals("roles", roles)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Welcome to protected route!",
		"user_id": userID,
		"roles":   roles,
	})
}
