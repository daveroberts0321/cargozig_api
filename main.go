package main

import (
	"cargozig_api/config"
	"cargozig_api/handlers"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("CargoZig Backend Golang")
	// load .env file if it exists
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found")
	}

	engine := html.New("./views", ".html")
	engine.Reload(true) // Enable template reloading in development

	// Pass the engine to the Fiber config
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	// Initialize database connection during startup
	db := config.GetDB()

	// Test the connection to ensure it's working
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	// Ping the database to verify the connection is alive
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173, https://cargozig.com, https://dashboard.cargozig.com",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Csrf-Token",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
	}))
	app.Use(limiter.New(limiter.Config{
		Max:               100,
		Expiration:        60 * time.Second,
		LimiterMiddleware: limiter.SlidingWindow{},
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use IP + User ID if authenticated for more accurate limiting
			return c.IP()
		},
	}))

	// Routes
	// Public routes (no auth required)
	handlers.SetupPublicRoutes(app) // public routes

	// Group
	adminGroup := app.Group("/toc")             // tactical operations center group admin- index page group
	superAdminGroup := app.Group("/superadmin") // super admin portal for platform management
	apiGroup := app.Group("/api")
	//mcpv1Group := app.Group("/mcpv1")

	// Register routes from handlers.
	handlers.SetupAdminRoutes(adminGroup)           // deathstar
	handlers.SetupSuperAdminRoutes(superAdminGroup) // super admin portal
	handlers.SetupApiAuthRoutes(apiGroup)           // api
	//handlers.SetupMcpv1Routes(mcpv1Group) // mcpv1

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	fmt.Printf("Server starting on port %s...\n", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
