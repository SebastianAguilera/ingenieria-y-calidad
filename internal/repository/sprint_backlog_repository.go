package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"ingenieria-y-calidad/internal/domain"
	"ingenieria-y-calidad/internal/repository/models"
)

type sprintBacklogRepository struct {
	db *gorm.DB
}

// NuevoSprintBacklogRepository construye el repositorio de Sprint Backlog.
func NuevoSprintBacklogRepository(db *gorm.DB) *sprintBacklogRepository {
	return &sprintBacklogRepository{db: db}
}

var _ domain.SprintBacklogRepository = (*sprintBacklogRepository)(nil)

func (r *sprintBacklogRepository) Asignar(ctx context.Context, item *domain.SprintBacklogItem) error {
	ahora := time.Now().UTC()
	item.CreadoEn = ahora
	item.ActualizadoEn = ahora
	if item.FechaAsignacion.IsZero() {
		item.FechaAsignacion = ahora
	}

	m := models.SprintBacklogModel{
		ProyectoID:      item.ProyectoID,
		SprintID:        item.SprintID,
		HistoriaID:      item.HistoriaID,
		Orden:           item.Orden,
		FechaAsignacion: item.FechaAsignacion,
		CreadoEn:        item.CreadoEn,
		ActualizadoEn:   item.ActualizadoEn,
	}

	if err := r.db.WithContext(ctx).Omit("Proyecto", "Sprint", "Historia").Create(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrHistoriaYaAsignada
		}
		return fmt.Errorf("asignando historia al sprint backlog: %w", err)
	}

	item.ID = m.ID
	item.CreadoEn = m.CreadoEn
	item.ActualizadoEn = m.ActualizadoEn
	return nil
}

func (r *sprintBacklogRepository) Desasignar(ctx context.Context, proyectoID, sprintID, historiaID uint) error {
	resultado := r.db.WithContext(ctx).
		Where("proyecto_id = ? AND sprint_id = ? AND historia_id = ?", proyectoID, sprintID, historiaID).
		Delete(&models.SprintBacklogModel{})
	if resultado.Error != nil {
		return fmt.Errorf("desasignando historia del sprint backlog: %w", resultado.Error)
	}
	if resultado.RowsAffected == 0 {
		return domain.ErrHistoriaNoPerteneceASprint
	}
	return nil
}

func (r *sprintBacklogRepository) ExisteAsignacionHistoria(ctx context.Context, historiaID uint) (bool, *domain.SprintBacklogItem, error) {
	var m models.SprintBacklogModel
	err := r.db.WithContext(ctx).Where("historia_id = ?", historiaID).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("verificando asignación de historia: %w", err)
	}
	item := toDomainSprintBacklogItem(m)
	return true, &item, nil
}

func (r *sprintBacklogRepository) ObtenerPorHistoria(ctx context.Context, historiaID uint) (*domain.SprintBacklogItem, error) {
	var m models.SprintBacklogModel
	err := r.db.WithContext(ctx).Where("historia_id = ?", historiaID).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("obteniendo asignación de historia: %w", err)
	}
	item := toDomainSprintBacklogItem(m)
	return &item, nil
}

func (r *sprintBacklogRepository) ObtenerPorSprint(ctx context.Context, proyectoID, sprintID uint) ([]domain.SprintBacklogItem, error) {
	var modelos []models.SprintBacklogModel
	err := r.db.WithContext(ctx).
		Where("proyecto_id = ? AND sprint_id = ?", proyectoID, sprintID).
		Order("orden ASC, id ASC").
		Find(&modelos).Error
	if err != nil {
		return nil, fmt.Errorf("listando sprint backlog: %w", err)
	}
	items := make([]domain.SprintBacklogItem, 0, len(modelos))
	for _, m := range modelos {
		items = append(items, toDomainSprintBacklogItem(m))
	}
	return items, nil
}

func (r *sprintBacklogRepository) ObtenerMaxOrden(ctx context.Context, proyectoID, sprintID uint) (int, error) {
	var fila struct {
		Maximo *int
	}
	err := r.db.WithContext(ctx).
		Model(&models.SprintBacklogModel{}).
		Select("MAX(orden) AS maximo").
		Where("proyecto_id = ? AND sprint_id = ?", proyectoID, sprintID).
		Scan(&fila).Error
	if err != nil {
		return 0, fmt.Errorf("obteniendo orden máximo del sprint backlog: %w", err)
	}
	if fila.Maximo == nil {
		return 0, nil
	}
	return *fila.Maximo, nil
}

func toDomainSprintBacklogItem(m models.SprintBacklogModel) domain.SprintBacklogItem {
	return domain.SprintBacklogItem{
		ID:              m.ID,
		ProyectoID:      m.ProyectoID,
		SprintID:        m.SprintID,
		HistoriaID:      m.HistoriaID,
		Orden:           m.Orden,
		FechaAsignacion: m.FechaAsignacion,
		CreadoEn:        m.CreadoEn,
		ActualizadoEn:   m.ActualizadoEn,
	}
}
