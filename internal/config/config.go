package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv     string
	Port       string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	var dbName string
	switch appEnv {
	case "test":
		dbName = os.Getenv("DB_NAME_TEST")
		if dbName == "" {
			dbName = "metrics_db_test"
		}
	case "production":
		dbName = os.Getenv("DB_NAME_PROD")
		if dbName == "" {
			dbName = "metrics_db"
		}
	default: // development
		dbName = os.Getenv("DB_NAME_DEV")
		if dbName == "" {
			dbName = "metrics_db_dev"
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	cfg := &Config{
		AppEnv:     appEnv,
		Port:       port,
		DBHost:     getEnvOrDefault("DB_HOST", "localhost"),
		DBPort:     getEnvOrDefault("DB_PORT", "5432"),
		DBUser:     getEnvOrDefault("DB_USER", "postgres"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     dbName,
		DBSSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
