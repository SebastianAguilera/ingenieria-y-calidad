package handler

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

// Tests de US-17: endpoints de acuerdo y cancelación de Planning Poker y
// exposición del resultado final en detalle e historial.

func detalleCerradoDeEjemplo(puntos float64) *domain.DetalleSesionPoker {
	detalle := detallePokerDeEjemplo(domain.EstadoSesionCerrada)
	cerrado := time.Date(2026, 11, 14, 18, 30, 0, 0, time.UTC)
	sesionID, rondaID, anaID := uint(9), uint(14), uint(2)
	detalle.Sesion.PuntosAcordados = &puntos
	detalle.Sesion.CerradoEn = &cerrado
	detalle.Sesion.Estimacion = &domain.StoryEstimate{
		ID:          11,
		HistoriaID:  5,
		SesionID:    &sesionID,
		RondaID:     &rondaID,
		Puntos:      puntos,
		Fuente:      domain.FuenteEstimacionPlanningPoker,
		EstimadoPor: &anaID,
		EstimadoEn:  cerrado,
	}
	return detalle
}

func detalleCanceladoDeEjemplo() *domain.DetalleSesionPoker {
	detalle := detallePokerDeEjemplo(domain.EstadoSesionCancelada)
	cancelado := time.Date(2026, 11, 14, 19, 0, 0, 0, time.UTC)
	detalle.Sesion.CanceladoEn = &cancelado
	return detalle
}

func TestAcordarPokerHandler(t *testing.T) {
	const ruta = "/api/proyectos/1/poker-sessions/9/acuerdo"

	t.Run("E-01 acuerda y devuelve la sesión cerrada con su story_estimate", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("Acordar", mock.Anything, uint(1), uint(9), uint(2), 5.0).Return(detalleCerradoDeEjemplo(5), nil)

		w := ejecutar(t, router, http.MethodPost, ruta, `{"puntos": 5}`)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		respuesta := decodificarDetallePoker(t, w)
		assert.Equal(t, domain.EstadoSesionCerrada, respuesta.Estado)
		require.NotNil(t, respuesta.PuntosAcordados)
		assert.Equal(t, 5.0, *respuesta.PuntosAcordados)
		require.NotNil(t, respuesta.CerradoEn)
		assert.Equal(t, "2026-11-14T18:30:00Z", *respuesta.CerradoEn)
		require.NotNil(t, respuesta.StoryEstimate)
		assert.Equal(t, uint(11), respuesta.StoryEstimate.ID)
		assert.Equal(t, uint(5), respuesta.StoryEstimate.HistoriaID)
		assert.Equal(t, uint(9), *respuesta.StoryEstimate.SesionID)
		assert.Equal(t, uint(14), *respuesta.StoryEstimate.RondaID)
		assert.Equal(t, 5.0, respuesta.StoryEstimate.Puntos)
		assert.Equal(t, "planning_poker", respuesta.StoryEstimate.Fuente)
		assert.Equal(t, uint(2), *respuesta.StoryEstimate.EstimadoPor)
		assert.Equal(t, "2026-11-14T18:30:00Z", respuesta.StoryEstimate.EstimadoEn)
	})

	t.Run("E-02 acepta 0 y 0.5", func(t *testing.T) {
		for _, caso := range []struct {
			cuerpo string
			puntos float64
		}{{`{"puntos": 0}`, 0}, {`{"puntos": 0.5}`, 0.5}} {
			router, servicio := nuevoRouterDePoker(t, 2)
			servicio.On("Acordar", mock.Anything, uint(1), uint(9), uint(2), caso.puntos).
				Return(detalleCerradoDeEjemplo(caso.puntos), nil)

			w := ejecutar(t, router, http.MethodPost, ruta, caso.cuerpo)

			assert.Equal(t, http.StatusOK, w.Code, caso.cuerpo)
		}
	})

	t.Run("E-06 puntos no numéricos o ausentes responden 422 sin llamar al servicio", func(t *testing.T) {
		for _, cuerpo := range []string{`{"puntos": "cinco"}`, `{}`, `{"puntos": null}`, `{"puntos": true}`} {
			router, _ := nuevoRouterDePoker(t, 2)

			w := ejecutar(t, router, http.MethodPost, ruta, cuerpo)

			require.Equal(t, http.StatusUnprocessableEntity, w.Code, cuerpo)
			respuesta := decodificarError(t, w)
			assert.Equal(t, codigoPuntosPokerInvalidos, respuesta.Error)
			assert.Equal(t, "puntos", respuesta.Campo)
		}
	})

	t.Run("E-06 puntos fuera de escala rechazados por el servicio", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("Acordar", mock.Anything, uint(1), uint(9), uint(2), 4.0).Return(nil, domain.ErrPuntosPokerInvalidos)

		w := ejecutar(t, router, http.MethodPost, ruta, `{"puntos": 4}`)

		require.Equal(t, http.StatusUnprocessableEntity, w.Code)
		assert.Equal(t, codigoPuntosPokerInvalidos, decodificarError(t, w).Error)
	})

	t.Run("E-07 cuerpo malformado responde 400", func(t *testing.T) {
		router, _ := nuevoRouterDePoker(t, 2)

		w := ejecutar(t, router, http.MethodPost, ruta, `{"puntos": `)

		require.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, codigoJSONInvalido, decodificarError(t, w).Error)
	})

	t.Run("E-04 sesión en Votando responde 409 SESION_POKER_NO_REVELADA", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("Acordar", mock.Anything, uint(1), uint(9), uint(2), 5.0).Return(nil, domain.ErrSesionPokerNoRevelada)

		w := ejecutar(t, router, http.MethodPost, ruta, `{"puntos": 5}`)

		require.Equal(t, http.StatusConflict, w.Code)
		assert.Equal(t, codigoSesionPokerNoRevelada, decodificarError(t, w).Error)
	})

	t.Run("E-05 sesión finalizada responde 409 SESION_POKER_CERRADA", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("Acordar", mock.Anything, uint(1), uint(9), uint(2), 5.0).Return(nil, domain.ErrSesionPokerCerrada)

		w := ejecutar(t, router, http.MethodPost, ruta, `{"puntos": 5}`)

		require.Equal(t, http.StatusConflict, w.Code)
		respuesta := decodificarError(t, w)
		assert.Equal(t, codigoSesionPokerCerrada, respuesta.Error)
		assert.Equal(t, "SESION_POKER_CERRADA", codigoSesionPokerCerrada)
		assert.Equal(t, domain.ErrSesionPokerCerrada.Error(), respuesta.Mensaje)
	})

	t.Run("E-18 sesión inexistente responde 404", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("Acordar", mock.Anything, uint(1), uint(9), uint(2), 5.0).Return(nil, domain.ErrSesionPokerNoEncontrada)

		w := ejecutar(t, router, http.MethodPost, ruta, `{"puntos": 5}`)

		require.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, codigoSesionPokerNoEncontrada, decodificarError(t, w).Error)
	})

	t.Run("E-19 sesionId inválido responde 400", func(t *testing.T) {
		router, _ := nuevoRouterDePoker(t, 2)

		w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/poker-sessions/abc/acuerdo", `{"puntos": 5}`)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("E-20 sin identidad responde 401", func(t *testing.T) {
		router, _ := nuevoRouterDePoker(t, 0)

		w := ejecutar(t, router, http.MethodPost, ruta, `{"puntos": 5}`)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestCancelarPokerHandler(t *testing.T) {
	const ruta = "/api/proyectos/1/poker-sessions/9/cancelar"

	t.Run("E-11 cancela y devuelve la sesión Cancelada sin story_estimate", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("Cancelar", mock.Anything, uint(1), uint(9), uint(2)).Return(detalleCanceladoDeEjemplo(), nil)

		w := ejecutar(t, router, http.MethodPost, ruta, nil)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		respuesta := decodificarDetallePoker(t, w)
		assert.Equal(t, domain.EstadoSesionCancelada, respuesta.Estado)
		require.NotNil(t, respuesta.CanceladoEn)
		assert.Equal(t, "2026-11-14T19:00:00Z", *respuesta.CanceladoEn)
		assert.Nil(t, respuesta.PuntosAcordados)

		var crudo map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &crudo))
		assert.NotContains(t, crudo, "story_estimate", "una sesión cancelada no informa story_estimate")
		assert.Contains(t, crudo, "puntos_acordados")
		assert.Nil(t, crudo["puntos_acordados"])
	})

	t.Run("E-12 sesión finalizada responde 409 SESION_POKER_CERRADA", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("Cancelar", mock.Anything, uint(1), uint(9), uint(2)).Return(nil, domain.ErrSesionPokerCerrada)

		w := ejecutar(t, router, http.MethodPost, ruta, nil)

		require.Equal(t, http.StatusConflict, w.Code)
		assert.Equal(t, codigoSesionPokerCerrada, decodificarError(t, w).Error)
	})

	t.Run("E-18 sesión inexistente responde 404", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("Cancelar", mock.Anything, uint(1), uint(9), uint(2)).Return(nil, domain.ErrSesionPokerNoEncontrada)

		w := ejecutar(t, router, http.MethodPost, ruta, nil)

		require.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, codigoSesionPokerNoEncontrada, decodificarError(t, w).Error)
	})

	t.Run("E-19 sesionId inválido responde 400", func(t *testing.T) {
		router, _ := nuevoRouterDePoker(t, 2)

		w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/poker-sessions/0/cancelar", nil)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("E-20 sin identidad responde 401", func(t *testing.T) {
		router, _ := nuevoRouterDePoker(t, 0)

		w := ejecutar(t, router, http.MethodPost, ruta, nil)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestHistorialPokerExponeResultadoFinal(t *testing.T) {
	router, servicio := nuevoRouterDePoker(t, 2)
	creado := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	cancelado := time.Date(2026, 11, 11, 9, 0, 0, 0, time.UTC)
	cerrado := time.Date(2026, 11, 12, 9, 0, 0, 0, time.UTC)
	cinco := 5.0
	sesionID := uint(3)
	servicio.On("ListarPorHistoria", mock.Anything, uint(1), uint(5)).Return([]domain.SesionPoker{
		{
			ID: 3, Estado: domain.EstadoSesionCerrada, RondaActual: 2, CreadoPor: 2, CreadoEn: creado,
			PuntosAcordados: &cinco, CerradoEn: &cerrado,
			Estimacion: &domain.StoryEstimate{ID: 11, HistoriaID: 5, SesionID: &sesionID, Puntos: 5,
				Fuente: domain.FuenteEstimacionPlanningPoker, EstimadoEn: cerrado},
		},
		{ID: 2, Estado: domain.EstadoSesionCancelada, RondaActual: 1, CreadoPor: 2, CreadoEn: creado, CanceladoEn: &cancelado},
	}, nil)

	w := ejecutar(t, router, http.MethodGet, "/api/proyectos/1/historias/5/poker-sessions", nil)

	require.Equal(t, http.StatusOK, w.Code)
	var respuestas []respuestaHistorialSesion
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuestas))
	require.Len(t, respuestas, 2)

	cerrada := respuestas[0]
	require.NotNil(t, cerrada.StoryEstimate, "E-15 la cerrada informa su story_estimate")
	assert.Equal(t, 5.0, cerrada.StoryEstimate.Puntos)
	assert.Equal(t, 5.0, *cerrada.PuntosAcordados)
	assert.Nil(t, cerrada.CanceladoEn)

	cancelada := respuestas[1]
	assert.Equal(t, domain.EstadoSesionCancelada, cancelada.Estado)
	require.NotNil(t, cancelada.CanceladoEn, "E-15 la cancelada informa cancelado_en")
	assert.Equal(t, "2026-11-11T09:00:00Z", *cancelada.CanceladoEn)
	assert.Nil(t, cancelada.StoryEstimate)
}
