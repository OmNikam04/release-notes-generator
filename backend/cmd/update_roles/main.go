package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID    string `gorm:"type:uuid;primaryKey"`
	Email string `gorm:"uniqueIndex;not null"`
	Role  string `gorm:"type:varchar(50)"`
}

func main() {
	// Load .env file
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Get database URL from environment
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable is not set")
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("✅ Connected to database")

	// Update developer role
	result := db.Model(&User{}).Where("email = ?", "om.nikam@arista.com").Update("role", "developer")
	if result.Error != nil {
		log.Fatalf("Failed to update developer role: %v", result.Error)
	}
	fmt.Printf("✅ Updated developer role (rows affected: %d)\n", result.RowsAffected)

	// Update manager role
	result = db.Model(&User{}).Where("email = ?", "devang.vyas@arista.com").Update("role", "manager")
	if result.Error != nil {
		log.Fatalf("Failed to update manager role: %v", result.Error)
	}
	fmt.Printf("✅ Updated manager role (rows affected: %d)\n", result.RowsAffected)

	// Verify updates
	var users []User
	db.Where("email IN ?", []string{"om.nikam@arista.com", "devang.vyas@arista.com"}).Find(&users)

	fmt.Println("\n📋 Updated users:")
	for _, user := range users {
		fmt.Printf("  - %s: %s\n", user.Email, user.Role)
	}

	fmt.Println("\n🎉 Roles updated successfully!")
}

