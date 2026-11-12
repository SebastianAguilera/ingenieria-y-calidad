package repository

import (
	"fmt"

	"gorm.io/gorm"

	"ingenieria-y-calidad/internal/repository/models"
)

// MigrarEsquema crea o actualiza el esquema de la base de datos a partir de
// los modelos de persistencia. Es idempotente: puede ejecutarse en cada arranque.
func MigrarEsquema(db *gorm.DB) error {
	if err := db.AutoMigrate(&models.ProyectoModel{}, &models.UsuarioModel{}, &models.ProyectoUsuarioModel{}, &models.HistoriaModel{}, &models.SprintModel{}, &models.SprintBacklogModel{}, &models.DefectoModel{}, &models.AcceptanceCriterionModel{}); err != nil {
		return fmt.Errorf("migrando el esquema: %w", err)
	}
	if err := migrarRolesDeProyecto(db); err != nil {
		return err
	}
	return nil
}

// migrarRolesDeProyecto aplica US-11 a los datos existentes de forma idempotente:
//  1. el rol por defecto de un vínculo nuevo pasa a MIEMBRO;
//  2. los roles que ya no existen (los 3 de US-08) pasan a MIEMBRO;
//  3. en cada proyecto sin líder, el integrante de menor id de usuario (proxy
//     del más antiguo: la tabla puente no guarda fecha) pasa a LIDER.
func migrarRolesDeProyecto(db *gorm.DB) error {
	sentencias := []string{
		`ALTER TABLE proyecto_usuarios ALTER COLUMN rol SET DEFAULT 'MIEMBRO'`,
		`UPDATE proyecto_usuarios SET rol = 'MIEMBRO' WHERE rol NOT IN ('LIDER', 'MIEMBRO')`,
		`UPDATE proyecto_usuarios pu SET rol = 'LIDER'
		 WHERE pu.usuario_id = (
		     SELECT MIN(menor.usuario_id) FROM proyecto_usuarios menor WHERE menor.proyecto_id = pu.proyecto_id)
		   AND NOT EXISTS (
		     SELECT 1 FROM proyecto_usuarios lider WHERE lider.proyecto_id = pu.proyecto_id AND lider.rol = 'LIDER')`,
	}
	for _, sentencia := range sentencias {
		if err := db.Exec(sentencia).Error; err != nil {
			return fmt.Errorf("migrando los roles de proyecto: %w", err)
		}
	}
	return nil
}
