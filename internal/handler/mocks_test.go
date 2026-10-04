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
