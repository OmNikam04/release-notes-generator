package db

import (
	"fmt"
	"log"

	"github.com/omnikam04/release-notes-generator/internal/models"
	"gorm.io/gorm"
)

// RunMigrations runs all database migrations
func RunMigrations(db *gorm.DB) error {
	// Enable UUID extensions for PostgreSQL
	if err := enableUUIDExtensions(db); err != nil {
		return fmt.Errorf("failed to enable UUID extensions: %w", err)
	}

	// Run custom migrations BEFORE auto-migrate to handle schema changes
	if err := runCustomMigrations(db); err != nil {
		return fmt.Errorf("failed to run custom migrations: %w", err)
	}

	// Auto-migrate all models
	if err := migrateModels(db); err != nil {
		return fmt.Errorf("failed to migrate models: %w", err)
	}

	fmt.Println("✅ Database migrations completed successfully")
	return nil
}

// enableUUIDExtensions enables PostgreSQL UUID extensions
func enableUUIDExtensions(db *gorm.DB) error {
	// Enable uuid-ossp extension (older PostgreSQL versions)
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		log.Printf("Warning: Failed to create uuid-ossp extension: %v", err)
	}

	// Enable pgcrypto extension (PostgreSQL 13+, provides gen_random_uuid())
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\"").Error; err != nil {
		log.Printf("Warning: Failed to create pgcrypto extension: %v", err)
	}

	return nil
}

// migrateModels runs auto-migration for all models
func migrateModels(db *gorm.DB) error {
	// Add all your models here
	models := []interface{}{
		&models.User{},
		&models.RefreshToken{},
	}

	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate %T: %w", model, err)
		}
		log.Printf("✅ Migrated model: %T", model)
	}

	return nil
}

// runCustomMigrations runs custom SQL migrations that can't be handled by AutoMigrate
// This handles schema changes like dropping columns, renaming columns, etc.
func runCustomMigrations(db *gorm.DB) error {
	// Migration 1: Remove 'name' and 'password' columns from users table if they exist
	if db.Migrator().HasColumn(&models.User{}, "name") {
		if err := db.Migrator().DropColumn(&models.User{}, "name"); err != nil {
			log.Printf("Warning: Failed to drop 'name' column from users: %v", err)
		} else {
			log.Println("✅ Dropped 'name' column from users table")
		}
	}

	if db.Migrator().HasColumn(&models.User{}, "password") {
		if err := db.Migrator().DropColumn(&models.User{}, "password"); err != nil {
			log.Printf("Warning: Failed to drop 'password' column from users: %v", err)
		} else {
			log.Println("✅ Dropped 'password' column from users table")
		}
	}

	// Migration 2: Add 'role' column if it doesn't exist (will be handled by AutoMigrate, but we can add default)
	if !db.Migrator().HasColumn(&models.User{}, "role") {
		// Add role column with default value
		if err := db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR NOT NULL DEFAULT 'developer'").Error; err != nil {
			log.Printf("Warning: Failed to add 'role' column: %v", err)
		} else {
			log.Println("✅ Added 'role' column to users table")
		}
	}

	log.Println("✅ Custom migrations completed")
	return nil
}

// DropAllTables drops all tables (use with caution!)
// Only use this in development/testing
func DropAllTables(db *gorm.DB) error {
	models := []interface{}{
		&models.RefreshToken{}, // Drop refresh_tokens first due to foreign key
		&models.User{},
	}

	for _, model := range models {
		if err := db.Migrator().DropTable(model); err != nil {
			return fmt.Errorf("failed to drop table for %T: %w", model, err)
		}
		log.Printf("⚠️  Dropped table: %T", model)
	}

	return nil
}

// ResetDatabase drops and recreates all tables (use with caution!)
// Only use this in development/testing
func ResetDatabase(db *gorm.DB) error {
	log.Println("⚠️  Resetting database...")

	if err := DropAllTables(db); err != nil {
		return err
	}

	if err := RunMigrations(db); err != nil {
		return err
	}

	log.Println("✅ Database reset completed")
	return nil
}
