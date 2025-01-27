package handlers

import (
	"sync"

	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	once sync.Once
)

// InitDB initializes the database instance for all handlers
func InitDB(database *gorm.DB) {
	once.Do(func() {
		db = database
	})
}
