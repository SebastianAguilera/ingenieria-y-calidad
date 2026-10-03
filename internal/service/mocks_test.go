package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"ingenieria-y-calidad/internal/domain"
)

type MockProyectoRepository struct {
	mock.Mock
}

func nuevoMockProyectoRepository(t *testing.T) *MockProyectoRepository {
	t.Helper()
	m := &MockProyectoRepository{}
	m.Mock.Test(t)
	return m
}

func errorDeRetorno(ret mock.Arguments) error {
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Error(0)
}

func (m *MockProyectoRepository) CrearConUsuarios(ctx context.Context, p *domain.Proyecto, usuarios []domain.Usuario) error {
	return errorDeRetorno(m.Called(ctx, p, usuarios))
}

func (m *MockProyectoRepository) ObtenerPorID(ctx context.Context, id uint) (*domain.Proyecto, error) {
	ret := m.Called(ctx, id)
	var proyecto *domain.Proyecto
	if ret.Get(0) != nil {
		proyecto = ret.Get(0).(*domain.Proyecto)
	}
	return proyecto, ret.Error(1)
}

func (m *MockProyectoRepository) Listar(ctx context.Context) ([]domain.Proyecto, error) {
	ret := m.Called(ctx)
	var proyectos []domain.Proyecto
	if ret.Get(0) != nil {
		proyectos = ret.Get(0).([]domain.Proyecto)
	}
	return proyectos, ret.Error(1)
}

func (m *MockProyectoRepository) Actualizar(ctx context.Context, p *domain.Proyecto) error {
	return errorDeRetorno(m.Called(ctx, p))
}

func (m *MockProyectoRepository) Eliminar(ctx context.Context, id uint) error {
	return errorDeRetorno(m.Called(ctx, id))
}

func (m *MockProyectoRepository) TieneHistorial(ctx context.Context, id uint) (bool, error) {
	ret := m.Called(ctx, id)
	return ret.Bool(0), ret.Error(1)
}

type MockUsuarioRepository struct {
	mock.Mock
}

func nuevoMockUsuarioRepository(t *testing.T) *MockUsuarioRepository {
	t.Helper()
	m := &MockUsuarioRepository{}
	m.Mock.Test(t)
	return m
}

func (m *MockUsuarioRepository) ObtenerPorEmail(ctx context.Context, email string) (*domain.Usuario, error) {
	ret := m.Called(ctx, email)
	var usuario *domain.Usuario
	if ret.Get(0) != nil {
		usuario = ret.Get(0).(*domain.Usuario)
	}
	return usuario, ret.Error(1)
}

func (m *MockUsuarioRepository) Crear(ctx context.Context, u *domain.Usuario) error {
	return errorDeRetorno(m.Called(ctx, u))
}

func (m *MockUsuarioRepository) ActualizarNombre(ctx context.Context, id uint, nombre string) error {
	return errorDeRetorno(m.Called(ctx, id, nombre))
}

func (m *MockUsuarioRepository) EstaAsociado(ctx context.Context, proyectoID, usuarioID uint) (bool, error) {
	ret := m.Called(ctx, proyectoID, usuarioID)
	return ret.Bool(0), ret.Error(1)
}

func (m *MockUsuarioRepository) Vincular(ctx context.Context, proyectoID, usuarioID uint) error {
	return errorDeRetorno(m.Called(ctx, proyectoID, usuarioID))
}

func (m *MockUsuarioRepository) Desvincular(ctx context.Context, proyectoID, usuarioID uint) error {
	return errorDeRetorno(m.Called(ctx, proyectoID, usuarioID))
}

func (m *MockUsuarioRepository) ListarPorProyecto(ctx context.Context, proyectoID uint) ([]domain.Usuario, error) {
	ret := m.Called(ctx, proyectoID)
	var usuarios []domain.Usuario
	if ret.Get(0) != nil {
		usuarios = ret.Get(0).([]domain.Usuario)
	}
	return usuarios, ret.Error(1)
}

func proyectoDePrueba(id uint, nombre string, inicio time.Time, fin *time.Time, estado domain.EstadoProyecto) *domain.Proyecto {
	return &domain.Proyecto{
		ID:          id,
		Nombre:      nombre,
		FechaInicio: inicio,
		FechaFin:    fin,
		Estado:      estado,
		Usuarios:    []domain.Usuario{},
	}
}

// asignarIDProyecto emula la generacion de clave primaria del repositorio.
func asignarIDProyecto(id uint) func(mock.Arguments) {
	return func(args mock.Arguments) {
		args.Get(1).(*domain.Proyecto).ID = id
	}
}

// asignarIDUsuario emula la generacion de clave primaria del repositorio.
func asignarIDUsuario(id uint) func(mock.Arguments) {
	return func(args mock.Arguments) {
		args.Get(1).(*domain.Usuario).ID = id
	}
}
