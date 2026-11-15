package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func TestCrearHistoriaConHorasEstimadas(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		horas       *float64
		errEsperado error
		escenario   string
	}{
		{"sin horas estimadas", nil, nil, "BDD-02"},
		{"horas estimadas enteras", floatPtr(5), nil, "BDD-01"},
		{"horas estimadas decimales", floatPtr(0.25), nil, "CL-01"},
		{"horas estimadas cero", floatPtr(0), domain.ErrHorasEstimadasInvalidas, "BDD-09"},
		{"horas estimadas negativas", floatPtr(-2), domain.ErrHorasEstimadasInvalidas, "BDD-10"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()

			servicio, proyectos, historias := nuevoServicioHistorias(t)
			if c.errEsperado == nil {
				proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
				historias.On("SiguienteNumero", mock.Anything, uint(1)).Return(1, nil)
				historias.On("Crear", mock.Anything, mock.Anything).Return(nil)
			}

			historia, err := servicio.Crear(context.Background(), 1, domain.NuevaHistoria{Titulo: "Historia", HorasEstimadas: c.horas})

			if c.errEsperado != nil {
				require.ErrorIs(t, err, c.errEsperado, "escenario BDD %s", c.escenario)
				assert.Nil(t, historia)
				proyectos.AssertExpectations(t)
				historias.AssertExpectations(t)
				return
			}
			require.NoError(t, err, "escenario BDD %s", c.escenario)
			assert.Equal(t, c.horas, historia.HorasEstimadas)
			proyectos.AssertExpectations(t)
			historias.AssertExpectations(t)
		})
	}
}

func TestActualizarHistoriaHorasEstimadas(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre        string
		actualizacion domain.ActualizacionHistoria
		preparar      bool
		errEsperado   error
		verificar     func(t *testing.T, historia *domain.Historia)
		escenario     string
	}{
		{
			nombre:        "registrar horas estimadas sin tocar story points ni esfuerzo real",
			actualizacion: domain.ActualizacionHistoria{HorasEstimadas: floatPtr(8)},
			preparar:      true,
			verificar: func(t *testing.T, historia *domain.Historia) {
				require.NotNil(t, historia.HorasEstimadas)
				assert.Equal(t, 8.0, *historia.HorasEstimadas)
				assert.Equal(t, "Original", historia.Titulo, "los campos no informados no cambian")
				require.NotNil(t, historia.StoryPoints)
				assert.Equal(t, 5.0, *historia.StoryPoints, "BDD-07: story_points no cambia")
				require.NotNil(t, historia.EsfuerzoReal)
				assert.Equal(t, 3.0, *historia.EsfuerzoReal, "BDD-07: esfuerzo_real no cambia")
			},
			escenario: "BDD-04/BDD-07",
		},
		{
			nombre:        "cero es invalido",
			actualizacion: domain.ActualizacionHistoria{HorasEstimadas: floatPtr(0)},
			errEsperado:   domain.ErrHorasEstimadasInvalidas,
			escenario:     "BDD-09",
		},
		{
			nombre:        "negativo es invalido",
			actualizacion: domain.ActualizacionHistoria{HorasEstimadas: floatPtr(-2)},
			errEsperado:   domain.ErrHorasEstimadasInvalidas,
			escenario:     "BDD-10",
		},
		{
			nombre:        "null como unico campo equivale a no informado",
			actualizacion: domain.ActualizacionHistoria{},
			errEsperado:   domain.ErrActualizacionHistoriaVacia,
			escenario:     "CL-04",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()

			servicio, proyectos, historias := nuevoServicioHistorias(t)
			if c.preparar {
				proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
				historias.On("ObtenerPorID", mock.Anything, uint(1), uint(5)).Return(&domain.Historia{
					ID: 5, ProyectoID: 1, Titulo: "Original",
					StoryPoints: floatPtr(5), EsfuerzoReal: floatPtr(3), HorasEstimadas: floatPtr(2),
				}, nil)
				historias.On("Actualizar", mock.Anything, mock.Anything).Return(nil)
			}

			historia, err := servicio.Actualizar(context.Background(), 1, 5, c.actualizacion)

			if c.errEsperado != nil {
				require.ErrorIs(t, err, c.errEsperado, "escenario BDD %s", c.escenario)
				assert.Nil(t, historia)
				proyectos.AssertExpectations(t)
				historias.AssertExpectations(t)
				return
			}
			require.NoError(t, err, "escenario BDD %s", c.escenario)
			c.verificar(t, historia)
			proyectos.AssertExpectations(t)
			historias.AssertExpectations(t)
		})
	}
}

func TestActualizarHistoriaSinHorasEstimadasNoBorraElValor(t *testing.T) {
	t.Parallel()

	servicio, proyectos, historias := nuevoServicioHistorias(t)
	proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
	historias.On("ObtenerPorID", mock.Anything, uint(1), uint(5)).Return(&domain.Historia{
		ID: 5, ProyectoID: 1, Titulo: "Original", HorasEstimadas: floatPtr(8),
	}, nil)
	historias.On("Actualizar", mock.Anything, mock.Anything).Return(nil)

	historia, err := servicio.Actualizar(context.Background(), 1, 5, domain.ActualizacionHistoria{Titulo: strPtr("Nuevo titulo")})

	require.NoError(t, err)
	require.NotNil(t, historia.HorasEstimadas, "RN-05: no informado no borra las horas estimadas")
	assert.Equal(t, 8.0, *historia.HorasEstimadas)
	proyectos.AssertExpectations(t)
	historias.AssertExpectations(t)
}
