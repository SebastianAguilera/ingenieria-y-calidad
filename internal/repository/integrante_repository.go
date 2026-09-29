package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"ingenieria-y-calidad/internal/domain"
)

type integranteRepository struct {
	db *gorm.DB
}

// NuevoIntegranteRepository construye el repositorio de integrantes sobre GORM.
func NuevoIntegranteRepository(db *gorm.DB) *integranteRepository {
	return &integranteRepository{db: db}
}

var _ domain.IntegranteRepository = (*integranteRepository)(nil)

// proyectoIntegrante modela la tabla puente de la relacion muchos a muchos.
type proyectoIntegrante struct {
	ProyectoID   uint `gorm:"column:proyecto_id;primaryKey"`
	IntegranteID uint `gorm:"column:integrante_id;primaryKey"`
}

func (proyectoIntegrante) TableName() string {
	return "proyecto_integrantes"
}

// ObtenerPorEmail busca una persona por su clave natural normalizada.
func (r *integranteRepository) ObtenerPorEmail(ctx context.Context, email string) (*domain.Integrante, error) {
	var integrante domain.Integrante
	err := r.db.WithContext(ctx).
		Where("email = ?", domain.NormalizarEmail(email)).
		First(&integrante).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrIntegranteNoEncontrado
		}
		return nil, fmt.Errorf("buscando integrante por email: %w", err)
	}
	return &integrante, nil
}

// Crear inserta un integrante. El email es unico en todo el sistema.
func (r *integranteRepository) Crear(ctx context.Context, i *domain.Integrante) error {
	i.Email = domain.NormalizarEmail(i.Email)
	if err := r.db.WithContext(ctx).Create(i).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrIntegranteYaAsociado
		}
		return fmt.Errorf("creando integrante: %w", err)
	}
	return nil
}

// ActualizarNombre refresca el nombre de un integrante preexistente.
func (r *integranteRepository) ActualizarNombre(ctx context.Context, id uint, nombre string) error {
	if err := r.db.WithContext(ctx).
		Model(&domain.Integrante{}).
		Where("id = ?", id).
		Update("nombre", strings.TrimSpace(nombre)).Error; err != nil {
		return fmt.Errorf("actualizando nombre del integrante %d: %w", id, err)
	}
	return nil
}

// EstaAsociado informa si la persona ya forma parte del equipo del proyecto.
func (r *integranteRepository) EstaAsociado(ctx context.Context, proyectoID, integranteID uint) (bool, error) {
	var vinculo proyectoIntegrante
	err := r.db.WithContext(ctx).
		Where("proyecto_id = ? AND integrante_id = ?", proyectoID, integranteID).
		First(&vinculo).Error
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return false, nil
	default:
		return false, fmt.Errorf("verificando asociacion del integrante: %w", err)
	}
}

// Vincular agrega el vinculo entre proyecto e integrante.
func (r *integranteRepository) Vincular(ctx context.Context, proyectoID, integranteID uint) error {
	vinculo := proyectoIntegrante{ProyectoID: proyectoID, IntegranteID: integranteID}
	if err := r.db.WithContext(ctx).Create(&vinculo).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrIntegranteYaAsociado
		}
		return fmt.Errorf("vinculando integrante: %w", err)
	}
	return nil
}

// Desvincular quita el vinculo entre proyecto e integrante.
func (r *integranteRepository) Desvincular(ctx context.Context, proyectoID, integranteID uint) error {
	resultado := r.db.WithContext(ctx).
		Where("proyecto_id = ? AND integrante_id = ?", proyectoID, integranteID).
		Delete(&proyectoIntegrante{})
	if resultado.Error != nil {
		return fmt.Errorf("desvinculando integrante: %w", resultado.Error)
	}
	if resultado.RowsAffected == 0 {
		return domain.ErrIntegranteNoAsociado
	}
	return nil
}

// ListarPorProyecto devuelve el equipo de un proyecto en orden de alta.
func (r *integranteRepository) ListarPorProyecto(ctx context.Context, proyectoID uint) ([]domain.Integrante, error) {
	integrantes := make([]domain.Integrante, 0)
	err := r.db.WithContext(ctx).
		Joins("JOIN proyecto_integrantes ON proyecto_integrantes.integrante_id = integrantes.id").
		Where("proyecto_integrantes.proyecto_id = ?", proyectoID).
		Order("integrantes.id ASC").
		Find(&integrantes).Error
	if err != nil {
		return nil, fmt.Errorf("listando integrantes del proyecto %d: %w", proyectoID, err)
	}
	return integrantes, nil
}
