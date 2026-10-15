package repository

import (
	"fmt"

	"gorm.io/gorm"

	"ingenieria-y-calidad/internal/repository/models"
)

func MigrarEsquema(db *gorm.DB) error {
	if err := db.AutoMigrate(&models.ProyectoModel{}, &models.UsuarioModel{}); err != nil {
		return fmt.Errorf("migrando el esquema: %w", err)
	}
	return nil
}
