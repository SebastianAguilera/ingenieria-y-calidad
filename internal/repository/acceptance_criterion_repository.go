package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"ingenieria-y-calidad/internal/domain"
	"ingenieria-y-calidad/internal/repository/models"
)

type acceptanceCriterionRepository struct {
	db *gorm.DB
}

func NuevoAcceptanceCriterionRepository(db *gorm.DB) *acceptanceCriterionRepository {
	return &acceptanceCriterionRepository{db: db}
}

var _ domain.AcceptanceCriterionRepository = (*acceptanceCriterionRepository)(nil)

func (r *acceptanceCriterionRepository) Crear(ctx context.Context, c *domain.AcceptanceCriterion) error {
	m := models.FromDomainAcceptanceCriterion(*c)
	if err := r.db.WithContext(ctx).Omit("Historia").Create(&m).Error; err != nil {
		return fmt.Errorf("persistiendo criterio de aceptación: %w", err)
	}
	c.ID = m.ID
	return nil
}

func (r *acceptanceCriterionRepository) ListarPorHistoria(ctx context.Context, historiaID uint) ([]domain.AcceptanceCriterion, error) {
	var modelos []models.AcceptanceCriterionModel
	err := r.db.WithContext(ctx).
		Where("backlog_item_id = ?", historiaID).
		Order("position ASC").
		Find(&modelos).Error
	if err != nil {
		return nil, fmt.Errorf("listando criterios de aceptación: %w", err)
	}
	return models.ToDomainAcceptanceCriteria(modelos), nil
}

func (r *acceptanceCriterionRepository) ObtenerPorID(ctx context.Context, historiaID, id uint) (*domain.AcceptanceCriterion, error) {
	var m models.AcceptanceCriterionModel
	err := r.db.WithContext(ctx).
		Where("id = ? AND backlog_item_id = ?", id, historiaID).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrCriterioNoEncontrado
		}
		return nil, fmt.Errorf("obteniendo criterio de aceptación %d: %w", id, err)
	}
	c := models.ToDomainAcceptanceCriterion(m)
	return &c, nil
}

func (r *acceptanceCriterionRepository) Actualizar(ctx context.Context, c *domain.AcceptanceCriterion) error {
	resultado := r.db.WithContext(ctx).
		Model(&models.AcceptanceCriterionModel{}).
		Where("id = ? AND backlog_item_id = ?", c.ID, c.HistoriaID).
		Updates(map[string]any{
			"description": c.Descripcion,
			"is_met":      c.IsMet,
		})
	if resultado.Error != nil {
		return fmt.Errorf("actualizando criterio de aceptación %d: %w", c.ID, resultado.Error)
	}
	if resultado.RowsAffected == 0 {
		return domain.ErrCriterioNoEncontrado
	}
	return nil
}

func (r *acceptanceCriterionRepository) Eliminar(ctx context.Context, historiaID, id uint) error {
	resultado := r.db.WithContext(ctx).
		Where("id = ? AND backlog_item_id = ?", id, historiaID).
		Delete(&models.AcceptanceCriterionModel{})
	if resultado.Error != nil {
		return fmt.Errorf("eliminando criterio de aceptación %d: %w", id, resultado.Error)
	}
	if resultado.RowsAffected == 0 {
		return domain.ErrCriterioNoEncontrado
	}
	return nil
}

func (r *acceptanceCriterionRepository) SiguientePosicion(ctx context.Context, historiaID uint) (int, error) {
	var fila struct {
		Maximo *int
	}
	err := r.db.WithContext(ctx).
		Model(&models.AcceptanceCriterionModel{}).
		Select("MAX(position) AS maximo").
		Where("backlog_item_id = ?", historiaID).
		Scan(&fila).Error
	if err != nil {
		return 0, fmt.Errorf("calculando la siguiente posición del criterio: %w", err)
	}
	if fila.Maximo == nil {
		return 1, nil
	}
	return *fila.Maximo + 1, nil
}
