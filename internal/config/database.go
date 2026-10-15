package config

import "os"

type baseDeDatos struct {
	host     string
	port     string
	user     string
	password string
	name     string
	sslmode  string
}

func cargarBaseDeDatos(entorno string) baseDeDatos {
	return baseDeDatos{
		host:     getEnvOrDefault("DB_HOST", "localhost"),
		port:     getEnvOrDefault("DB_PORT", "5432"),
		user:     getEnvOrDefault("DB_USER", "postgres"),
		password: os.Getenv("DB_PASSWORD"),
		name:     nombreBaseDeDatos(entorno),
		sslmode:  getEnvOrDefault("DB_SSLMODE", "disable"),
	}
}

func nombreBaseDeDatos(entorno string) string {
	var especifico, defecto string

	switch entorno {
	case "test":
		especifico, defecto = "DB_NAME_TEST", "metrics_db_test"
	case "production":
		especifico, defecto = "DB_NAME_PROD", "metrics_db"
	default:
		especifico, defecto = "DB_NAME_DEV", "metrics_db_dev"
	}

	if nombre := os.Getenv(especifico); nombre != "" {
		return nombre
	}
	if nombre := os.Getenv("DB_NAME"); nombre != "" {
		return nombre
	}
	return defecto
}
