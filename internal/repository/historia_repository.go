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

type historiaRepository struct {
	db *gorm.DB
}

// NuevoHistoriaRepository construye el repositorio de historias sobre GORM.
func NuevoHistoriaRepository(db *gorm.DB) *historiaRepository {
	return &historiaRepository{db: db}
}

var _ domain.HistoriaRepository = (*historiaRepository)(nil)

// Crear persiste una historia dentro del proyecto indicado. Cuando el
// consecutivo ya fue tomado por otra alta concurrente, el indice unico
// (proyecto_id, numero) aborta el INSERT y se traduce a ErrCodigoHistoriaDuplicado.
func (r *historiaRepository) Crear(ctx context.Context, h *domain.Historia) error {
	ahora := time.Now().UTC()
	h.ActualizadoEn = ahora
	if h.CreadoEn.IsZero() {
		h.CreadoEn = ahora
	}

	m := models.FromDomainHistoria(*h)
	if err := r.db.WithContext(ctx).Omit("Proyecto").Create(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrCodigoHistoriaDuplicado
		}
		return fmt.Errorf("persistiendo historia: %w", err)
	}

	h.ID = m.ID
	h.CreadoEn = m.CreadoEn
	h.ActualizadoEn = m.ActualizadoEn
	return nil
}

// SiguienteNumero devuelve el consecutivo que le corresponde a la proxima
// historia del proyecto. La primera historia es la numero 1.
func (r *historiaRepository) SiguienteNumero(ctx context.Context, proyectoID uint) (int, error) {
	var fila struct {
		Maximo *int
	}
	err := r.db.WithContext(ctx).
		Model(&models.HistoriaModel{}).
		Select("MAX(numero) AS maximo").
		Where("proyecto_id = ?", proyectoID).
		Scan(&fila).Error
	if err != nil {
		return 0, fmt.Errorf("calculando el consecutivo de la historia: %w", err)
	}
	if fila.Maximo == nil {
		return 1, nil
	}
	return *fila.Maximo + 1, nil
}

// ObtenerPorID busca una historia dentro de un proyecto concreto. Una historia
// de otro proyecto se comporta como inexistente.
func (r *historiaRepository) ObtenerPorID(ctx context.Context, proyectoID, id uint) (*domain.Historia, error) {
	var m models.HistoriaModel
	err := r.db.WithContext(ctx).
		Where("id = ? AND proyecto_id = ?", id, proyectoID).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrHistoriaNoEncontrada
		}
		return nil, fmt.Errorf("obteniendo historia %d: %w", id, err)
	}
	historia := models.ToDomainHistoria(m)
	return &historia, nil
}

// ListarPorProyecto devuelve el backlog del proyecto en orden de alta. Cuando
// no hay historias devuelve un slice vacio, nunca nil.
func (r *historiaRepository) ListarPorProyecto(ctx context.Context, proyectoID uint) ([]domain.Historia, error) {
	modelos := make([]models.HistoriaModel, 0)
	err := r.db.WithContext(ctx).
		Where("proyecto_id = ?", proyectoID).
		Order("numero ASC").
		Find(&modelos).Error
	if err != nil {
		return nil, fmt.Errorf("listando historias del proyecto %d: %w", proyectoID, err)
	}
	return models.ToDomainHistorias(modelos), nil
}

// Actualizar escribe titulo, descripcion, prioridad, estado y story points de
// la historia y refresca la marca de auditoria.
func (r *historiaRepository) Actualizar(ctx context.Context, h *domain.Historia) error {
	ahora := time.Now().UTC()
	h.ActualizadoEn = ahora

	resultado := r.db.WithContext(ctx).
		Model(&models.HistoriaModel{}).
		Where("id = ? AND proyecto_id = ?", h.ID, h.ProyectoID).
		Updates(map[string]any{
			"titulo":         h.Titulo,
			"descripcion":    h.Descripcion,
			"prioridad":      string(h.Prioridad),
			"estado":         string(h.Estado),
			"story_points":   h.StoryPoints,
			"actualizado_en": ahora,
		})
	if resultado.Error != nil {
		return fmt.Errorf("actualizando historia %d: %w", h.ID, resultado.Error)
	}
	if resultado.RowsAffected == 0 {
		return domain.ErrHistoriaNoEncontrada
	}
	return nil
}
