package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func nuevoRouterDeHistorias(t *testing.T) (*gin.Engine, *mockServicioHistorias) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	servicio := nuevoMockServicioHistorias(t)
	router := gin.New()
	RegistrarRutasHistorias(router, servicio)
	return router, servicio
}

func historiaDeEjemplo() *domain.Historia {
	creado := time.Date(2026, 3, 2, 15, 4, 5, 0, time.UTC)
	return &domain.Historia{
		ID:            10,
		ProyectoID:    1,
		Numero:        1,
		Codigo:        "US-01",
		Titulo:        "Como Product Builder quiero gestionar el backlog",
		Descripcion:   "Cubrir el alcance del proyecto",
		Prioridad:     domain.PrioridadAlta,
		Estado:        domain.EstadoHistoriaBacklog,
		CreadoEn:      creado,
		ActualizadoEn: creado,
	}
}

func decodificarHistoria(t *testing.T, w *httptest.ResponseRecorder) respuestaHistoria {
	t.Helper()
	var respuesta respuestaHistoria
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta), "cuerpo: %s", w.Body.String())
	return respuesta
}

func TestCrearHistoriaHandler(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre        string
		ruta          string
		cuerpo        string
		preparar      func(m *mockServicioHistorias)
		estadoHTTP    int
		errorEsperado string
		campoEsperado string
		escenario     string
	}{
		{
			nombre: "E-01 alta con todos los campos",
			ruta:   "/api/proyectos/1/historias",
			cuerpo: `{"titulo":"Como Product Builder quiero gestionar el backlog","descripcion":"Cubrir el alcance del proyecto","prioridad":"Alta","estado":"Backlog"}`,
			preparar: func(m *mockServicioHistorias) {
				m.On("Crear", mock.Anything, uint(1), domain.NuevaHistoria{
					Titulo:      "Como Product Builder quiero gestionar el backlog",
					Descripcion: "Cubrir el alcance del proyecto",
					Prioridad:   domain.PrioridadAlta,
					Estado:      domain.EstadoHistoriaBacklog,
				}).Return(historiaDeEjemplo(), nil)
			},
			estadoHTTP: http.StatusCreated,
			escenario:  "E-01",
		},
		{
			nombre: "E-02 sin prioridad ni estado",
			ruta:   "/api/proyectos/1/historias",
			cuerpo: `{"titulo":"Historia minima"}`,
			preparar: func(m *mockServicioHistorias) {
				m.On("Crear", mock.Anything, uint(1), domain.NuevaHistoria{Titulo: "Historia minima"}).
					Return(historiaDeEjemplo(), nil)
			},
			estadoHTTP: http.StatusCreated,
			escenario:  "E-02",
		},
		{
			nombre: "E-06 titulo vacio",
			ruta:   "/api/proyectos/1/historias",
			cuerpo: `{"titulo":""}`,
			preparar: func(m *mockServicioHistorias) {
				m.On("Crear", mock.Anything, uint(1), domain.NuevaHistoria{Titulo: ""}).
					Return(nil, domain.ErrTituloHistoriaObligatorio)
			},
			estadoHTTP:    http.StatusUnprocessableEntity,
			errorEsperado: codigoValidacion,
			campoEsperado: "titulo",
			escenario:     "E-06",
		},
		{
			nombre: "E-10 prioridad invalida",
			ruta:   "/api/proyectos/1/historias",
			cuerpo: `{"titulo":"Historia","prioridad":"Urgente"}`,
			preparar: func(m *mockServicioHistorias) {
				m.On("Crear", mock.Anything, uint(1), domain.NuevaHistoria{Titulo: "Historia", Prioridad: "Urgente"}).
					Return(nil, domain.ErrPrioridadHistoriaInvalida)
			},
			estadoHTTP:    http.StatusUnprocessableEntity,
			errorEsperado: codigoValidacion,
			campoEsperado: "prioridad",
			escenario:     "E-10",
		},
		{
			nombre: "E-11 estado invalido",
			ruta:   "/api/proyectos/1/historias",
			cuerpo: `{"titulo":"Historia","estado":"Cerrada"}`,
			preparar: func(m *mockServicioHistorias) {
				m.On("Crear", mock.Anything, uint(1), domain.NuevaHistoria{Titulo: "Historia", Estado: "Cerrada"}).
					Return(nil, domain.ErrEstadoHistoriaInvalido)
			},
			estadoHTTP:    http.StatusUnprocessableEntity,
			errorEsperado: codigoValidacion,
			campoEsperado: "estado",
			escenario:     "E-11",
		},
		{
			nombre: "E-14 codigo duplicado",
			ruta:   "/api/proyectos/1/historias",
			cuerpo: `{"titulo":"Historia"}`,
			preparar: func(m *mockServicioHistorias) {
				m.On("Crear", mock.Anything, uint(1), mock.Anything).Return(nil, domain.ErrCodigoHistoriaDuplicado)
			},
			estadoHTTP:    http.StatusConflict,
			errorEsperado: codigoHistoriaDuplicada,
			escenario:     "E-14",
		},
		{
			nombre: "E-15 proyecto inexistente",
			ruta:   "/api/proyectos/99/historias",
			cuerpo: `{"titulo":"Historia"}`,
			preparar: func(m *mockServicioHistorias) {
				m.On("Crear", mock.Anything, uint(99), mock.Anything).Return(nil, domain.ErrProyectoNoEncontrado)
			},
			estadoHTTP:    http.StatusNotFound,
			errorEsperado: codigoProyectoNoEncontrado,
			escenario:     "E-15",
		},
		{
			nombre:        "E-16 id de proyecto no numerico",
			ruta:          "/api/proyectos/abc/historias",
			cuerpo:        `{"titulo":"Historia"}`,
			preparar:      func(m *mockServicioHistorias) {},
			estadoHTTP:    http.StatusBadRequest,
			errorEsperado: codigoParametroInvalido,
			escenario:     "E-16",
		},
		{
			nombre:        "E-17 cuerpo JSON invalido",
			ruta:          "/api/proyectos/1/historias",
			cuerpo:        `{"titulo":`,
			preparar:      func(m *mockServicioHistorias) {},
			estadoHTTP:    http.StatusBadRequest,
			errorEsperado: codigoJSONInvalido,
			escenario:     "E-17",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			router, servicio := nuevoRouterDeHistorias(t)
			c.preparar(servicio)

			w := ejecutar(t, router, http.MethodPost, c.ruta, c.cuerpo)

			assert.Equal(t, c.estadoHTTP, w.Code, "escenario BDD %s", c.escenario)
			if c.errorEsperado != "" {
				respuesta := decodificarError(t, w)
				assert.Equal(t, c.errorEsperado, respuesta.Error, "escenario BDD %s", c.escenario)
				if c.campoEsperado != "" {
					assert.Equal(t, c.campoEsperado, respuesta.Campo, "escenario BDD %s", c.escenario)
				}
			}
			servicio.AssertExpectations(t)
		})
	}
}

func TestCrearHistoriaHandlerDevuelveLaRepresentacion(t *testing.T) {
	t.Parallel()

	router, servicio := nuevoRouterDeHistorias(t)
	servicio.On("Crear", mock.Anything, uint(1), mock.Anything).Return(historiaDeEjemplo(), nil)

	w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/historias", `{"titulo":"x"}`)

	require.Equal(t, http.StatusCreated, w.Code)
	var respuesta respuestaHistoria
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta))
	assert.Equal(t, uint(10), respuesta.ID)
	assert.Equal(t, "US-01", respuesta.Codigo)
	assert.Equal(t, domain.PrioridadAlta, respuesta.Prioridad)
	assert.Equal(t, "2026-03-02T15:04:05Z", respuesta.CreadoEn)
	assert.NotContains(t, w.Body.String(), "proyecto_id", "la ruta ya identifica el proyecto")
}

func TestListarHistoriasHandler(t *testing.T) {
	t.Parallel()

	t.Run("E-18 listado con historias", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeHistorias(t)
		servicio.On("Listar", mock.Anything, uint(1)).Return([]domain.Historia{*historiaDeEjemplo()}, nil)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/1/historias", nil)

		require.Equal(t, http.StatusOK, w.Code)
		var respuesta struct {
			Total     int                 `json:"total"`
			Historias []respuestaHistoria `json:"historias"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta))
		assert.Equal(t, 1, respuesta.Total)
		require.Len(t, respuesta.Historias, 1)
		servicio.AssertExpectations(t)
	})

	t.Run("E-19 listado vacio emite arreglo vacio", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeHistorias(t)
		servicio.On("Listar", mock.Anything, uint(1)).Return([]domain.Historia{}, nil)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/1/historias", nil)

		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"total":0,"historias":[]}`, w.Body.String())
	})

	t.Run("E-20 listado de proyecto inexistente", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeHistorias(t)
		servicio.On("Listar", mock.Anything, uint(99)).Return(nil, domain.ErrProyectoNoEncontrado)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/99/historias", nil)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, codigoProyectoNoEncontrado, decodificarError(t, w).Error)
	})
}

func TestActualizarHistoriaHandler(t *testing.T) {
	t.Parallel()

	t.Run("E-23 modificar la prioridad", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeHistorias(t)
		prioridad := domain.PrioridadAlta
		servicio.On("Actualizar", mock.Anything, uint(1), uint(5), domain.ActualizacionHistoria{
			Prioridad: &prioridad,
		}).Return(historiaDeEjemplo(), nil)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/historias/5", `{"prioridad":"Alta"}`)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "US-01", decodificarHistoria(t, w).Codigo)
		servicio.AssertExpectations(t)
	})

	t.Run("E-26 sin ningun campo informado", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeHistorias(t)
		servicio.On("Actualizar", mock.Anything, uint(1), uint(5), domain.ActualizacionHistoria{}).
			Return(nil, domain.ErrActualizacionHistoriaVacia)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/historias/5", `{}`)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		assert.Equal(t, codigoValidacion, decodificarError(t, w).Error)
	})

	t.Run("E-30 historia inexistente", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeHistorias(t)
		servicio.On("Actualizar", mock.Anything, uint(1), uint(5), mock.Anything).
			Return(nil, domain.ErrHistoriaNoEncontrada)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/historias/5", `{"estado":"Hecho"}`)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, codigoHistoriaNoEncontrada, decodificarError(t, w).Error)
	})

	t.Run("E-33 historiaId no numerico", func(t *testing.T) {
		t.Parallel()

		router, _ := nuevoRouterDeHistorias(t)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/historias/abc", `{"estado":"Hecho"}`)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, codigoParametroInvalido, decodificarError(t, w).Error)
	})

	t.Run("E-34 cuerpo JSON invalido", func(t *testing.T) {
		t.Parallel()

		router, _ := nuevoRouterDeHistorias(t)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/historias/5", `{"estado":`)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, codigoJSONInvalido, decodificarError(t, w).Error)
	})
}
