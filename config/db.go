package config

import (
	"cargozig_api/models"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db     *gorm.DB
	dbOnce sync.Once
)

// GetDB returns the database instance, initializing it if necessary
func GetDB() *gorm.DB {
	dbOnce.Do(func() {
		var err error
		db, err = InitDB()
		if err != nil {
			log.Fatal("Failed to initialize database:", err)
		}
	})
	return db
}

// InitDB initializes the database connection
func InitDB() (*gorm.DB, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found, using environment variables")
	}

	// Get database connection string from environment
	dbString := os.Getenv("DB_STRING")
	if dbString == "" {
		return nil, fmt.Errorf("DB_STRING environment variable is not set")
	}

	// Configure GORM logger
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,        // Enable color
		},
	)

	// Open database connection
	db, err := gorm.Open(postgres.Open(dbString), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC() // Use UTC for all timestamps
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Enable PostGIS extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS postgis").Error; err != nil {
		return nil, fmt.Errorf("failed to enable PostGIS extension: %v", err)
	}

	// Auto-migrate the database schema
	if err := db.AutoMigrate(
		&models.User{},
		&models.Company{},
	); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate database: %v", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %v", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)           // Maximum number of idle connections
	sqlDB.SetMaxOpenConns(100)          // Maximum number of open connections
	sqlDB.SetConnMaxLifetime(time.Hour) // Maximum lifetime of a connection

	return db, nil
}
