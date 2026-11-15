package repository

import (
	"fmt"
	"log"
	"os"
	"time"

	"ingenieria-y-calidad/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConnectDB abre la conexion con PostgreSQL a partir de la configuracion.
// TranslateError traduce los errores nativos del driver a los errores
// centinelas de GORM, de modo que las capas superiores puedan distinguir una
// violacion de unicidad sin depender de PostgreSQL. IgnoreRecordNotFoundError
// evita registrar como error el "record not found", que en este dominio es una
// respuesta de negocio esperada.
func ConnectDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
		cfg.DBSSLMode,
	)

	gormConfig := &gorm.Config{
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

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("error conectando a la base de datos: %w", err)
	}

	// Connection pooling configuration (KISS/YAGNI production-ready standard)
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(5)
	}

	return db, nil
}
