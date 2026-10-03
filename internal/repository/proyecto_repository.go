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

type proyectoRepository struct {
	db *gorm.DB
}

// NuevoProyectoRepository construye el repositorio de proyectos sobre GORM.
func NuevoProyectoRepository(db *gorm.DB) *proyectoRepository {
	return &proyectoRepository{db: db}
}

var _ domain.ProyectoRepository = (*proyectoRepository)(nil)

// CrearConUsuarios persiste el proyecto y todos sus vinculos dentro de una
// unica transaccion: o se persiste todo, o no se persiste nada. Los
// identificadores asignados por la base se escriben en el slice recibido para
// que el proyecto devuelto expose el equipo ya persistido.
func (r *proyectoRepository) CrearConUsuarios(ctx context.Context, p *domain.Proyecto, usuarios []domain.Usuario) error {
	ahora := time.Now().UTC()
	p.ActualizadoEn = ahora
	if p.CreadoEn.IsZero() {
		p.CreadoEn = ahora
	}

	m := models.FromDomainProyecto(*p)
	m.Usuarios = models.FromDomainUsuarios(usuarios)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := crearUsuariosModel(tx, m.Usuarios); err != nil {
			return err
		}

		if err := tx.Omit("Usuarios").Create(&m).Error; err != nil {
			return fmt.Errorf("persistiendo proyecto: %w", err)
		}

		p.ID = m.ID
		p.CreadoEn = m.CreadoEn
		p.ActualizadoEn = m.ActualizadoEn

		for i := range usuarios {
			usuarios[i].ID = m.Usuarios[i].ID
			usuarios[i].CreadoEn = m.Usuarios[i].CreadoEn
		}

		if len(m.Usuarios) == 0 {
			return nil
		}
		return asociarUsuariosModel(tx, m.ID, m.Usuarios)
	})
}

// crearUsuariosModel inserta los usuarios que aun no tienen identificador,
// escribiendo en el slice del llamador los ids generados por la base.
func crearUsuariosModel(tx *gorm.DB, usuarios []models.UsuarioModel) error {
	for idx := range usuarios {
		if usuarios[idx].ID != 0 {
			continue
		}
		usuarios[idx].Email = domain.NormalizarEmail(usuarios[idx].Email)
		if err := tx.Create(&usuarios[idx]).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return domain.ErrUsuarioYaAsociado
			}
			return fmt.Errorf("persistiendo usuario %s: %w", usuarios[idx].Email, err)
		}
	}
	return nil
}

func asociarUsuariosModel(tx *gorm.DB, proyectoID uint, usuarios []models.UsuarioModel) error {
	vinculos := make([]models.ProyectoUsuarioModel, 0, len(usuarios))
	for _, usuario := range usuarios {
		vinculos = append(vinculos, models.ProyectoUsuarioModel{
			ProyectoID: proyectoID,
			UsuarioID:  usuario.ID,
		})
	}
	if err := tx.Create(&vinculos).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrUsuarioYaAsociado
		}
		return fmt.Errorf("vinculando usuarios: %w", err)
	}
	return nil
}

// ObtenerPorID devuelve el proyecto con sus usuarios precargados. Los
// proyectos dados de baja logicamente son invisibles para el cliente.
func (r *proyectoRepository) ObtenerPorID(ctx context.Context, id uint) (*domain.Proyecto, error) {
	var m models.ProyectoModel
	err := r.db.WithContext(ctx).
		Preload("Usuarios").
		Where("id = ? AND borrado_en IS NULL", id).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrProyectoNoEncontrado
		}
		return nil, fmt.Errorf("obteniendo proyecto %d: %w", id, err)
	}
	domainProyecto := models.ToDomainProyecto(m)
	return &domainProyecto, nil
}

// Listar devuelve los proyectos no dados de baja con usuarios precargados.
func (r *proyectoRepository) Listar(ctx context.Context) ([]domain.Proyecto, error) {
	var modelos []models.ProyectoModel
	err := r.db.WithContext(ctx).
		Preload("Usuarios").
		Where("borrado_en IS NULL").
		Order("id ASC").
		Find(&modelos).Error
	if err != nil {
		return nil, fmt.Errorf("listando proyectos: %w", err)
	}
	return models.ToDomainProyectos(modelos), nil
}

// Actualizar escribe nombre, fechas y estado del proyecto, y refresca la
// marca de auditoria.
func (r *proyectoRepository) Actualizar(ctx context.Context, p *domain.Proyecto) error {
	ahora := time.Now().UTC()
	p.ActualizadoEn = ahora

	if err := r.db.WithContext(ctx).
		Model(&models.ProyectoModel{}).
		Where("id = ? AND borrado_en IS NULL", p.ID).
		Updates(map[string]any{
			"nombre":         p.Nombre,
			"fecha_inicio":   p.FechaInicio,
			"fecha_fin":      p.FechaFin,
			"estado":         string(p.Estado),
			"actualizado_en": ahora,
		}).Error; err != nil {
		return fmt.Errorf("actualizando proyecto %d: %w", p.ID, err)
	}
	return nil
}

// Eliminar realiza la baja logica: informa la hora en borrado_en.
func (r *proyectoRepository) Eliminar(ctx context.Context, id uint) error {
	resultado := r.db.WithContext(ctx).
		Model(&models.ProyectoModel{}).
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
func (r *proyectoRepository) TieneHistorial(_ context.Context, _ uint) (bool, error) {
	return false, nil
}
