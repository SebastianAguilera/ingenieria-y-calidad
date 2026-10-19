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

type sprintRepository struct {
	db *gorm.DB
}

// NuevoSprintRepository construye el repositorio de Sprints sobre GORM.
func NuevoSprintRepository(db *gorm.DB) *sprintRepository {
	return &sprintRepository{db: db}
}

var _ domain.SprintRepository = (*sprintRepository)(nil)

// Crear persiste un Sprint dentro del proyecto indicado. Cuando el
// identificador ya fue tomado en el mismo proyecto, el indice unico
// (proyecto_id, identificador) aborta el INSERT y se traduce a
// ErrIdentificadorSprintDuplicado.
func (r *sprintRepository) Crear(ctx context.Context, s *domain.Sprint) error {
	ahora := time.Now().UTC()
	s.ActualizadoEn = ahora
	if s.CreadoEn.IsZero() {
		s.CreadoEn = ahora
	}
	if s.Estado == "" {
		s.Estado = domain.EstadoSprintPlanificado
	}

	m := models.FromDomainSprint(*s)
	if err := r.db.WithContext(ctx).Omit("Proyecto").Create(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrIdentificadorSprintDuplicado
		}
		return fmt.Errorf("persistiendo sprint: %w", err)
	}

	s.ID = m.ID
	s.CreadoEn = m.CreadoEn
	s.ActualizadoEn = m.ActualizadoEn
	return nil
}

// ObtenerPorID busca un Sprint dentro de un proyecto concreto. Un Sprint de
// otro proyecto se comporta como inexistente.
func (r *sprintRepository) ObtenerPorID(ctx context.Context, proyectoID, id uint) (*domain.Sprint, error) {
	var m models.SprintModel
	err := r.db.WithContext(ctx).
		Where("id = ? AND proyecto_id = ?", id, proyectoID).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSprintNoEncontrado
		}
		return nil, fmt.Errorf("obteniendo sprint %d: %w", id, err)
	}
	sprint := models.ToDomainSprint(m)
	return &sprint, nil
}

// ListarPorProyecto devuelve los Sprints del proyecto ordenados por
// fecha_inicio ASC y desempate por id ASC. Cuando no hay Sprints devuelve un
// slice vacio, nunca nil.
func (r *sprintRepository) ListarPorProyecto(ctx context.Context, proyectoID uint) ([]domain.Sprint, error) {
	modelos := make([]models.SprintModel, 0)
	err := r.db.WithContext(ctx).
		Where("proyecto_id = ?", proyectoID).
		Order("fecha_inicio ASC, id ASC").
		Find(&modelos).Error
	if err != nil {
		return nil, fmt.Errorf("listando sprints del proyecto %d: %w", proyectoID, err)
	}
	return models.ToDomainSprints(modelos), nil
}

// ActualizarEstado escribe el estado del Sprint y refresca la marca de
// auditoria.
func (r *sprintRepository) ActualizarEstado(ctx context.Context, s *domain.Sprint) error {
	ahora := time.Now().UTC()
	s.ActualizadoEn = ahora

	resultado := r.db.WithContext(ctx).
		Model(&models.SprintModel{}).
		Where("id = ? AND proyecto_id = ?", s.ID, s.ProyectoID).
		Updates(map[string]any{
			"estado":         string(s.Estado),
			"actualizado_en": ahora,
		})
	if resultado.Error != nil {
		return fmt.Errorf("actualizando estado del sprint %d: %w", s.ID, resultado.Error)
	}
	if resultado.RowsAffected == 0 {
		return domain.ErrSprintNoEncontrado
	}
	return nil
}

// ObtenerActivo devuelve el Sprint en estado Activo del proyecto, o
// (nil, nil) si el proyecto no tiene ningun Sprint activo.
func (r *sprintRepository) ObtenerActivo(ctx context.Context, proyectoID uint) (*domain.Sprint, error) {
	var m models.SprintModel
	err := r.db.WithContext(ctx).
		Where("proyecto_id = ? AND estado = ?", proyectoID, string(domain.EstadoSprintActivo)).
		Order("id ASC").
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("obteniendo el sprint activo del proyecto %d: %w", proyectoID, err)
	}
	sprint := models.ToDomainSprint(m)
	return &sprint, nil
}

// HaySolapamiento responde si existe un Sprint del proyecto cuyo rango se
// interseca estrictamente con [inicio, fin]. Los Sprints contiguos (la fin de
// uno es igual a la inicio del siguiente) no se consideran solapados.
func (r *sprintRepository) HaySolapamiento(ctx context.Context, proyectoID uint, inicio, fin time.Time) (bool, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&models.SprintModel{}).
		Where("proyecto_id = ? AND fecha_inicio < ? AND fecha_fin > ?", proyectoID, fin, inicio).
		Count(&total).Error
	if err != nil {
		return false, fmt.Errorf("verificando el solapamiento de fechas del proyecto %d: %w", proyectoID, err)
	}
	return total > 0, nil
}
