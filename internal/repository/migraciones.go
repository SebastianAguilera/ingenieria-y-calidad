package repository

import (
	"fmt"

	"gorm.io/gorm"

	"ingenieria-y-calidad/internal/repository/models"
)

// MigrarEsquema crea o actualiza el esquema de la base de datos a partir de
// los modelos de persistencia. Es idempotente: puede ejecutarse en cada arranque.
func MigrarEsquema(db *gorm.DB) error {
	if err := db.AutoMigrate(&models.ProyectoModel{}, &models.UsuarioModel{}); err != nil {
		return fmt.Errorf("migrando el esquema: %w", err)
	}
	return nil
}
