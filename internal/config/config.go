package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// errJWTSecretFaltante se devuelve cuando JWT_SECRET no esta definida: sin
// secreto no se pueden firmar ni verificar tokens (US-10).
var errJWTSecretFaltante = errors.New("JWT_SECRET es obligatoria: definila en el archivo .env para poder emitir tokens")

// expiracionJWTPorDefecto es la vigencia del token cuando JWT_EXPIRY no esta
// definida.
const expiracionJWTPorDefecto = 24 * time.Hour

type Config struct {
	AppEnv     string
	Port       string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	JWTSecret  string
	JWTExpiry  time.Duration
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
			dbName = "metrics_db"
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errJWTSecretFaltante
	}

	jwtExpiry := expiracionJWTPorDefecto
	if valor := os.Getenv("JWT_EXPIRY"); valor != "" {
		d, err := time.ParseDuration(valor)
		if err != nil || d <= 0 {
			return nil, fmt.Errorf("JWT_EXPIRY invalida (%q): use una duracion valida de time.ParseDuration, por ejemplo 24h", valor)
		}
		jwtExpiry = d
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
		JWTSecret:  jwtSecret,
		JWTExpiry:  jwtExpiry,
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
