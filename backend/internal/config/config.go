package config

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Port      string
	DBUrl     string
	JWTSecret string
}

func Load() (*Config, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	viper.AutomaticEnv()

	cfg := &Config{
		Port:      viper.GetString("PORT"),
		DBUrl:     viper.GetString("DB_URL"),     // Match .env
		JWTSecret: viper.GetString("JWT_SECRET"), // Match .env
	}

	// Validate required fields
	if cfg.DBUrl == "" {
		return nil, fmt.Errorf("DB_URL is required")
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	// Set default port if not provided
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return cfg, nil
}
