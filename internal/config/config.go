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

	entorno := entornoApp()
	bd := cargarBaseDeDatos(entorno)

	return &Config{
		AppEnv:     entorno,
		Port:       getEnvOrDefault("PORT", "8081"),
		DBHost:     bd.host,
		DBPort:     bd.port,
		DBUser:     bd.user,
		DBPassword: bd.password,
		DBName:     bd.name,
		DBSSLMode:  bd.sslmode,
	}, nil
}

func entornoApp() string {
	if appEnv := os.Getenv("APP_ENV"); appEnv != "" {
		return appEnv
	}
	return "development"
}

func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
