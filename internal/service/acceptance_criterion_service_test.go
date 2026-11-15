package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func nuevoServicioCriterios(t *testing.T) (*AcceptanceCriterionService, *MockProyectoRepository, *MockHistoriaRepository, *MockAcceptanceCriterionRepository) {
	t.Helper()
	proyectos := nuevoMockProyectoRepository(t)
	historias := nuevoMockHistoriaRepository(t)
	criterios := nuevoMockAcceptanceCriterionRepository(t)
	return NewAcceptanceCriterionService(proyectos, historias, criterios), proyectos, historias, criterios
}

func TestAcceptanceCriterionServiceCrearAsignaPosicionSiguiente(t *testing.T) {
	ctx := context.Background()
	servicio, proyectos, historias, criterios := nuevoServicioCriterios(t)

	proyectos.On("ObtenerPorID", ctx, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
	historias.On("ObtenerPorID", ctx, uint(1), uint(2)).Return(&domain.Historia{ID: 2}, nil)
	criterios.On("SiguientePosicion", ctx, uint(2)).Return(3, nil)
	criterios.On("Crear", ctx, mock.MatchedBy(func(c *domain.AcceptanceCriterion) bool {
		return c.HistoriaID == 2 && c.Position == 3 && c.Descripcion == "El usuario ve el saldo" && !c.IsMet
	})).Return(nil)

	criterio, err := servicio.Crear(ctx, 1, 2, domain.NuevoAcceptanceCriterion{Descripcion: "  El usuario ve el saldo  "})

	require.NoError(t, err)
	assert.Equal(t, 3, criterio.Position)
	assert.False(t, criterio.IsMet)
	proyectos.AssertExpectations(t)
	historias.AssertExpectations(t)
	criterios.AssertExpectations(t)
}

func TestAcceptanceCriterionServiceCrearRechazaDescripcionVacia(t *testing.T) {
	ctx := context.Background()
	servicio, _, _, _ := nuevoServicioCriterios(t)

	_, err := servicio.Crear(ctx, 1, 2, domain.NuevoAcceptanceCriterion{Descripcion: "   "})

	assert.ErrorIs(t, err, domain.ErrDescripcionCriterioObligatoria)
}

func TestAcceptanceCriterionServiceCrearFallaSiProyectoNoExiste(t *testing.T) {
	ctx := context.Background()
	servicio, proyectos, _, _ := nuevoServicioCriterios(t)

	proyectos.On("ObtenerPorID", ctx, uint(1)).Return(nil, domain.ErrProyectoNoEncontrado)

	_, err := servicio.Crear(ctx, 1, 2, domain.NuevoAcceptanceCriterion{Descripcion: "Algo"})

	assert.ErrorIs(t, err, domain.ErrProyectoNoEncontrado)
}

func TestAcceptanceCriterionServiceListarDevuelveListaVacia(t *testing.T) {
	ctx := context.Background()
	servicio, proyectos, historias, criterios := nuevoServicioCriterios(t)

	proyectos.On("ObtenerPorID", ctx, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
	historias.On("ObtenerPorID", ctx, uint(1), uint(2)).Return(&domain.Historia{ID: 2}, nil)
	criterios.On("ListarPorHistoria", ctx, uint(2)).Return(nil, nil)

	lista, err := servicio.Listar(ctx, 1, 2)

	require.NoError(t, err)
	assert.NotNil(t, lista)
	assert.Empty(t, lista)
}

func TestAcceptanceCriterionServiceActualizarRechazaActualizacionVacia(t *testing.T) {
	ctx := context.Background()
	servicio, _, _, _ := nuevoServicioCriterios(t)

	_, err := servicio.Actualizar(ctx, 1, 2, 3, domain.ActualizacionAcceptanceCriterion{})

	assert.ErrorIs(t, err, domain.ErrActualizacionCriterioVacia)
}

func TestAcceptanceCriterionServiceActualizarMarcaCumplido(t *testing.T) {
	ctx := context.Background()
	servicio, proyectos, historias, criterios := nuevoServicioCriterios(t)

	existente := &domain.AcceptanceCriterion{ID: 3, HistoriaID: 2, Descripcion: "Algo", Position: 1}
	proyectos.On("ObtenerPorID", ctx, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
	historias.On("ObtenerPorID", ctx, uint(1), uint(2)).Return(&domain.Historia{ID: 2}, nil)
	criterios.On("ObtenerPorID", ctx, uint(2), uint(3)).Return(existente, nil)
	criterios.On("Actualizar", ctx, mock.MatchedBy(func(c *domain.AcceptanceCriterion) bool {
		return c.IsMet && c.Descripcion == "Algo"
	})).Return(nil)

	cumplido := true
	criterio, err := servicio.Actualizar(ctx, 1, 2, 3, domain.ActualizacionAcceptanceCriterion{IsMet: &cumplido})

	require.NoError(t, err)
	assert.True(t, criterio.IsMet)
}

func TestAcceptanceCriterionServiceEliminarPropagaNoEncontrado(t *testing.T) {
	ctx := context.Background()
	servicio, proyectos, historias, criterios := nuevoServicioCriterios(t)

	proyectos.On("ObtenerPorID", ctx, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
	historias.On("ObtenerPorID", ctx, uint(1), uint(2)).Return(&domain.Historia{ID: 2}, nil)
	criterios.On("Eliminar", ctx, uint(2), uint(3)).Return(domain.ErrCriterioNoEncontrado)

	err := servicio.Eliminar(ctx, 1, 2, 3)

	assert.True(t, errors.Is(err, domain.ErrCriterioNoEncontrado))
}
