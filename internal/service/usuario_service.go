package service

import (
	"context"
	"fmt"
	"strings"

	"ingenieria-y-calidad/internal/domain"
)

// UsuarioService implementa las reglas de negocio para la gestión independiente de usuarios.
type UsuarioService struct {
	usuarios domain.UsuarioRepository
}

// NewUsuarioService construye el servicio de usuarios con inyección de dependencias.
func NewUsuarioService(usuarios domain.UsuarioRepository) *UsuarioService {
	return &UsuarioService{usuarios: usuarios}
}

// Crear registra un usuario independiente.
func (s *UsuarioService) Crear(ctx context.Context, entrada domain.UsuarioInput) (*domain.Usuario, error) {
	if err := validarUsuario(entrada); err != nil {
		return nil, err
	}

	usuario := &domain.Usuario{
		Nombre: strings.TrimSpace(entrada.Nombre),
		Email:  domain.NormalizarEmail(entrada.Email),
	}

	if err := s.usuarios.Crear(ctx, usuario); err != nil {
		return nil, fmt.Errorf("creando usuario: %w", err)
	}

	return usuario, nil
}

// Listar devuelve todos los usuarios registrados.
func (s *UsuarioService) Listar(ctx context.Context) ([]domain.Usuario, error) {
	usuarios, err := s.usuarios.Listar(ctx)
	if err != nil {
		return nil, fmt.Errorf("listando usuarios: %w", err)
	}
	if usuarios == nil {
		return []domain.Usuario{}, nil
	}
	return usuarios, nil
}

// ObtenerPorID devuelve un usuario por su identificador.
func (s *UsuarioService) ObtenerPorID(ctx context.Context, id uint) (*domain.Usuario, error) {
	usuario, err := s.usuarios.ObtenerPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obteniendo usuario: %w", err)
	}
	return usuario, nil
}
