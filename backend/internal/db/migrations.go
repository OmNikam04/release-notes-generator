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

	// Run post-migration fixes AFTER auto-migrate
	if err := runPostMigrationFixes(db); err != nil {
		return fmt.Errorf("failed to run post-migration fixes: %w", err)
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
	// Add all your models here in order of dependencies
	// Models with no foreign keys first, then models that depend on them
	models := []interface{}{
		&models.User{},
		&models.RefreshToken{},
		&models.Bug{},
		&models.ReleaseNote{},
		&models.Pattern{},
		&models.Feedback{},
		&models.FeedbackPattern{},
		&models.AuditLog{},
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

	// Migration 3: Create GIN indexes for JSONB columns (for pattern matching)
	// These indexes improve performance for JSONB queries
	createGINIndexes(db)

	log.Println("✅ Custom migrations completed")
	return nil
}

// runPostMigrationFixes runs migrations that need to happen AFTER AutoMigrate
// This is for fixing column types that AutoMigrate doesn't handle
func runPostMigrationFixes(db *gorm.DB) error {
	log.Println("🔧 Running post-migration fixes...")

	if !db.Migrator().HasTable(&models.Bug{}) {
		log.Println("⚠️  Bugs table does not exist, skipping post-migration fixes")
		return nil
	}

	// Fix 1: Alter bugsby_id column type from varchar(10) to varchar(50)
	// This is needed because bug IDs from Bugsby can be 6-7 digits
	// AutoMigrate doesn't change existing column types, so we need to do it manually
	alterColumnIfNeeded(db, "bugs", "bugsby_id", 10, 50)

	// Fix 2: Alter priority column type from varchar(10) to varchar(50)
	// This is needed because Bugsby may return priority values longer than 10 characters
	alterColumnIfNeeded(db, "bugs", "priority", 10, 50)

	log.Println("✅ Post-migration fixes completed")
	return nil
}

// alterColumnIfNeeded checks if a column needs to be altered and alters it if necessary
func alterColumnIfNeeded(db *gorm.DB, tableName, columnName string, oldLength, newLength int) {
	var result struct {
		DataType               string
		CharacterMaximumLength *int
	}
	err := db.Raw(`
		SELECT data_type, character_maximum_length
		FROM information_schema.columns
		WHERE table_schema = CURRENT_SCHEMA()
		AND table_name = ?
		AND column_name = ?
	`, tableName, columnName).Scan(&result).Error

	if err != nil {
		log.Printf("❌ Could not check %s.%s column type: %v", tableName, columnName, err)
		return
	}

	// Log the current column type with proper dereferencing
	if result.CharacterMaximumLength != nil {
		log.Printf("🔍 Current %s.%s column type: %s(%d)", tableName, columnName, result.DataType, *result.CharacterMaximumLength)
	} else {
		log.Printf("🔍 Current %s.%s column type: %s(NULL)", tableName, columnName, result.DataType)
	}

	// Check if it needs to be altered
	if result.DataType == "character varying" && result.CharacterMaximumLength != nil && *result.CharacterMaximumLength == oldLength {
		// Column exists and is varchar(oldLength), need to alter it
		log.Printf("⚠️  %s.%s column is varchar(%d), altering to varchar(%d)...", tableName, columnName, oldLength, newLength)
		if err := db.Exec(fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE varchar(%d)", tableName, columnName, newLength)).Error; err != nil {
			log.Printf("❌ Failed to alter %s.%s column type: %v", tableName, columnName, err)
		} else {
			log.Printf("✅ Altered %s.%s column from varchar(%d) to varchar(%d)", tableName, columnName, oldLength, newLength)
		}
	} else if result.CharacterMaximumLength != nil && *result.CharacterMaximumLength == newLength {
		log.Printf("✅ %s.%s column is already varchar(%d), no migration needed", tableName, columnName, newLength)
	} else if result.CharacterMaximumLength != nil {
		log.Printf("ℹ️  %s.%s column is varchar(%d), no migration needed", tableName, columnName, *result.CharacterMaximumLength)
	}
}

// createGINIndexes creates GIN indexes for JSONB columns
func createGINIndexes(db *gorm.DB) {
	indexes := []struct {
		table  string
		column string
		name   string
	}{
		{"feedbacks", "bug_context", "idx_feedback_bug_context"},
		{"feedbacks", "extracted_patterns", "idx_feedback_patterns"},
		{"patterns", "applicable_when", "idx_pattern_applicable"},
	}

	for _, idx := range indexes {
		sql := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s USING GIN (%s)", idx.name, idx.table, idx.column)
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("Warning: Failed to create GIN index %s: %v", idx.name, err)
		} else {
			log.Printf("✅ Created GIN index: %s", idx.name)
		}
	}
}

// DropAllTables drops all tables (use with caution!)
// Only use this in development/testing
func DropAllTables(db *gorm.DB) error {
	// Drop tables in reverse order of dependencies
	models := []interface{}{
		&models.AuditLog{},        // No dependencies on other tables (except User, but uses SET NULL)
		&models.FeedbackPattern{}, // Depends on Feedback and Pattern
		&models.Feedback{},        // Depends on ReleaseNote, Bug, User
		&models.Pattern{},         // No dependencies
		&models.ReleaseNote{},     // Depends on Bug
		&models.Bug{},             // Depends on User
		&models.RefreshToken{},    // Depends on User
		&models.User{},            // Base table
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
