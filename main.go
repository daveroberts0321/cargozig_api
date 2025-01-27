package main

import (
	"cargozig_api/config"
	"cargozig_api/handlers"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
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
	app.Use(cors.New(cors.Config{}))

	// Routes
	app.Get("/", handlers.LandingPage)

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
