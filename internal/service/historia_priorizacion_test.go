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

func TestPriorizarBacklogService(t *testing.T) {
	t.Parallel()

	t.Run("E-07 ordena por WSJF y excluye Hecho", func(t *testing.T) {
		t.Parallel()

		servicio, proyectos, historias := nuevoServicioHistorias(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).
			Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
		historias.On("ListarPorProyecto", mock.Anything, uint(1)).
			Return([]domain.Historia{
				{ID: 1, Estado: domain.EstadoHistoriaBacklog, StoryPoints: floatPtr(5), ValorNegocio: floatPtr(8)},
				{ID: 2, Estado: domain.EstadoHistoriaBacklog, StoryPoints: floatPtr(5), ValorNegocio: floatPtr(3)},
				{ID: 3, Estado: domain.EstadoHistoriaHecho, StoryPoints: floatPtr(5), ValorNegocio: floatPtr(21)},
			}, nil)

		resultado, err := servicio.Priorizar(context.Background(), 1)

		require.NoError(t, err)
		require.Len(t, resultado, 2)
		assert.Equal(t, uint(1), resultado[0].Historia.ID)
		assert.Equal(t, uint(2), resultado[1].Historia.ID)
		proyectos.AssertExpectations(t)
		historias.AssertExpectations(t)
	})

	t.Run("proyecto inexistente", func(t *testing.T) {
		t.Parallel()

		servicio, proyectos, historias := nuevoServicioHistorias(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(99)).
			Return(nil, domain.ErrProyectoNoEncontrado)

		resultado, err := servicio.Priorizar(context.Background(), 99)

		require.ErrorIs(t, err, domain.ErrProyectoNoEncontrado)
		assert.Nil(t, resultado)
		proyectos.AssertExpectations(t)
		historias.AssertNotCalled(t, "ListarPorProyecto", mock.Anything, mock.Anything)
	})

	t.Run("error del repositorio se conserva", func(t *testing.T) {
		t.Parallel()

		servicio, proyectos, historias := nuevoServicioHistorias(t)
		errPersistencia := errors.New("fallo de persistencia")
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).
			Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
		historias.On("ListarPorProyecto", mock.Anything, uint(1)).
			Return(nil, errPersistencia)

		_, err := servicio.Priorizar(context.Background(), 1)

		require.ErrorIs(t, err, errPersistencia)
		proyectos.AssertExpectations(t)
		historias.AssertExpectations(t)
	})

	t.Run("listado nil devuelve slice vacio", func(t *testing.T) {
		t.Parallel()

		servicio, proyectos, historias := nuevoServicioHistorias(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).
			Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
		historias.On("ListarPorProyecto", mock.Anything, uint(1)).
			Return(nil, nil)

		resultado, err := servicio.Priorizar(context.Background(), 1)

		require.NoError(t, err)
		assert.NotNil(t, resultado)
		assert.Empty(t, resultado)
		proyectos.AssertExpectations(t)
		historias.AssertExpectations(t)
	})
}

func TestCrearHistoriaValoresDeNegocio(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		nuevo       domain.NuevaHistoria
		errEsperado error
		escenario   string
	}{
		{
			nombre: "valor de negocio valido",
			nuevo: domain.NuevaHistoria{
				Titulo: "Historia con valor", ValorNegocio: floatPtr(8),
			},
			escenario: "E-05",
		},
		{
			nombre: "criticidad de tiempo valida",
			nuevo: domain.NuevaHistoria{
				Titulo: "Historia con criticidad", CriticidadTiempo: floatPtr(13),
			},
			escenario: "E-05",
		},
		{
			nombre: "reduccion de riesgo valida",
			nuevo: domain.NuevaHistoria{
				Titulo: "Historia con riesgo", ReduccionRiesgo: floatPtr(1),
			},
			escenario: "E-05",
		},
		{
			nombre: "valor de negocio fuera de escala",
			nuevo: domain.NuevaHistoria{
				Titulo: "Historia invalida", ValorNegocio: floatPtr(4),
			},
			errEsperado: domain.ErrValorNegocioInvalido,
			escenario:   "E-05",
		},
		{
			nombre: "valor de negocio cero",
			nuevo: domain.NuevaHistoria{
				Titulo: "Historia invalida", ValorNegocio: floatPtr(0),
			},
			errEsperado: domain.ErrValorNegocioInvalido,
			escenario:   "E-05",
		},
		{
			nombre: "criticidad de tiempo negativa",
			nuevo: domain.NuevaHistoria{
				Titulo: "Historia invalida", CriticidadTiempo: floatPtr(-1),
			},
			errEsperado: domain.ErrCriticidadTiempoInvalida,
			escenario:   "E-05",
		},
		{
			nombre: "reduccion de riesgo decimal",
			nuevo: domain.NuevaHistoria{
				Titulo: "Historia invalida", ReduccionRiesgo: floatPtr(0.5),
			},
			errEsperado: domain.ErrReduccionRiesgoInvalida,
			escenario:   "E-05",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()

			servicio, proyectos, historias := nuevoServicioHistorias(t)
			if c.errEsperado == nil {
				proyectos.On("ObtenerPorID", mock.Anything, uint(1)).
					Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
				historias.On("SiguienteNumero", mock.Anything, uint(1)).Return(1, nil)
				historias.On("Crear", mock.Anything, mock.Anything).Return(nil)
			}

			historia, err := servicio.Crear(context.Background(), 1, c.nuevo)

			if c.errEsperado != nil {
				require.ErrorIs(t, err, c.errEsperado, "escenario BDD %s", c.escenario)
				assert.Nil(t, historia)
				proyectos.AssertNotCalled(t, "ObtenerPorID", mock.Anything, mock.Anything)
				historias.AssertNotCalled(t, "Crear", mock.Anything, mock.Anything)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, historia)
			proyectos.AssertExpectations(t)
			historias.AssertExpectations(t)
		})
	}
}

func TestActualizarHistoriaValoresDeNegocio(t *testing.T) {
	t.Parallel()

	t.Run("PATCH solo con valor de negocio se persiste", func(t *testing.T) {
		t.Parallel()

		servicio, proyectos, historias := nuevoServicioHistorias(t)
		existente := &domain.Historia{
			ID: 10, ProyectoID: 1, Titulo: "Historia", Estado: domain.EstadoHistoriaBacklog,
			Prioridad: domain.PrioridadMedia,
		}
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).
			Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
		historias.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(existente, nil)
		historias.On("Actualizar", mock.Anything, mock.Anything).Return(nil)

		actualizada, err := servicio.Actualizar(context.Background(), 1, 10, domain.ActualizacionHistoria{
			ValorNegocio: floatPtr(8),
		})

		require.NoError(t, err)
		require.NotNil(t, actualizada.ValorNegocio)
		assert.Equal(t, 8.0, *actualizada.ValorNegocio)
		proyectos.AssertExpectations(t)
		historias.AssertExpectations(t)
	})

	casos := []struct {
		nombre        string
		actualizacion domain.ActualizacionHistoria
		errEsperado   error
		campoEsperado string
		escenario     string
	}{
		{
			nombre:        "valor de negocio fuera de escala",
			actualizacion: domain.ActualizacionHistoria{ValorNegocio: floatPtr(22)},
			errEsperado:   domain.ErrValorNegocioInvalido,
			campoEsperado: "valor_negocio",
			escenario:     "E-06",
		},
		{
			nombre:        "criticidad de tiempo cero",
			actualizacion: domain.ActualizacionHistoria{CriticidadTiempo: floatPtr(0)},
			errEsperado:   domain.ErrCriticidadTiempoInvalida,
			campoEsperado: "criticidad_tiempo",
			escenario:     "E-06",
		},
		{
			nombre:        "reduccion de riesgo fuera de escala",
			actualizacion: domain.ActualizacionHistoria{ReduccionRiesgo: floatPtr(4)},
			errEsperado:   domain.ErrReduccionRiesgoInvalida,
			campoEsperado: "reduccion_riesgo",
			escenario:     "E-06",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()

			servicio, proyectos, historias := nuevoServicioHistorias(t)

			historia, err := servicio.Actualizar(context.Background(), 1, 10, c.actualizacion)

			require.ErrorIs(t, err, c.errEsperado, "escenario BDD %s (campo %s)", c.escenario, c.campoEsperado)
			assert.Nil(t, historia)
			proyectos.AssertNotCalled(t, "ObtenerPorID", mock.Anything, mock.Anything)
			historias.AssertNotCalled(t, "Actualizar", mock.Anything, mock.Anything)
		})
	}
}
