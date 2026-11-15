package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

type respuestaHistoriaPriorizadaTest struct {
	ID               uint     `json:"id"`
	Codigo           string   `json:"codigo"`
	Titulo           string   `json:"titulo"`
	Prioridad        string   `json:"prioridad"`
	Estado           string   `json:"estado"`
	StoryPoints      *float64 `json:"story_points"`
	ValorNegocio     *float64 `json:"valor_negocio"`
	CriticidadTiempo *float64 `json:"criticidad_tiempo"`
	ReduccionRiesgo  *float64 `json:"reduccion_riesgo"`
	ROI              *float64 `json:"roi"`
	WSJF             *float64 `json:"wsjf"`
}

type respuestaListadoPriorizadoTest struct {
	Total     int                               `json:"total"`
	Historias []respuestaHistoriaPriorizadaTest `json:"historias"`
}

func TestBacklogPriorizadoHandler(t *testing.T) {
	t.Parallel()

	t.Run("E-10 contrato completo, orden y nulos", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeHistorias(t)
		servicio.On("Priorizar", mock.Anything, uint(1)).Return([]domain.HistoriaPriorizada{
			{
				Historia: domain.Historia{
					ID: 1, Codigo: "US-01", Titulo: "Priorizada", Prioridad: domain.PrioridadAlta,
					Estado: domain.EstadoHistoriaBacklog, StoryPoints: floatPtr(5),
					ValorNegocio: floatPtr(8), CriticidadTiempo: floatPtr(3), ReduccionRiesgo: floatPtr(2),
				},
				ROI:  floatPtr(1.6),
				WSJF: floatPtr(2.6),
			},
			{
				Historia: domain.Historia{
					ID: 2, Codigo: "US-02", Titulo: "Sin estimar", Prioridad: domain.PrioridadMedia,
					Estado: domain.EstadoHistoriaBacklog, ValorNegocio: floatPtr(21),
				},
			},
		}, nil)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/1/backlog/priorizado", nil)

		require.Equal(t, http.StatusOK, w.Code)
		var respuesta respuestaListadoPriorizadoTest
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta))
		require.Equal(t, 2, respuesta.Total)
		require.Len(t, respuesta.Historias, 2)
		assert.Equal(t, uint(1), respuesta.Historias[0].ID)
		require.NotNil(t, respuesta.Historias[0].ValorNegocio)
		assert.Equal(t, 8.0, *respuesta.Historias[0].ValorNegocio)
		require.NotNil(t, respuesta.Historias[0].ROI)
		assert.InDelta(t, 1.6, *respuesta.Historias[0].ROI, 1e-9)
		require.NotNil(t, respuesta.Historias[0].WSJF)
		assert.InDelta(t, 2.6, *respuesta.Historias[0].WSJF, 1e-9)
		assert.Nil(t, respuesta.Historias[1].ROI)
		assert.Nil(t, respuesta.Historias[1].WSJF)
		servicio.AssertExpectations(t)
	})

	t.Run("E-09 proyecto inexistente o no miembro", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeHistorias(t)
		servicio.On("Priorizar", mock.Anything, uint(1)).Return(nil, domain.ErrProyectoNoEncontrado)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/1/backlog/priorizado", nil)

		require.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, codigoProyectoNoEncontrado, decodificarError(t, w).Error)
		servicio.AssertExpectations(t)
	})

	t.Run("error interno", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeHistorias(t)
		servicio.On("Priorizar", mock.Anything, uint(1)).Return(nil, errors.New("boom"))

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/1/backlog/priorizado", nil)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, codigoErrorInterno, decodificarError(t, w).Error)
		servicio.AssertExpectations(t)
	})

	t.Run("id invalido responde 400 sin llamar al servicio", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeHistorias(t)

		for _, ruta := range []string{
			"/api/proyectos/abc/backlog/priorizado",
			"/api/proyectos/0/backlog/priorizado",
		} {
			w := ejecutar(t, router, http.MethodGet, ruta, nil)
			require.Equal(t, http.StatusBadRequest, w.Code, "ruta: %s", ruta)
			assert.Equal(t, codigoParametroInvalido, decodificarError(t, w).Error)
		}
		servicio.AssertNotCalled(t, "Priorizar", mock.Anything, mock.Anything)
	})
}

func TestAltaHistoriaConValoresDeNegocioHandler(t *testing.T) {
	t.Parallel()

	t.Run("alta con los tres valores", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeHistorias(t)
		esperada := historiaDeEjemplo()
		esperada.ValorNegocio = floatPtr(8)
		esperada.CriticidadTiempo = floatPtr(5)
		esperada.ReduccionRiesgo = floatPtr(3)
		servicio.On("Crear", mock.Anything, uint(1), domain.NuevaHistoria{
			Titulo:           "Con valores",
			ValorNegocio:     floatPtr(8),
			CriticidadTiempo: floatPtr(5),
			ReduccionRiesgo:  floatPtr(3),
		}).Return(esperada, nil)

		w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/historias",
			`{"titulo":"Con valores","valor_negocio":8,"criticidad_tiempo":5,"reduccion_riesgo":3}`)

		require.Equal(t, http.StatusCreated, w.Code)
		var respuesta struct {
			ValorNegocio     *float64 `json:"valor_negocio"`
			CriticidadTiempo *float64 `json:"criticidad_tiempo"`
			ReduccionRiesgo  *float64 `json:"reduccion_riesgo"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta))
		require.NotNil(t, respuesta.ValorNegocio)
		assert.Equal(t, 8.0, *respuesta.ValorNegocio)
		require.NotNil(t, respuesta.CriticidadTiempo)
		assert.Equal(t, 5.0, *respuesta.CriticidadTiempo)
		require.NotNil(t, respuesta.ReduccionRiesgo)
		assert.Equal(t, 3.0, *respuesta.ReduccionRiesgo)
		servicio.AssertExpectations(t)
	})

	t.Run("valor no numerico responde 422 con campo", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeHistorias(t)

		w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/historias",
			`{"titulo":"Con valores","valor_negocio":"ocho"}`)

		require.Equal(t, http.StatusUnprocessableEntity, w.Code)
		respuesta := decodificarError(t, w)
		assert.Equal(t, codigoValidacion, respuesta.Error)
		assert.Equal(t, "valor_negocio", respuesta.Campo)
		servicio.AssertNotCalled(t, "Crear", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("valor fuera de escala traducido desde el service", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeHistorias(t)
		servicio.On("Crear", mock.Anything, uint(1), domain.NuevaHistoria{
			Titulo:       "Historia invalida",
			ValorNegocio: floatPtr(4),
		}).Return(nil, domain.ErrValorNegocioInvalido)

		w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/historias",
			`{"titulo":"Historia invalida","valor_negocio":4}`)

		require.Equal(t, http.StatusUnprocessableEntity, w.Code)
		respuesta := decodificarError(t, w)
		assert.Equal(t, codigoValidacion, respuesta.Error)
		assert.Equal(t, "valor_negocio", respuesta.Campo)
		servicio.AssertExpectations(t)
	})
}

func TestPatchHistoriaConValoresDeNegocioHandler(t *testing.T) {
	t.Parallel()

	t.Run("PATCH solo con valor de negocio", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeHistorias(t)
		actualizada := historiaDeEjemplo()
		actualizada.ValorNegocio = floatPtr(8)
		servicio.On("Actualizar", mock.Anything, uint(1), uint(10), domain.ActualizacionHistoria{
			ValorNegocio: floatPtr(8),
		}).Return(actualizada, nil)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/historias/10",
			`{"valor_negocio":8}`)

		require.Equal(t, http.StatusOK, w.Code)
		var respuesta struct {
			ValorNegocio *float64 `json:"valor_negocio"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta))
		require.NotNil(t, respuesta.ValorNegocio)
		assert.Equal(t, 8.0, *respuesta.ValorNegocio)
		servicio.AssertExpectations(t)
	})

	t.Run("valor fuera de escala traducido desde el service", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeHistorias(t)
		servicio.On("Actualizar", mock.Anything, uint(1), uint(10), domain.ActualizacionHistoria{
			CriticidadTiempo: floatPtr(4),
		}).Return(nil, domain.ErrCriticidadTiempoInvalida)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/historias/10",
			`{"criticidad_tiempo":4}`)

		require.Equal(t, http.StatusUnprocessableEntity, w.Code)
		respuesta := decodificarError(t, w)
		assert.Equal(t, codigoValidacion, respuesta.Error)
		assert.Equal(t, "criticidad_tiempo", respuesta.Campo)
		servicio.AssertExpectations(t)
	})
}
