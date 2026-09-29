package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"ingenieria-y-calidad/internal/domain"
)

type proyectoRepository struct {
	db *gorm.DB
}

// NuevoProyectoRepository construye el repositorio de proyectos sobre GORM.
func NuevoProyectoRepository(db *gorm.DB) *proyectoRepository {
	return &proyectoRepository{db: db}
}

var _ domain.ProyectoRepository = (*proyectoRepository)(nil)

// CrearConIntegrantes persiste el proyecto y todos sus vinculos dentro de una
// unica transaccion: o se persiste todo, o no se persiste nada. Los
// identificadores asignados por la base se escriben en el slice recibido para
// que el proyecto devuelto expose el equipo ya persistido.
func (r *proyectoRepository) CrearConIntegrantes(ctx context.Context, p *domain.Proyecto, integrantes []domain.Integrante) error {
	ahora := time.Now().UTC()
	p.ActualizadoEn = ahora
	if p.CreadoEn.IsZero() {
		p.CreadoEn = ahora
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := crearIntegrantes(tx, integrantes); err != nil {
			return err
		}

		if err := tx.Omit("Integrantes").Create(p).Error; err != nil {
			return fmt.Errorf("persistiendo proyecto: %w", err)
		}

		if len(integrantes) == 0 {
			return nil
		}
		return asociarIntegrantes(tx, p.ID, integrantes)
	})
}

// crearIntegrantes inserta los integrantes que aun no tienen identificador,
// escribiendo en el slice del llamador los ids generados por la base.
func crearIntegrantes(tx *gorm.DB, integrantes []domain.Integrante) error {
	for idx := range integrantes {
		if integrantes[idx].ID != 0 {
			continue
		}
		if err := tx.Create(&integrantes[idx]).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return domain.ErrIntegranteYaAsociado
			}
			return fmt.Errorf("persistiendo integrante %s: %w", integrantes[idx].Email, err)
		}
	}
	return nil
}

func asociarIntegrantes(tx *gorm.DB, proyectoID uint, integrantes []domain.Integrante) error {
	vinculos := make([]proyectoIntegrante, 0, len(integrantes))
	for _, integrante := range integrantes {
		vinculos = append(vinculos, proyectoIntegrante{
			ProyectoID:   proyectoID,
			IntegranteID: integrante.ID,
		})
	}
	if err := tx.Create(&vinculos).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrIntegranteYaAsociado
		}
		return fmt.Errorf("vinculando integrantes: %w", err)
	}
	return nil
}

// ObtenerPorID devuelve el proyecto con sus integrantes precargados. Los
// proyectos dados de baja logicamente son invisibles para el cliente.
func (r *proyectoRepository) ObtenerPorID(ctx context.Context, id uint) (*domain.Proyecto, error) {
	var proyecto domain.Proyecto
	err := r.db.WithContext(ctx).
		Preload("Integrantes").
		Where("id = ? AND borrado_en IS NULL", id).
		First(&proyecto).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrProyectoNoEncontrado
		}
		return nil, fmt.Errorf("obteniendo proyecto %d: %w", id, err)
	}
	return &proyecto, nil
}

// Listar devuelve los proyectos no dados de baja con integrantes precargados.
func (r *proyectoRepository) Listar(ctx context.Context) ([]domain.Proyecto, error) {
	proyectos := make([]domain.Proyecto, 0)
	err := r.db.WithContext(ctx).
		Preload("Integrantes").
		Where("borrado_en IS NULL").
		Order("id ASC").
		Find(&proyectos).Error
	if err != nil {
		return nil, fmt.Errorf("listando proyectos: %w", err)
	}
	return proyectos, nil
}

// Actualizar escribe nombre, fechas y estado del proyecto, y refresca la
// marca de auditoria.
func (r *proyectoRepository) Actualizar(ctx context.Context, p *domain.Proyecto) error {
	ahora := time.Now().UTC()
	p.ActualizadoEn = ahora

	if err := r.db.WithContext(ctx).
		Model(&domain.Proyecto{}).
		Where("id = ? AND borrado_en IS NULL", p.ID).
		Updates(map[string]any{
			"nombre":         p.Nombre,
			"fecha_inicio":   p.FechaInicio,
			"fecha_fin":      p.FechaFin,
			"estado":         p.Estado,
			"actualizado_en": ahora,
		}).Error; err != nil {
		return fmt.Errorf("actualizando proyecto %d: %w", p.ID, err)
	}
	return nil
}

// Eliminar realiza la baja logica: informa la hora en borrado_en.
func (r *proyectoRepository) Eliminar(ctx context.Context, id uint) error {
	resultado := r.db.WithContext(ctx).
		Model(&domain.Proyecto{}).
		Where("id = ? AND borrado_en IS NULL", id).
		Update("borrado_en", gorm.Expr("now()"))
	if resultado.Error != nil {
		return fmt.Errorf("eliminando proyecto %d: %w", id, resultado.Error)
	}
	if resultado.RowsAffected == 0 {
		return domain.ErrProyectoNoEncontrado
	}
	return nil
}

// TieneHistorial informa si el proyecto ya tiene entidades dependientes.
// En US-01 todavia no existen Sprints, historias ni worklogs, por lo que la
// implementacion actual devuelve siempre false. La interfaz queda declarada
// para que las historias siguientes completen la regla sin tocar el service.
func (r *proyectoRepository) TieneHistorial(_ context.Context, _ uint) (bool, error) {
	return false, nil
}
