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

func (m *MockProyectoRepository) CrearConIntegrantes(ctx context.Context, p *domain.Proyecto, integrantes []domain.Integrante) error {
	return errorDeRetorno(m.Called(ctx, p, integrantes))
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

type MockIntegranteRepository struct {
	mock.Mock
}

func nuevoMockIntegranteRepository(t *testing.T) *MockIntegranteRepository {
	t.Helper()
	m := &MockIntegranteRepository{}
	m.Mock.Test(t)
	return m
}

func (m *MockIntegranteRepository) ObtenerPorEmail(ctx context.Context, email string) (*domain.Integrante, error) {
	ret := m.Called(ctx, email)
	var integrante *domain.Integrante
	if ret.Get(0) != nil {
		integrante = ret.Get(0).(*domain.Integrante)
	}
	return integrante, ret.Error(1)
}

func (m *MockIntegranteRepository) Crear(ctx context.Context, i *domain.Integrante) error {
	return errorDeRetorno(m.Called(ctx, i))
}

func (m *MockIntegranteRepository) ActualizarNombre(ctx context.Context, id uint, nombre string) error {
	return errorDeRetorno(m.Called(ctx, id, nombre))
}

func (m *MockIntegranteRepository) EstaAsociado(ctx context.Context, proyectoID, integranteID uint) (bool, error) {
	ret := m.Called(ctx, proyectoID, integranteID)
	return ret.Bool(0), ret.Error(1)
}

func (m *MockIntegranteRepository) Vincular(ctx context.Context, proyectoID, integranteID uint) error {
	return errorDeRetorno(m.Called(ctx, proyectoID, integranteID))
}

func (m *MockIntegranteRepository) Desvincular(ctx context.Context, proyectoID, integranteID uint) error {
	return errorDeRetorno(m.Called(ctx, proyectoID, integranteID))
}

func (m *MockIntegranteRepository) ListarPorProyecto(ctx context.Context, proyectoID uint) ([]domain.Integrante, error) {
	ret := m.Called(ctx, proyectoID)
	var integrantes []domain.Integrante
	if ret.Get(0) != nil {
		integrantes = ret.Get(0).([]domain.Integrante)
	}
	return integrantes, ret.Error(1)
}

func proyectoDePrueba(id uint, nombre string, inicio time.Time, fin *time.Time, estado domain.EstadoProyecto) *domain.Proyecto {
	return &domain.Proyecto{
		ID:          id,
		Nombre:      nombre,
		FechaInicio: inicio,
		FechaFin:    fin,
		Estado:      estado,
		Integrantes: []domain.Integrante{},
	}
}

// asignarIDProyecto emula la generacion de clave primaria del repositorio.
func asignarIDProyecto(id uint) func(mock.Arguments) {
	return func(args mock.Arguments) {
		args.Get(1).(*domain.Proyecto).ID = id
	}
}

// asignarIDIntegrante emula la generacion de clave primaria del repositorio.
func asignarIDIntegrante(id uint) func(mock.Arguments) {
	return func(args mock.Arguments) {
		args.Get(1).(*domain.Integrante).ID = id
	}
}
