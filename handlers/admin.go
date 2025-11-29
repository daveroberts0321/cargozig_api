package handlers

import (
	"cargozig_api/config"
	"cargozig_api/middleware"
	"cargozig_api/models"
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
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

	// Development-only route to create super admins
	router.Post("/dev/create-super-admin", DevCreateSuperAdmin)
	router.Get("/dev/create-super-admin", DevCreateSuperAdminPage)

	router.Get("/", LandingPage) // Root route should come last
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
		Roles:       models.RoleArray{models.RoleAdmin}, // Admin role only
		Permissions: models.PermissionArray{},           // No custom permissions initially
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

// DevCreateSuperAdminPage serves a simple form to create super admins (development only)
func DevCreateSuperAdminPage(c *fiber.Ctx) error {
	// Security check - only allow in development mode
	env := os.Getenv("APP_ENV")
	if env == "production" {
		return c.Status(fiber.StatusForbidden).SendString("This endpoint is disabled in production")
	}

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Create Super Admin - Development Only</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100 min-h-screen flex items-center justify-center">
    <div class="bg-white p-8 rounded-lg shadow-lg max-w-md w-full">
        <div class="bg-yellow-100 border-l-4 border-yellow-500 text-yellow-700 p-4 mb-6">
            <p class="font-bold">⚠️ Development Only</p>
            <p class="text-sm">This endpoint is disabled in production</p>
        </div>
        
        <h1 class="text-2xl font-bold mb-6 text-gray-900">Create Super Admin</h1>
        
        <form id="createAdminForm">
            <div class="mb-4">
                <label class="block text-gray-700 text-sm font-bold mb-2" for="username">
                    Username
                </label>
                <input 
                    class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline" 
                    id="username" 
                    type="text" 
                    required
                    placeholder="Enter username">
            </div>
            
            <div class="mb-4">
                <label class="block text-gray-700 text-sm font-bold mb-2" for="email">
                    Email
                </label>
                <input 
                    class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline" 
                    id="email" 
                    type="email" 
                    required
                    placeholder="Enter email">
            </div>
            
            <div class="mb-6">
                <label class="block text-gray-700 text-sm font-bold mb-2" for="password">
                    Password
                </label>
                <input 
                    class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 mb-3 leading-tight focus:outline-none focus:shadow-outline" 
                    id="password" 
                    type="password" 
                    required
                    placeholder="Enter password">
            </div>
            
            <div id="message" class="mb-4 hidden"></div>
            
            <button 
                class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline w-full" 
                type="submit">
                Create Super Admin
            </button>
        </form>
        
        <div class="mt-4 text-center">
            <a href="/toc/login" class="text-blue-500 hover:text-blue-700 text-sm">Go to Admin Login</a>
        </div>
    </div>

    <script>
        document.getElementById('createAdminForm').addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const messageDiv = document.getElementById('message');
            messageDiv.classList.add('hidden');
            
            const formData = {
                username: document.getElementById('username').value,
                email: document.getElementById('email').value,
                password: document.getElementById('password').value
            };
            
            try {
                const response = await fetch('/toc/dev/create-super-admin', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(formData)
                });
                
                const data = await response.json();
                
                if (response.ok) {
                    messageDiv.className = 'bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded mb-4';
                    messageDiv.textContent = data.message || 'Super admin created successfully!';
                    messageDiv.classList.remove('hidden');
                    document.getElementById('createAdminForm').reset();
                    
                    // Redirect to login after 2 seconds
                    setTimeout(() => {
                        window.location.href = '/toc/login';
                    }, 2000);
                } else {
                    messageDiv.className = 'bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4';
                    messageDiv.textContent = data.error || 'Failed to create admin';
                    messageDiv.classList.remove('hidden');
                }
            } catch (error) {
                messageDiv.className = 'bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4';
                messageDiv.textContent = 'Network error: ' + error.message;
                messageDiv.classList.remove('hidden');
            }
        });
    </script>
</body>
</html>`

	return c.Type("html").SendString(html)
}

// DevCreateSuperAdmin creates a super admin user (development only)
func DevCreateSuperAdmin(c *fiber.Ctx) error {
	// Security check - only allow in development mode
	env := os.Getenv("APP_ENV")
	if env == "production" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "This endpoint is disabled in production"})
	}

	fmt.Println("=== DevCreateSuperAdmin called ===")

	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		fmt.Println("Error parsing super admin request:", err)
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
	if err := db.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "User with this email or username already exists"})
	}

	// Create or get a default "System Admin" company
	var systemCompany models.Company
	if err := db.Where("email = ?", "system@cargozig.com").First(&systemCompany).Error; err != nil {
		// Company doesn't exist, create it
		systemCompany = models.Company{
			Name:  "System Admin",
			Email: "system@cargozig.com",
		}
		if err := db.Create(&systemCompany).Error; err != nil {
			fmt.Println("Error creating system company:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create system company"})
		}
		fmt.Printf("System company created: %s\n", systemCompany.ID)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error hashing password:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	// Create super admin user with all roles and permissions
	superAdmin := models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		CompanyID: systemCompany.ID,
		Active:    true,
	}

	// Save to database first without roles/permissions
	if err := db.Create(&superAdmin).Error; err != nil {
		fmt.Println("Error creating super admin:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create super admin"})
	}

	// Now update with roles and permissions using raw SQL to properly handle PostgreSQL arrays
	roles := pq.Array([]string{"admin", "shipper", "carrier"})
	permissions := pq.Array([]string{
		"system_admin",
		"manage_users",
		"create_shipment",
		"view_shipment",
		"edit_shipment",
		"delete_shipment",
		"manage_rates",
		"view_rates",
		"add_routes",
		"view_routes",
		"view_financials",
		"manage_payments",
		"manage_settings",
		"view_settings",
		"view_users",
	})

	if err := db.Model(&superAdmin).Updates(map[string]interface{}{
		"roles":       roles,
		"permissions": permissions,
	}).Error; err != nil {
		fmt.Println("Error updating super admin roles/permissions:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to set super admin permissions"})
	}

	// Reload the user to get the updated roles and permissions
	if err := db.First(&superAdmin, superAdmin.ID).Error; err != nil {
		fmt.Println("Error reloading super admin:", err)
	}

	fmt.Printf("Super admin created successfully: %s (%s)\n", superAdmin.Username, superAdmin.Email)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Super admin created successfully",
		"user_id": superAdmin.ID,
		"user": fiber.Map{
			"id":          superAdmin.ID,
			"username":    superAdmin.Username,
			"email":       superAdmin.Email,
			"roles":       superAdmin.Roles,
			"permissions": superAdmin.Permissions,
		},
	})
}
