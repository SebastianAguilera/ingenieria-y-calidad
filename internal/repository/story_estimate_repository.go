package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"ingenieria-y-calidad/internal/domain"
	"ingenieria-y-calidad/internal/repository/models"
)

type storyEstimateRepository struct {
	db *gorm.DB
}

// NuevoStoryEstimateRepository construye el repositorio de estimaciones
// acordadas (tabla story_estimates) sobre GORM.
func NuevoStoryEstimateRepository(db *gorm.DB) *storyEstimateRepository {
	return &storyEstimateRepository{db: db}
}

var _ domain.StoryEstimateRepository = (*storyEstimateRepository)(nil)

// Crear inserta la estimación. Una historia, sesión o ronda inexistente viola
// la clave foránea y devuelve error, lo que revierte la UnidadDeTrabajo.
func (r *storyEstimateRepository) Crear(ctx context.Context, e *domain.StoryEstimate) error {
	modelo := models.FromDomainStoryEstimate(*e)
	if err := r.db.WithContext(ctx).Omit("Historia", "Sesion", "Ronda").Create(&modelo).Error; err != nil {
		return fmt.Errorf("creando el story_estimate de la historia %d: %w", e.HistoriaID, err)
	}
	e.ID = modelo.ID
	e.EstimadoEn = modelo.EstimadoEn
	return nil
}

// ListarPorSesiones devuelve las estimaciones de las sesiones indicadas en una
// sola consulta, ordenadas por identificador.
func (r *storyEstimateRepository) ListarPorSesiones(ctx context.Context, sesionIDs []uint) ([]domain.StoryEstimate, error) {
	if len(sesionIDs) == 0 {
		return []domain.StoryEstimate{}, nil
	}

	var filas []models.StoryEstimateModel
	if err := r.db.WithContext(ctx).
		Where("sesion_id IN ?", sesionIDs).
		Order("id ASC").
		Find(&filas).Error; err != nil {
		return nil, fmt.Errorf("listando los story_estimate de las sesiones: %w", err)
	}

	estimaciones := make([]domain.StoryEstimate, len(filas))
	for idx, fila := range filas {
		estimaciones[idx] = models.ToDomainStoryEstimate(fila)
	}
	return estimaciones, nil
}
