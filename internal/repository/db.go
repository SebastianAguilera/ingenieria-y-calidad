package repository

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConnectDB abre la conexion con PostgreSQL a partir de las variables de
// entorno. TranslateError traduce los errores nativos del driver a los errores
// centinelas de GORM, de modo que las capas superiores puedan distinguir una
// violacion de unicidad sin depender de PostgreSQL. IgnoreRecordNotFoundError
// evita registrar como error el "record not found", que en este dominio es una
// respuesta de negocio esperada.
func ConnectDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
	)

	config := &gorm.Config{
		TranslateError: true,
		Logger: logger.New(
			log.New(os.Stdout, "", log.LstdFlags),
			logger.Config{
				SlowThreshold:             500 * time.Millisecond,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
			},
		),
	}

	db, err := gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return nil, fmt.Errorf("error conectando a la base de datos: %w", err)
	}

	return db, nil
}
