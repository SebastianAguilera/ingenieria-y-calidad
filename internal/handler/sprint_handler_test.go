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

func nuevoRouterDeSprints(t *testing.T) (*gin.Engine, *mockServicioSprints) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	servicio := nuevoMockServicioSprints(t)
	router := gin.New()
	RegistrarRutasSprints(router, servicio)
	return router, servicio
}

func sprintDeEjemplo() *domain.Sprint {
	creado := time.Date(2026, 10, 1, 9, 15, 0, 0, time.UTC)
	return &domain.Sprint{
		ID:            1,
		ProyectoID:    1,
		Identificador: "Sprint 1",
		SprintGoal:    "Entregar el MVP del motor de métricas",
		Estado:        domain.EstadoSprintPlanificado,
		FechaInicio:   time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC),
		FechaFin:      time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC),
		CreadoEn:      creado,
		ActualizadoEn: creado,
	}
}

func decodificarSprint(t *testing.T, w *httptest.ResponseRecorder) respuestaSprint {
	t.Helper()
	var respuesta respuestaSprint
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta), "cuerpo: %s", w.Body.String())
	return respuesta
}

func TestCrearSprintHandler(t *testing.T) {
	t.Parallel()

	cuerpoValido := `{"identificador":"Sprint 1","sprint_goal":"Entregar el MVP","fecha_inicio":"2026-10-06","fecha_fin":"2026-10-20"}`

	casos := []struct {
		nombre        string
		ruta          string
		cuerpo        string
		preparar      func(m *mockServicioSprints)
		estadoHTTP    int
		errorEsperado string
		campoEsperado string
		escenario     string
	}{
		{
			nombre: "E-01 alta con todos los campos",
			ruta:   "/api/proyectos/1/sprints",
			cuerpo: cuerpoValido,
			preparar: func(m *mockServicioSprints) {
				m.On("Crear", mock.Anything, uint(1), domain.NuevoSprint{
					Identificador: "Sprint 1",
					SprintGoal:    "Entregar el MVP",
					FechaInicio:   time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC),
					FechaFin:      time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC),
				}).Return(sprintDeEjemplo(), nil)
			},
			estadoHTTP: http.StatusCreated,
			escenario:  "E-01",
		},
		{
			nombre: "E-03 identificador vacio",
			ruta:   "/api/proyectos/1/sprints",
			cuerpo: `{"identificador":"","sprint_goal":"Goal","fecha_inicio":"2026-10-06","fecha_fin":"2026-10-20"}`,
			preparar: func(m *mockServicioSprints) {
				m.On("Crear", mock.Anything, uint(1), mock.Anything).Return(nil, domain.ErrIdentificadorSprintObligatorio)
			},
			estadoHTTP:    http.StatusUnprocessableEntity,
			errorEsperado: codigoValidacion,
			campoEsperado: "identificador",
			escenario:     "E-03",
		},
		{
			nombre: "E-05 sprint goal vacio",
			ruta:   "/api/proyectos/1/sprints",
			cuerpo: `{"identificador":"Sprint 1","sprint_goal":"","fecha_inicio":"2026-10-06","fecha_fin":"2026-10-20"}`,
			preparar: func(m *mockServicioSprints) {
				m.On("Crear", mock.Anything, uint(1), mock.Anything).Return(nil, domain.ErrSprintGoalObligatorio)
			},
			estadoHTTP:    http.StatusUnprocessableEntity,
			errorEsperado: codigoValidacion,
			campoEsperado: "sprint_goal",
			escenario:     "E-05",
		},
		{
			nombre: "E-07 identificador de 51 caracteres",
			ruta:   "/api/proyectos/1/sprints",
			cuerpo: `{"identificador":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","sprint_goal":"Goal","fecha_inicio":"2026-10-06","fecha_fin":"2026-10-20"}`,
			preparar: func(m *mockServicioSprints) {
				m.On("Crear", mock.Anything, uint(1), mock.Anything).Return(nil, domain.ErrIdentificadorSprintDemasiadoLargo)
			},
			estadoHTTP:    http.StatusUnprocessableEntity,
			errorEsperado: codigoValidacion,
			campoEsperado: "identificador",
			escenario:     "E-07",
		},
		{
			nombre: "E-09 sprint goal de 301 caracteres",
			ruta:   "/api/proyectos/1/sprints",
			cuerpo: `{"identificador":"Sprint 1","sprint_goal":"Goal","fecha_inicio":"2026-10-06","fecha_fin":"2026-10-20"}`,
			preparar: func(m *mockServicioSprints) {
				m.On("Crear", mock.Anything, uint(1), mock.Anything).Return(nil, domain.ErrSprintGoalDemasiadoLargo)
			},
			estadoHTTP:    http.StatusUnprocessableEntity,
			errorEsperado: codigoValidacion,
			campoEsperado: "sprint_goal",
			escenario:     "E-09",
		},
		{
			nombre: "E-10 fecha de fin anterior a la de inicio",
			ruta:   "/api/proyectos/1/sprints",
			cuerpo: `{"identificador":"Sprint 1","sprint_goal":"Goal","fecha_inicio":"2026-10-06","fecha_fin":"2026-10-05"}`,
			preparar: func(m *mockServicioSprints) {
				m.On("Crear", mock.Anything, uint(1), mock.Anything).Return(nil, domain.ErrSprintFechaFinAnteriorAlInicio)
			},
			estadoHTTP:    http.StatusUnprocessableEntity,
			errorEsperado: codigoValidacion,
			campoEsperado: "fecha_fin",
			escenario:     "E-10",
		},
		{
			nombre: "E-12 identificador duplicado",
			ruta:   "/api/proyectos/1/sprints",
			cuerpo: cuerpoValido,
			preparar: func(m *mockServicioSprints) {
				m.On("Crear", mock.Anything, uint(1), mock.Anything).Return(nil, domain.ErrIdentificadorSprintDuplicado)
			},
			estadoHTTP:    http.StatusConflict,
			errorEsperado: codigoSprintDuplicado,
			campoEsperado: "identificador",
			escenario:     "E-12",
		},
		{
			nombre: "E-14 fechas solapadas",
			ruta:   "/api/proyectos/1/sprints",
			cuerpo: cuerpoValido,
			preparar: func(m *mockServicioSprints) {
				m.On("Crear", mock.Anything, uint(1), mock.Anything).Return(nil, domain.ErrSprintFechasSolapadas)
			},
			estadoHTTP:    http.StatusConflict,
			errorEsperado: codigoSprintFechasSolapadas,
			escenario:     "E-14",
		},
		{
			nombre: "E-17 proyecto inexistente",
			ruta:   "/api/proyectos/99/sprints",
			cuerpo: cuerpoValido,
			preparar: func(m *mockServicioSprints) {
				m.On("Crear", mock.Anything, uint(99), mock.Anything).Return(nil, domain.ErrProyectoNoEncontrado)
			},
			estadoHTTP:    http.StatusNotFound,
			errorEsperado: codigoProyectoNoEncontrado,
			escenario:     "E-17",
		},
		{
			nombre:        "E-18 fecha_inicio ausente",
			ruta:          "/api/proyectos/1/sprints",
			cuerpo:        `{"identificador":"Sprint 1","sprint_goal":"Goal","fecha_fin":"2026-10-20"}`,
			preparar:      func(m *mockServicioSprints) {},
			estadoHTTP:    http.StatusBadRequest,
			errorEsperado: codigoParametroInvalido,
			campoEsperado: "fecha_inicio",
			escenario:     "E-18",
		},
		{
			nombre:        "E-18 fecha_inicio mal formada",
			ruta:          "/api/proyectos/1/sprints",
			cuerpo:        `{"identificador":"Sprint 1","sprint_goal":"Goal","fecha_inicio":"06-10-2026","fecha_fin":"2026-10-20"}`,
			preparar:      func(m *mockServicioSprints) {},
			estadoHTTP:    http.StatusBadRequest,
			errorEsperado: codigoParametroInvalido,
			campoEsperado: "fecha_inicio",
			escenario:     "E-18",
		},
		{
			nombre:        "E-18 fecha_fin mal formada",
			ruta:          "/api/proyectos/1/sprints",
			cuerpo:        `{"identificador":"Sprint 1","sprint_goal":"Goal","fecha_inicio":"2026-10-06","fecha_fin":"20/10/2026"}`,
			preparar:      func(m *mockServicioSprints) {},
			estadoHTTP:    http.StatusBadRequest,
			errorEsperado: codigoParametroInvalido,
			campoEsperado: "fecha_fin",
			escenario:     "E-18",
		},
		{
			nombre:        "E-19 id de proyecto no numerico",
			ruta:          "/api/proyectos/abc/sprints",
			cuerpo:        cuerpoValido,
			preparar:      func(m *mockServicioSprints) {},
			estadoHTTP:    http.StatusBadRequest,
			errorEsperado: codigoParametroInvalido,
			escenario:     "E-19",
		},
		{
			nombre:        "E-20 cuerpo JSON invalido",
			ruta:          "/api/proyectos/1/sprints",
			cuerpo:        `{"identificador":`,
			preparar:      func(m *mockServicioSprints) {},
			estadoHTTP:    http.StatusBadRequest,
			errorEsperado: codigoJSONInvalido,
			escenario:     "E-20",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			router, servicio := nuevoRouterDeSprints(t)
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

func TestCrearSprintHandlerDevuelveLaRepresentacion(t *testing.T) {
	t.Parallel()

	router, servicio := nuevoRouterDeSprints(t)
	servicio.On("Crear", mock.Anything, uint(1), mock.Anything).Return(sprintDeEjemplo(), nil)

	w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/sprints",
		`{"identificador":"Sprint 1","sprint_goal":"Entregar el MVP","fecha_inicio":"2026-10-06","fecha_fin":"2026-10-20"}`)

	require.Equal(t, http.StatusCreated, w.Code)
	respuesta := decodificarSprint(t, w)
	assert.Equal(t, uint(1), respuesta.ID)
	assert.Equal(t, "Sprint 1", respuesta.Identificador)
	assert.Equal(t, domain.EstadoSprintPlanificado, respuesta.Estado)
	assert.Equal(t, "2026-10-06", respuesta.FechaInicio)
	assert.Equal(t, "2026-10-20", respuesta.FechaFin)
	assert.Equal(t, "2026-10-01T09:15:00Z", respuesta.CreadoEn, "escenario BDD E-22")
	assert.NotContains(t, w.Body.String(), "proyecto_id", "la ruta ya identifica el proyecto")
}

func TestListarSprintsHandler(t *testing.T) {
	t.Parallel()

	t.Run("E-25 listado con sprints", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeSprints(t)
		servicio.On("Listar", mock.Anything, uint(1)).Return([]domain.Sprint{*sprintDeEjemplo()}, nil)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/1/sprints", nil)

		require.Equal(t, http.StatusOK, w.Code)
		var respuesta respuestaListadoSprints
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta))
		assert.Equal(t, 1, respuesta.Total)
		require.Len(t, respuesta.Sprints, 1)
		assert.Equal(t, "Sprint 1", respuesta.Sprints[0].Identificador)
		servicio.AssertExpectations(t)
	})

	t.Run("E-26 listado vacio emite arreglo vacio", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeSprints(t)
		servicio.On("Listar", mock.Anything, uint(1)).Return([]domain.Sprint{}, nil)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/1/sprints", nil)

		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"total":0,"sprints":[]}`, w.Body.String())
	})

	t.Run("E-27 listado de proyecto inexistente", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeSprints(t)
		servicio.On("Listar", mock.Anything, uint(99)).Return(nil, domain.ErrProyectoNoEncontrado)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/99/sprints", nil)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, codigoProyectoNoEncontrado, decodificarError(t, w).Error)
	})

	t.Run("E-28 id de proyecto no numerico", func(t *testing.T) {
		t.Parallel()

		router, _ := nuevoRouterDeSprints(t)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/abc/sprints", nil)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, codigoParametroInvalido, decodificarError(t, w).Error)
	})
}

func TestCambiarEstadoSprintHandler(t *testing.T) {
	t.Parallel()

	t.Run("E-29 activar un sprint", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeSprints(t)
		activo := sprintDeEjemplo()
		activo.Estado = domain.EstadoSprintActivo
		servicio.On("CambiarEstado", mock.Anything, uint(1), uint(1), domain.EstadoSprintActivo).
			Return(activo, nil)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/sprints/1/estado", `{"estado":"Activo"}`)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, domain.EstadoSprintActivo, decodificarSprint(t, w).Estado)
		servicio.AssertExpectations(t)
	})

	t.Run("E-30 cerrar un sprint", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeSprints(t)
		cerrado := sprintDeEjemplo()
		cerrado.Estado = domain.EstadoSprintCerrado
		servicio.On("CambiarEstado", mock.Anything, uint(1), uint(1), domain.EstadoSprintCerrado).
			Return(cerrado, nil)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/sprints/1/estado", `{"estado":"Cerrado"}`)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, domain.EstadoSprintCerrado, decodificarSprint(t, w).Estado)
	})

	t.Run("E-31 transicion invalida", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeSprints(t)
		servicio.On("CambiarEstado", mock.Anything, uint(1), uint(1), domain.EstadoSprintPlanificado).
			Return(nil, domain.ErrTransicionSprintInvalida)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/sprints/1/estado", `{"estado":"Planificado"}`)

		assert.Equal(t, http.StatusConflict, w.Code)
		respuesta := decodificarError(t, w)
		assert.Equal(t, codigoTransicionSprintInvalida, respuesta.Error)
		assert.Empty(t, respuesta.Campo)
	})

	t.Run("E-32 literal de estado fuera del conjunto", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeSprints(t)
		servicio.On("CambiarEstado", mock.Anything, uint(1), uint(1), domain.EstadoSprint("Terminado")).
			Return(nil, domain.ErrEstadoSprintInvalido)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/sprints/1/estado", `{"estado":"Terminado"}`)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		respuesta := decodificarError(t, w)
		assert.Equal(t, codigoEstadoInvalido, respuesta.Error)
		assert.Equal(t, "estado", respuesta.Campo)
	})

	t.Run("E-33 cuerpo JSON invalido", func(t *testing.T) {
		t.Parallel()

		router, _ := nuevoRouterDeSprints(t)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/sprints/1/estado", `{"estado":`)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, codigoJSONInvalido, decodificarError(t, w).Error)
	})

	t.Run("E-34 sprint inexistente", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeSprints(t)
		servicio.On("CambiarEstado", mock.Anything, uint(1), uint(5), domain.EstadoSprintActivo).
			Return(nil, domain.ErrSprintNoEncontrado)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/sprints/5/estado", `{"estado":"Activo"}`)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, codigoSprintNoEncontrada, decodificarError(t, w).Error)
	})

	t.Run("E-36 proyecto inexistente", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeSprints(t)
		servicio.On("CambiarEstado", mock.Anything, uint(99), uint(1), domain.EstadoSprintActivo).
			Return(nil, domain.ErrProyectoNoEncontrado)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/99/sprints/1/estado", `{"estado":"Activo"}`)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, codigoProyectoNoEncontrado, decodificarError(t, w).Error)
	})

	t.Run("E-37 sprintId no numerico", func(t *testing.T) {
		t.Parallel()

		router, _ := nuevoRouterDeSprints(t)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/sprints/abc/estado", `{"estado":"Activo"}`)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, codigoParametroInvalido, decodificarError(t, w).Error)
	})

	t.Run("E-38 ya existe otro sprint activo", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeSprints(t)
		servicio.On("CambiarEstado", mock.Anything, uint(1), uint(2), domain.EstadoSprintActivo).
			Return(nil, domain.ErrSprintActivoExistente)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/sprints/2/estado", `{"estado":"Activo"}`)

		assert.Equal(t, http.StatusConflict, w.Code)
		respuesta := decodificarError(t, w)
		assert.Equal(t, codigoSprintActivoExistente, respuesta.Error)
		assert.Empty(t, respuesta.Campo)
	})
}
