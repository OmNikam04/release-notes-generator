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

	// Auto-migrate all models
	if err := migrateModels(db); err != nil {
		return fmt.Errorf("failed to migrate models: %w", err)
	}

	// Run custom migrations
	if err := runCustomMigrations(db); err != nil {
		return fmt.Errorf("failed to run custom migrations: %w", err)
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
func runCustomMigrations(db *gorm.DB) error {
	// Drop the old unique index if it exists
	if err := db.Exec("DROP INDEX IF EXISTS idx_bikes_registration_number").Error; err != nil {
		log.Printf("Warning: Failed to drop old index: %v", err)
	}

	// Create a partial unique index on registration_number that excludes soft-deleted records
	// This allows the same registration number to be reused after a bike is soft-deleted
	sql := `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_bikes_registration_number_active
		ON bikes (registration_number)
		WHERE deleted_at IS NULL AND registration_number != ''
	`
	if err := db.Exec(sql).Error; err != nil {
		return fmt.Errorf("failed to create partial unique index on bikes.registration_number: %w", err)
	}

	log.Println("✅ Custom migrations completed")
	return nil
}

// DropAllTables drops all tables (use with caution!)
// Only use this in development/testing
func DropAllTables(db *gorm.DB) error {
	models := []interface{}{
		&models.User{},
		&models.RefreshToken{},
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
