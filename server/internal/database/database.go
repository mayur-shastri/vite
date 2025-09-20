// internal/database/database.go
package database

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect establishes a connection to the database and returns the connection object.
func Connect() *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	log.Println("✅ Successfully connected to the database!")
	return db
}

// Migrate runs the database migrations for all the models.
func Migrate(db *gorm.DB) {
	log.Println("Running database migrations...")
	err := db.AutoMigrate(
		&User{},
		&PhysicalFile{},
		&Folder{},
		&UserFile{},
		&FileShare{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}
	log.Println("✅ Database migrations successful!")
}
