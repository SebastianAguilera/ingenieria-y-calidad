package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

type servicioHistoriasPriorizada struct {
	servicioHistoriasQueFalla
	consultas int
}

func (s *servicioHistoriasPriorizada) Priorizar(context.Context, uint) ([]domain.HistoriaPriorizada, error) {
	s.consultas++
	valor := 8.0
	sp := 5.0
	roi := 1.6
	wsjf := 2.6
	return []domain.HistoriaPriorizada{{
		Historia: domain.Historia{
			ID: 1, Codigo: "US-01", Titulo: "Priorizada", Estado: domain.EstadoHistoriaBacklog,
			Prioridad: domain.PrioridadAlta, StoryPoints: &sp, ValorNegocio: &valor,
		},
		ROI:  &roi,
		WSJF: &wsjf,
	}}, nil
}

func TestRutaBacklogPriorizadoEnElRouter(t *testing.T) {
	t.Parallel()

	const ruta = "/api/proyectos/1/backlog/priorizado"

	t.Run("miembro llega al handler", func(t *testing.T) {
		t.Parallel()

		servicio := &servicioHistoriasPriorizada{}
		router := setupRouter(servicioQueFalla{}, servicio, servicioSprintsQueFalla{}, nil, nil, nil, nil, nil, nil, nil, nil, secretoDePruebaRouter)

		w := pedirProtegidoEnRouter(t, router, http.MethodGet, ruta)

		require.Equal(t, http.StatusOK, w.Code, "la ruta no debe colisionar con otra: %s", w.Body.String())
		var cuerpo struct {
			Total     int `json:"total"`
			Historias []struct {
				ID   uint     `json:"id"`
				ROI  *float64 `json:"roi"`
				WSJF *float64 `json:"wsjf"`
			} `json:"historias"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cuerpo))
		assert.Equal(t, 1, cuerpo.Total)
		require.Len(t, cuerpo.Historias, 1)
		assert.Equal(t, uint(1), cuerpo.Historias[0].ID)
		assert.NotNil(t, cuerpo.Historias[0].ROI)
		assert.NotNil(t, cuerpo.Historias[0].WSJF)
		assert.Equal(t, 1, servicio.consultas, "el handler debe alcanzar el servicio inyectado")
	})

	t.Run("sin token responde 401", func(t *testing.T) {
		t.Parallel()

		servicio := &servicioHistoriasPriorizada{}
		router := setupRouter(servicioQueFalla{}, servicio, servicioSprintsQueFalla{}, nil, nil, nil, nil, nil, nil, nil, nil, secretoDePruebaRouter)

		req, err := http.NewRequest(http.MethodGet, ruta, nil)
		require.NoError(t, err)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Zero(t, servicio.consultas)
	})

	t.Run("no miembro recibe 404 sin llegar al handler", func(t *testing.T) {
		t.Parallel()

		acceso := &servicioSinMembresia{}
		servicio := &servicioHistoriasPriorizada{}
		router := setupRouter(acceso, servicio, servicioSprintsQueFalla{}, nil, nil, nil, nil, nil, nil, nil, nil, secretoDePruebaRouter)

		w := pedirProtegidoEnRouter(t, router, http.MethodGet, ruta)

		require.Equal(t, http.StatusNotFound, w.Code)
		var cuerpo struct {
			Error string `json:"error"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cuerpo))
		assert.Equal(t, "PROYECTO_NO_ENCONTRADO", cuerpo.Error)
		assert.Zero(t, servicio.consultas, "el guard de membresia debe cortar antes del handler")
	})

	t.Run("id invalido responde 400 sin llegar al handler", func(t *testing.T) {
		t.Parallel()

		servicio := &servicioHistoriasPriorizada{}
		router := setupRouter(servicioQueFalla{}, servicio, servicioSprintsQueFalla{}, nil, nil, nil, nil, nil, nil, nil, nil, secretoDePruebaRouter)

		for _, invalida := range []string{
			"/api/proyectos/abc/backlog/priorizado",
			"/api/proyectos/0/backlog/priorizado",
		} {
			w := pedirProtegidoEnRouter(t, router, http.MethodGet, invalida)
			assert.Equal(t, http.StatusBadRequest, w.Code, "ruta: %s", invalida)
		}
		assert.Zero(t, servicio.consultas)
	})
}
