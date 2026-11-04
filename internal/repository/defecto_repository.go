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

type defectoRepository struct {
	db *gorm.DB
}

// NuevoDefectoRepository construye el repositorio de defectos sobre GORM.
func NuevoDefectoRepository(db *gorm.DB) *defectoRepository {
	return &defectoRepository{db: db}
}

var _ domain.DefectoRepository = (*defectoRepository)(nil)

// Crear persiste un defecto. El estado inicial y la severidad por defecto se
// resuelven en el dominio; aquí solo se completa la auditoría.
func (r *defectoRepository) Crear(ctx context.Context, d *domain.Defecto) error {
	ahora := time.Now().UTC()
	if d.CreadoEn.IsZero() {
		d.CreadoEn = ahora
	}
	if d.Estado == "" {
		d.Estado = domain.EstadoDefectoAbierto
	}
	if d.Severidad == "" {
		d.Severidad = domain.SeveridadMinor
	}

	m := models.FromDomainDefecto(*d)
	if err := r.db.WithContext(ctx).
		Omit("Proyecto", "Historia", "SprintDeteccion", "SprintResolucion", "Reportado").
		Create(&m).Error; err != nil {
		return fmt.Errorf("persistiendo defecto: %w", err)
	}

	d.ID = m.ID
	d.CreadoEn = m.CreadoEn
	d.ResueltoEn = m.ResueltoEn
	return nil
}

// ObtenerPorID busca un defecto dentro de un proyecto concreto. Un defecto de
// otro proyecto se comporta como inexistente.
func (r *defectoRepository) ObtenerPorID(ctx context.Context, proyectoID, id uint) (*domain.Defecto, error) {
	var m models.DefectoModel
	err := r.db.WithContext(ctx).
		Where("id = ? AND proyecto_id = ?", id, proyectoID).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrDefectoNoEncontrado
		}
		return nil, fmt.Errorf("obteniendo defecto %d: %w", id, err)
	}
	defecto := models.ToDomainDefecto(m)
	return &defecto, nil
}

// Actualizar escribe el estado, el Sprint de resolución y la fecha de
// resolución del defecto.
func (r *defectoRepository) Actualizar(ctx context.Context, d *domain.Defecto) error {
	resultado := r.db.WithContext(ctx).
		Model(&models.DefectoModel{}).
		Where("id = ? AND proyecto_id = ?", d.ID, d.ProyectoID).
		Updates(map[string]any{
			"estado":               string(d.Estado),
			"sprint_resolucion_id": d.SprintResolucionID,
			"resuelto_en":          d.ResueltoEn,
		})
	if resultado.Error != nil {
		return fmt.Errorf("actualizando defecto %d: %w", d.ID, resultado.Error)
	}
	if resultado.RowsAffected == 0 {
		return domain.ErrDefectoNoEncontrado
	}
	return nil
}

// ListarPorProyecto devuelve los defectos del proyecto en orden de alta.
// Cuando no hay defectos devuelve un slice vacío, nunca nil.
func (r *defectoRepository) ListarPorProyecto(ctx context.Context, proyectoID uint) ([]domain.Defecto, error) {
	modelos := make([]models.DefectoModel, 0)
	err := r.db.WithContext(ctx).
		Where("proyecto_id = ?", proyectoID).
		Order("id ASC").
		Find(&modelos).Error
	if err != nil {
		return nil, fmt.Errorf("listando defectos del proyecto %d: %w", proyectoID, err)
	}
	return models.ToDomainDefectos(modelos), nil
}
