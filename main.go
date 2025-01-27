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
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("CargoZig Backend Golang")
	// load .env file if it exists
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found, using environment variables")
	}
	// Initialize database connection
	db := config.GetDB()
	// Initialize handlers with database
	handlers.InitDB(db)
	engine := html.New("./views", ".html")
	// Pass the engine to the Fiber config
	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: "layouts/main", // This will be your default layout
	})
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
	// CSRF Protection
	app.Use(csrf.New(csrf.Config{
		KeyLookup:      "header:X-Csrf-Token", // Keep this simple header-based approach
		CookieName:     "csrf_",
		CookieSameSite: "Lax",
		Expiration:     24 * time.Hour,
		KeyGenerator:   utils.UUIDv4,
	}))

	// Routes
	// Group
	admin := app.Group("/admin")

	admin.Get("/", handlers.LandingPage)

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
