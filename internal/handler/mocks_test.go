package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	"ingenieria-y-calidad/internal/domain"
)

type mockServicioProyectos struct {
	mock.Mock
}

func nuevoMockServicioProyectos(t *testing.T) *mockServicioProyectos {
	t.Helper()
	m := &mockServicioProyectos{}
	m.Mock.Test(t)
	return m
}

func (m *mockServicioProyectos) Crear(ctx context.Context, nuevo domain.NuevoProyecto) (*domain.Proyecto, error) {
	ret := m.Called(ctx, nuevo)
	var proyecto *domain.Proyecto
	if ret.Get(0) != nil {
		proyecto = ret.Get(0).(*domain.Proyecto)
	}
	return proyecto, ret.Error(1)
}

func (m *mockServicioProyectos) Listar(ctx context.Context) ([]domain.Proyecto, error) {
	ret := m.Called(ctx)
	var proyectos []domain.Proyecto
	if ret.Get(0) != nil {
		proyectos = ret.Get(0).([]domain.Proyecto)
	}
	return proyectos, ret.Error(1)
}

func (m *mockServicioProyectos) ObtenerPorID(ctx context.Context, id uint) (*domain.Proyecto, error) {
	ret := m.Called(ctx, id)
	var proyecto *domain.Proyecto
	if ret.Get(0) != nil {
		proyecto = ret.Get(0).(*domain.Proyecto)
	}
	return proyecto, ret.Error(1)
}

func (m *mockServicioProyectos) Actualizar(ctx context.Context, id uint, actualizacion domain.ActualizacionProyecto) (*domain.Proyecto, error) {
	ret := m.Called(ctx, id, actualizacion)
	var proyecto *domain.Proyecto
	if ret.Get(0) != nil {
		proyecto = ret.Get(0).(*domain.Proyecto)
	}
	return proyecto, ret.Error(1)
}

func (m *mockServicioProyectos) CambiarEstado(ctx context.Context, id uint, estado domain.EstadoProyecto) (*domain.Proyecto, error) {
	ret := m.Called(ctx, id, estado)
	var proyecto *domain.Proyecto
	if ret.Get(0) != nil {
		proyecto = ret.Get(0).(*domain.Proyecto)
	}
	return proyecto, ret.Error(1)
}

func (m *mockServicioProyectos) Eliminar(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockServicioProyectos) AgregarUsuario(ctx context.Context, proyectoID uint, entrada domain.UsuarioInput) (*domain.Usuario, error) {
	ret := m.Called(ctx, proyectoID, entrada)
	var usuario *domain.Usuario
	if ret.Get(0) != nil {
		usuario = ret.Get(0).(*domain.Usuario)
	}
	return usuario, ret.Error(1)
}

func (m *mockServicioProyectos) QuitarUsuario(ctx context.Context, proyectoID, usuarioID uint) error {
	return m.Called(ctx, proyectoID, usuarioID).Error(0)
}

type mockServicioHistorias struct {
	mock.Mock
}

func nuevoMockServicioHistorias(t *testing.T) *mockServicioHistorias {
	t.Helper()
	m := &mockServicioHistorias{}
	m.Mock.Test(t)
	return m
}

func (m *mockServicioHistorias) Crear(ctx context.Context, proyectoID uint, nuevo domain.NuevaHistoria) (*domain.Historia, error) {
	ret := m.Called(ctx, proyectoID, nuevo)
	var historia *domain.Historia
	if ret.Get(0) != nil {
		historia = ret.Get(0).(*domain.Historia)
	}
	return historia, ret.Error(1)
}

func (m *mockServicioHistorias) Listar(ctx context.Context, proyectoID uint) ([]domain.Historia, error) {
	ret := m.Called(ctx, proyectoID)
	var historias []domain.Historia
	if ret.Get(0) != nil {
		historias = ret.Get(0).([]domain.Historia)
	}
	return historias, ret.Error(1)
}

func (m *mockServicioHistorias) Actualizar(ctx context.Context, proyectoID, historiaID uint, actualizacion domain.ActualizacionHistoria) (*domain.Historia, error) {
	ret := m.Called(ctx, proyectoID, historiaID, actualizacion)
	var historia *domain.Historia
	if ret.Get(0) != nil {
		historia = ret.Get(0).(*domain.Historia)
	}
	return historia, ret.Error(1)
}

type mockServicioSprints struct {
	mock.Mock
}

func nuevoMockServicioSprints(t *testing.T) *mockServicioSprints {
	t.Helper()
	m := &mockServicioSprints{}
	m.Mock.Test(t)
	return m
}

func (m *mockServicioSprints) Crear(ctx context.Context, proyectoID uint, nuevo domain.NuevoSprint) (*domain.Sprint, error) {
	ret := m.Called(ctx, proyectoID, nuevo)
	var sprint *domain.Sprint
	if ret.Get(0) != nil {
		sprint = ret.Get(0).(*domain.Sprint)
	}
	return sprint, ret.Error(1)
}

func (m *mockServicioSprints) Listar(ctx context.Context, proyectoID uint) ([]domain.Sprint, error) {
	ret := m.Called(ctx, proyectoID)
	var sprints []domain.Sprint
	if ret.Get(0) != nil {
		sprints = ret.Get(0).([]domain.Sprint)
	}
	return sprints, ret.Error(1)
}

func (m *mockServicioSprints) CambiarEstado(ctx context.Context, proyectoID, sprintID uint, destino domain.EstadoSprint) (*domain.Sprint, error) {
	ret := m.Called(ctx, proyectoID, sprintID, destino)
	var sprint *domain.Sprint
	if ret.Get(0) != nil {
		sprint = ret.Get(0).(*domain.Sprint)
	}
	return sprint, ret.Error(1)
}
