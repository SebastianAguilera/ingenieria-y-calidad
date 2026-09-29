package repository

import (
	"fmt"

	"gorm.io/gorm"

	"ingenieria-y-calidad/internal/domain"
)

// MigrarEsquema crea o actualiza el esquema de la base de datos a partir de
// las entidades de dominio. Es idempotente: puede ejecutarse en cada arranque.
func MigrarEsquema(db *gorm.DB) error {
	if err := db.AutoMigrate(&domain.Proyecto{}, &domain.Integrante{}); err != nil {
		return fmt.Errorf("migrando el esquema: %w", err)
	}
	return nil
}
