package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"ingenieria-y-calidad/internal/domain"
	"ingenieria-y-calidad/internal/repository/models"
)

type usuarioRepository struct {
	db *gorm.DB
}

// NuevoUsuarioRepository construye el repositorio de usuarios sobre GORM.
func NuevoUsuarioRepository(db *gorm.DB) *usuarioRepository {
	return &usuarioRepository{db: db}
}

var _ domain.UsuarioRepository = (*usuarioRepository)(nil)

// ObtenerPorEmail busca una persona por su clave natural normalizada.
func (r *usuarioRepository) ObtenerPorEmail(ctx context.Context, email string) (*domain.Usuario, error) {
	var m models.UsuarioModel
	err := r.db.WithContext(ctx).
		Where("email = ?", domain.NormalizarEmail(email)).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUsuarioNoEncontrado
		}
		return nil, fmt.Errorf("buscando usuario por email: %w", err)
	}
	domainUsuario := models.ToDomainUsuario(m)
	return &domainUsuario, nil
}

// Crear inserta un usuario. El email es unico en todo el sistema.
func (r *usuarioRepository) Crear(ctx context.Context, u *domain.Usuario) error {
	u.Email = domain.NormalizarEmail(u.Email)
	m := models.FromDomainUsuario(*u)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrUsuarioYaAsociado
		}
		return fmt.Errorf("creando usuario: %w", err)
	}
	u.ID = m.ID
	u.CreadoEn = m.CreadoEn
	return nil
}

// ActualizarNombre refresca el nombre de un usuario preexistente.
func (r *usuarioRepository) ActualizarNombre(ctx context.Context, id uint, nombre string) error {
	if err := r.db.WithContext(ctx).
		Model(&models.UsuarioModel{}).
		Where("id = ?", id).
		Update("nombre", strings.TrimSpace(nombre)).Error; err != nil {
		return fmt.Errorf("actualizando nombre del usuario %d: %w", id, err)
	}
	return nil
}

// EstaAsociado informa si la persona ya forma parte del equipo del proyecto.
func (r *usuarioRepository) EstaAsociado(ctx context.Context, proyectoID, usuarioID uint) (bool, error) {
	var vinculo models.ProyectoUsuarioModel
	err := r.db.WithContext(ctx).
		Where("proyecto_id = ? AND usuario_id = ?", proyectoID, usuarioID).
		First(&vinculo).Error
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return false, nil
	default:
		return false, fmt.Errorf("verificando asociacion del usuario: %w", err)
	}
}

// Vincular agrega el vinculo entre proyecto e usuario.
func (r *usuarioRepository) Vincular(ctx context.Context, proyectoID, usuarioID uint) error {
	vinculo := models.ProyectoUsuarioModel{ProyectoID: proyectoID, UsuarioID: usuarioID}
	if err := r.db.WithContext(ctx).Create(&vinculo).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrUsuarioYaAsociado
		}
		return fmt.Errorf("vinculando usuario: %w", err)
	}
	return nil
}

// Desvincular quita el vinculo entre proyecto e usuario.
func (r *usuarioRepository) Desvincular(ctx context.Context, proyectoID, usuarioID uint) error {
	resultado := r.db.WithContext(ctx).
		Where("proyecto_id = ? AND usuario_id = ?", proyectoID, usuarioID).
		Delete(&models.ProyectoUsuarioModel{})
	if resultado.Error != nil {
		return fmt.Errorf("desvinculando usuario: %w", resultado.Error)
	}
	if resultado.RowsAffected == 0 {
		return domain.ErrUsuarioNoAsociado
	}
	return nil
}

// ListarPorProyecto devuelve el equipo de un proyecto en orden de alta.
func (r *usuarioRepository) ListarPorProyecto(ctx context.Context, proyectoID uint) ([]domain.Usuario, error) {
	var modelos []models.UsuarioModel
	err := r.db.WithContext(ctx).
		Joins("JOIN proyecto_usuarios ON proyecto_usuarios.usuario_id = usuarios.id").
		Where("proyecto_usuarios.proyecto_id = ?", proyectoID).
		Order("usuarios.id ASC").
		Find(&modelos).Error
	if err != nil {
		return nil, fmt.Errorf("listando usuarios del proyecto %d: %w", proyectoID, err)
	}
	return models.ToDomainUsuarios(modelos), nil
}
