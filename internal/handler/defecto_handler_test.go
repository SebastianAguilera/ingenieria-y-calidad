package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func nuevoRouterDeDefectos(t *testing.T) (*gin.Engine, *mockServicioDefectos) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	servicio := nuevoMockServicioDefectos(t)
	router := gin.New()
	RegistrarRutasDefectos(router, servicio)
	return router, servicio
}

func defectoDeEjemplo() *domain.Defecto {
	creado := time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC)
	sprintResolucion := uint(30)
	return &domain.Defecto{
		ID:                 1,
		ProyectoID:         1,
		HistoriaID:         10,
		SprintDeteccionID:  20,
		SprintResolucionID: &sprintResolucion,
		Descripcion:        "Error en validación",
		Severidad:          domain.SeveridadMajor,
		Estado:             domain.EstadoDefectoResuelto,
		CreadoEn:           creado,
	}
}

func decodificarDefecto(t *testing.T, w *httptest.ResponseRecorder) respuestaDefecto {
	t.Helper()
	var respuesta respuestaDefecto
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta), "cuerpo: %s", w.Body.String())
	return respuesta
}

func TestCrearDefectoHandler(t *testing.T) {
	t.Parallel()

	cuerpoValido := `{"descripcion":"Error en validación","severidad":"major","historia_id":10,"sprint_deteccion_id":20}`

	casos := []struct {
		nombre        string
		ruta          string
		cuerpo        string
		preparar      func(m *mockServicioDefectos)
		estadoHTTP    int
		errorEsperado string
		campoEsperado string
		escenario     string
	}{
		{
			nombre: "BDD-01 alta valida",
			ruta:   "/api/proyectos/1/defectos",
			cuerpo: cuerpoValido,
			preparar: func(m *mockServicioDefectos) {
				m.On("Crear", mock.Anything, uint(1), domain.NuevoDefecto{
					Descripcion:       "Error en validación",
					Severidad:         domain.SeveridadMajor,
					HistoriaID:        10,
					SprintDeteccionID: 20,
				}).Return(defectoDeEjemplo(), nil)
			},
			estadoHTTP: http.StatusCreated,
			escenario:  "BDD-01",
		},
		{
			nombre: "BDD-02 severidad invalida",
			ruta:   "/api/proyectos/1/defectos",
			cuerpo: `{"descripcion":"Error","severidad":"urgente","historia_id":10,"sprint_deteccion_id":20}`,
			preparar: func(m *mockServicioDefectos) {
				m.On("Crear", mock.Anything, uint(1), mock.Anything).Return(nil, domain.ErrSeveridadDefectoInvalida)
			},
			estadoHTTP:    http.StatusUnprocessableEntity,
			errorEsperado: codigoValidacion,
			campoEsperado: "severidad",
			escenario:     "BDD-02",
		},
		{
			nombre: "BDD-03 descripcion vacia",
			ruta:   "/api/proyectos/1/defectos",
			cuerpo: `{"descripcion":"   ","severidad":"minor","historia_id":10,"sprint_deteccion_id":20}`,
			preparar: func(m *mockServicioDefectos) {
				m.On("Crear", mock.Anything, uint(1), mock.Anything).Return(nil, domain.ErrDescripcionDefectoObligatoria)
			},
			estadoHTTP:    http.StatusUnprocessableEntity,
			errorEsperado: codigoValidacion,
			campoEsperado: "descripcion",
			escenario:     "BDD-03",
		},
		{
			nombre: "BDD-04 historia inexistente",
			ruta:   "/api/proyectos/1/defectos",
			cuerpo: `{"descripcion":"Error","severidad":"major","historia_id":99,"sprint_deteccion_id":20}`,
			preparar: func(m *mockServicioDefectos) {
				m.On("Crear", mock.Anything, uint(1), mock.Anything).Return(nil, domain.ErrHistoriaNoEncontrada)
			},
			estadoHTTP:    http.StatusNotFound,
			errorEsperado: codigoHistoriaNoEncontrada,
			escenario:     "BDD-04",
		},
		{
			nombre: "BDD-05 sprint de deteccion inexistente",
			ruta:   "/api/proyectos/1/defectos",
			cuerpo: `{"descripcion":"Error","severidad":"major","historia_id":10,"sprint_deteccion_id":99}`,
			preparar: func(m *mockServicioDefectos) {
				m.On("Crear", mock.Anything, uint(1), mock.Anything).Return(nil, domain.ErrSprintNoEncontrado)
			},
			estadoHTTP:    http.StatusNotFound,
			errorEsperado: codigoSprintNoEncontrada,
			escenario:     "BDD-05",
		},
		{
			nombre: "BDD-06 usuario reportante inexistente",
			ruta:   "/api/proyectos/1/defectos",
			cuerpo: `{"descripcion":"Error","severidad":"major","historia_id":10,"sprint_deteccion_id":20,"reportado_por":99}`,
			preparar: func(m *mockServicioDefectos) {
				m.On("Crear", mock.Anything, uint(1), mock.Anything).Return(nil, domain.ErrUsuarioNoEncontrado)
			},
			estadoHTTP:    http.StatusNotFound,
			errorEsperado: codigoUsuarioNoEncontrado,
			escenario:     "BDD-06",
		},
		{
			nombre:        "id de proyecto no numerico",
			ruta:          "/api/proyectos/abc/defectos",
			cuerpo:        cuerpoValido,
			preparar:      func(m *mockServicioDefectos) {},
			estadoHTTP:    http.StatusBadRequest,
			errorEsperado: codigoParametroInvalido,
			escenario:     "BDD-01",
		},
		{
			nombre:        "JSON invalido",
			ruta:          "/api/proyectos/1/defectos",
			cuerpo:        `{"descripcion":`,
			preparar:      func(m *mockServicioDefectos) {},
			estadoHTTP:    http.StatusBadRequest,
			errorEsperado: codigoJSONInvalido,
			escenario:     "BDD-01",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			router, servicio := nuevoRouterDeDefectos(t)
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

func TestCrearDefectoHandlerDevuelveLaRepresentacion(t *testing.T) {
	t.Parallel()

	router, servicio := nuevoRouterDeDefectos(t)
	creado := defectoDeEjemplo()
	creado.Estado = domain.EstadoDefectoAbierto
	creado.SprintResolucionID = nil
	servicio.On("Crear", mock.Anything, uint(1), mock.Anything).Return(creado, nil)

	w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/defectos",
		`{"descripcion":"Error en validación","severidad":"major","historia_id":10,"sprint_deteccion_id":20}`)

	require.Equal(t, http.StatusCreated, w.Code)
	respuesta := decodificarDefecto(t, w)
	assert.Equal(t, uint(1), respuesta.ID)
	assert.Equal(t, domain.EstadoDefectoAbierto, respuesta.Estado)
	assert.Equal(t, domain.SeveridadMajor, respuesta.Severidad)
	assert.Equal(t, uint(10), respuesta.HistoriaID)
	assert.Equal(t, uint(20), respuesta.SprintDeteccionID)
	assert.Nil(t, respuesta.SprintResolucionID)
	assert.NotEmpty(t, respuesta.CreadoEn)
	assert.Nil(t, respuesta.ResueltoEn)
	assert.NotContains(t, w.Body.String(), "proyecto_id", "la ruta ya identifica el proyecto")
}

func TestActualizarDefectoHandler(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre        string
		ruta          string
		cuerpo        string
		preparar      func(m *mockServicioDefectos)
		estadoHTTP    int
		errorEsperado string
		campoEsperado string
		escenario     string
	}{
		{
			nombre: "BDD-08 resolver con sprint",
			ruta:   "/api/proyectos/1/defectos/1",
			cuerpo: `{"estado":"resolved","sprint_resolucion_id":30}`,
			preparar: func(m *mockServicioDefectos) {
				sprintID := uint(30)
				m.On("Actualizar", mock.Anything, uint(1), uint(1), domain.ActualizacionDefecto{
					Estado:             func() *domain.EstadoDefecto { e := domain.EstadoDefectoResuelto; return &e }(),
					SprintResolucionID: &sprintID,
				}).Return(defectoDeEjemplo(), nil)
			},
			estadoHTTP: http.StatusOK,
			escenario:  "BDD-08",
		},
		{
			nombre: "BDD-10 resolver sin sprint",
			ruta:   "/api/proyectos/1/defectos/1",
			cuerpo: `{"estado":"resolved"}`,
			preparar: func(m *mockServicioDefectos) {
				m.On("Actualizar", mock.Anything, uint(1), uint(1), mock.Anything).Return(nil, domain.ErrSprintResolucionRequerido)
			},
			estadoHTTP:    http.StatusUnprocessableEntity,
			errorEsperado: codigoValidacion,
			campoEsperado: "sprint_resolucion_id",
			escenario:     "BDD-10",
		},
		{
			nombre: "BDD-11 sprint de resolucion con estado abierto",
			ruta:   "/api/proyectos/1/defectos/1",
			cuerpo: `{"estado":"open","sprint_resolucion_id":30}`,
			preparar: func(m *mockServicioDefectos) {
				m.On("Actualizar", mock.Anything, uint(1), uint(1), mock.Anything).Return(nil, domain.ErrSprintResolucionNoPermitido)
			},
			estadoHTTP:    http.StatusUnprocessableEntity,
			errorEsperado: codigoValidacion,
			campoEsperado: "sprint_resolucion_id",
			escenario:     "BDD-11",
		},
		{
			nombre: "BDD-12 estado invalido",
			ruta:   "/api/proyectos/1/defectos/1",
			cuerpo: `{"estado":"fixed"}`,
			preparar: func(m *mockServicioDefectos) {
				m.On("Actualizar", mock.Anything, uint(1), uint(1), mock.Anything).Return(nil, domain.ErrEstadoDefectoInvalido)
			},
			estadoHTTP:    http.StatusBadRequest,
			errorEsperado: codigoEstadoInvalido,
			campoEsperado: "estado",
			escenario:     "BDD-12",
		},
		{
			nombre: "BDD-13 actualizacion vacia",
			ruta:   "/api/proyectos/1/defectos/1",
			cuerpo: `{}`,
			preparar: func(m *mockServicioDefectos) {
				m.On("Actualizar", mock.Anything, uint(1), uint(1), domain.ActualizacionDefecto{}).Return(nil, domain.ErrActualizacionDefectoVacia)
			},
			estadoHTTP:    http.StatusUnprocessableEntity,
			errorEsperado: codigoValidacion,
			escenario:     "BDD-13",
		},
		{
			nombre: "BDD-14 defecto inexistente",
			ruta:   "/api/proyectos/1/defectos/99",
			cuerpo: `{"estado":"in_progress"}`,
			preparar: func(m *mockServicioDefectos) {
				m.On("Actualizar", mock.Anything, uint(1), uint(99), mock.Anything).Return(nil, domain.ErrDefectoNoEncontrado)
			},
			estadoHTTP:    http.StatusNotFound,
			errorEsperado: codigoDefectoNoEncontrado,
			escenario:     "BDD-14",
		},
		{
			nombre:        "id de defecto no numerico",
			ruta:          "/api/proyectos/1/defectos/abc",
			cuerpo:        `{"estado":"open"}`,
			preparar:      func(m *mockServicioDefectos) {},
			estadoHTTP:    http.StatusBadRequest,
			errorEsperado: codigoParametroInvalido,
			escenario:     "BDD-14",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			router, servicio := nuevoRouterDeDefectos(t)
			c.preparar(servicio)

			w := ejecutar(t, router, http.MethodPatch, c.ruta, c.cuerpo)

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

func TestListarDefectosHandler(t *testing.T) {
	t.Parallel()

	t.Run("BDD-16 listado con defectos", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDeDefectos(t)
		servicio.On("Listar", mock.Anything, uint(1)).Return([]domain.Defecto{*defectoDeEjemplo()}, nil)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/1/defectos", nil)

		require.Equal(t, http.StatusOK, w.Code)
		var respuesta respuestaListadoDefectos
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta))
		assert.Equal(t, 1, respuesta.Total)
		require.Len(t, respuesta.Defectos, 1)
		assert.Equal(t, "Error en validación", respuesta.Defectos[0].Descripcion)
		servicio.AssertExpectations(t)
	})

	t.Run("BDD-17 listado vacio emite arreglo vacio", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDeDefectos(t)
		servicio.On("Listar", mock.Anything, uint(1)).Return([]domain.Defecto{}, nil)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/1/defectos", nil)

		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"total":0,"defectos":[]}`, w.Body.String())
	})

	t.Run("proyecto inexistente", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDeDefectos(t)
		servicio.On("Listar", mock.Anything, uint(99)).Return(nil, domain.ErrProyectoNoEncontrado)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/99/defectos", nil)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, codigoProyectoNoEncontrado, decodificarError(t, w).Error)
		servicio.AssertExpectations(t)
	})
}

// TestEditarDefectoDescripcionSeveridadHandler cubre el PATCH extendido de
// US-18. E-15 es una guarda de regresión (mapeo ya existente de US-09): pasa
// desde el RED. El resto de casos son RED.
// Escenarios: docs/features/US-18-defectos-por-sprint.feature.
func TestEditarDefectoDescripcionSeveridadHandler(t *testing.T) {
	t.Parallel()

	texto := func(v string) *string { return &v }
	severidad := func(v domain.SeveridadDefecto) *domain.SeveridadDefecto { return &v }

	casos := []struct {
		nombre        string
		cuerpo        string
		preparar      func(m *mockServicioDefectos)
		estadoHTTP    int
		errorEsperado string
		campoEsperado string
		escenario     string
	}{
		{
			nombre: "E-09 editar descripcion valida",
			cuerpo: `{"descripcion":"Error corregido"}`,
			preparar: func(m *mockServicioDefectos) {
				m.On("Actualizar", mock.Anything, uint(1), uint(1), domain.ActualizacionDefecto{
					Descripcion: texto("Error corregido"),
				}).Return(defectoDeEjemplo(), nil)
			},
			estadoHTTP: http.StatusOK,
			escenario:  "E-09",
		},
		{
			nombre: "E-10 editar severidad con valor de la lista",
			cuerpo: `{"severidad":"critical"}`,
			preparar: func(m *mockServicioDefectos) {
				m.On("Actualizar", mock.Anything, uint(1), uint(1), domain.ActualizacionDefecto{
					Severidad: severidad(domain.SeveridadCritical),
				}).Return(defectoDeEjemplo(), nil)
			},
			estadoHTTP: http.StatusOK,
			escenario:  "E-10",
		},
		{
			nombre: "E-13 descripcion de 1001 caracteres se mapea a 422 campo descripcion",
			cuerpo: `{"descripcion":"` + strings.Repeat("a", 1001) + `"}`,
			preparar: func(m *mockServicioDefectos) {
				m.On("Actualizar", mock.Anything, uint(1), uint(1), mock.Anything).Return(nil, domain.ErrDescripcionDefectoDemasiadoLarga)
			},
			estadoHTTP:    http.StatusUnprocessableEntity,
			errorEsperado: codigoValidacion,
			campoEsperado: "descripcion",
			escenario:     "E-13",
		},
		{
			nombre: "E-15 [GUARDA DE REGRESION, no RED] severidad fuera de la lista se mapea a 422 campo severidad",
			cuerpo: `{"severidad":"Major"}`,
			preparar: func(m *mockServicioDefectos) {
				m.On("Actualizar", mock.Anything, uint(1), uint(1), mock.Anything).Return(nil, domain.ErrSeveridadDefectoInvalida)
			},
			estadoHTTP:    http.StatusUnprocessableEntity,
			errorEsperado: codigoValidacion,
			campoEsperado: "severidad",
			escenario:     "E-15",
		},
		{
			nombre: "E-17 sprint_deteccion_id en el body se rechaza con 422 campo sprint_deteccion_id",
			cuerpo: `{"sprint_deteccion_id":2,"severidad":"major"}`,
			preparar: func(m *mockServicioDefectos) {
				m.On("Actualizar", mock.Anything, uint(1), uint(1), domain.ActualizacionDefecto{
					Severidad:               severidad(domain.SeveridadMajor),
					SprintDeteccionPresente: true,
				}).Return(nil, domain.ErrSprintDeteccionNoModificable)
			},
			estadoHTTP:    http.StatusUnprocessableEntity,
			errorEsperado: codigoValidacion,
			campoEsperado: "sprint_deteccion_id",
			escenario:     "E-17",
		},
		{
			nombre: "E-21 sprint_deteccion_id con valor actual tambien se rechaza",
			cuerpo: `{"sprint_deteccion_id":1}`,
			preparar: func(m *mockServicioDefectos) {
				m.On("Actualizar", mock.Anything, uint(1), uint(1), domain.ActualizacionDefecto{
					SprintDeteccionPresente: true,
				}).Return(nil, domain.ErrSprintDeteccionNoModificable)
			},
			estadoHTTP:    http.StatusUnprocessableEntity,
			errorEsperado: codigoValidacion,
			campoEsperado: "sprint_deteccion_id",
			escenario:     "E-21",
		},
		{
			nombre: "E-22 sprint_deteccion_id null tambien se rechaza (la clave esta presente)",
			cuerpo: `{"sprint_deteccion_id":null}`,
			preparar: func(m *mockServicioDefectos) {
				m.On("Actualizar", mock.Anything, uint(1), uint(1), domain.ActualizacionDefecto{
					SprintDeteccionPresente: true,
				}).Return(nil, domain.ErrSprintDeteccionNoModificable)
			},
			estadoHTTP:    http.StatusUnprocessableEntity,
			errorEsperado: codigoValidacion,
			campoEsperado: "sprint_deteccion_id",
			escenario:     "E-22",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			router, servicio := nuevoRouterDeDefectos(t)
			c.preparar(servicio)

			w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/defectos/1", c.cuerpo)

			assert.Equal(t, c.estadoHTTP, w.Code, "escenario BDD %s", c.escenario)
			if c.errorEsperado != "" {
				respuesta := decodificarError(t, w)
				assert.Equal(t, c.errorEsperado, respuesta.Error, "escenario BDD %s", c.escenario)
				assert.Equal(t, c.campoEsperado, respuesta.Campo, "escenario BDD %s", c.escenario)
				if c.escenario == "E-17" {
					assert.Equal(t, domain.ErrSprintDeteccionNoModificable.Error(), respuesta.Mensaje, "escenario BDD E-17: el mensaje debe nombrar el campo")
				}
			}
			servicio.AssertExpectations(t)
		})
	}
}

// TestListarDefectosPorSprintHandler cubre GET
// /api/proyectos/:id/sprints/:sprintId/defectos?tipo= de US-18.
// Escenarios: docs/features/US-18-defectos-por-sprint.feature (E-01 a E-06).
func TestListarDefectosPorSprintHandler(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre        string
		ruta          string
		preparar      func(m *mockServicioDefectos)
		estadoHTTP    int
		errorEsperado string
		campoEsperado string
		escenario     string
	}{
		{
			nombre: "E-01 listado por deteccion",
			ruta:   "/api/proyectos/1/sprints/1/defectos?tipo=deteccion",
			preparar: func(m *mockServicioDefectos) {
				m.On("ListarPorSprint", mock.Anything, uint(1), uint(1), domain.TipoListadoDeteccion).
					Return([]domain.Defecto{*defectoDeEjemplo()}, nil)
			},
			estadoHTTP: http.StatusOK,
			escenario:  "E-01",
		},
		{
			nombre: "E-03 sprint sin defectos responde arreglo vacio",
			ruta:   "/api/proyectos/1/sprints/1/defectos?tipo=resolucion",
			preparar: func(m *mockServicioDefectos) {
				m.On("ListarPorSprint", mock.Anything, uint(1), uint(1), domain.TipoListadoResolucion).
					Return([]domain.Defecto{}, nil)
			},
			estadoHTTP: http.StatusOK,
			escenario:  "E-03",
		},
		{
			nombre: "E-04 falta tipo responde 422 campo tipo",
			ruta:   "/api/proyectos/1/sprints/1/defectos",
			preparar: func(m *mockServicioDefectos) {
				m.On("ListarPorSprint", mock.Anything, uint(1), uint(1), domain.TipoListadoDefecto("")).
					Return(nil, domain.ErrTipoListadoDefectoInvalido)
			},
			estadoHTTP:    http.StatusUnprocessableEntity,
			errorEsperado: codigoValidacion,
			campoEsperado: "tipo",
			escenario:     "E-04",
		},
		{
			nombre: "E-05 tipo con mayusculas responde 422 campo tipo",
			ruta:   "/api/proyectos/1/sprints/1/defectos?tipo=Deteccion",
			preparar: func(m *mockServicioDefectos) {
				m.On("ListarPorSprint", mock.Anything, uint(1), uint(1), domain.TipoListadoDefecto("Deteccion")).
					Return(nil, domain.ErrTipoListadoDefectoInvalido)
			},
			estadoHTTP:    http.StatusUnprocessableEntity,
			errorEsperado: codigoValidacion,
			campoEsperado: "tipo",
			escenario:     "E-05",
		},
		{
			nombre: "E-06 sprint inexistente responde 404",
			ruta:   "/api/proyectos/1/sprints/999/defectos?tipo=deteccion",
			preparar: func(m *mockServicioDefectos) {
				m.On("ListarPorSprint", mock.Anything, uint(1), uint(999), domain.TipoListadoDeteccion).
					Return(nil, domain.ErrSprintNoEncontrado)
			},
			estadoHTTP:    http.StatusNotFound,
			errorEsperado: codigoSprintNoEncontrada,
			escenario:     "E-06",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			router, servicio := nuevoRouterDeDefectos(t)
			c.preparar(servicio)

			w := ejecutar(t, router, http.MethodGet, c.ruta, nil)

			assert.Equal(t, c.estadoHTTP, w.Code, "escenario BDD %s", c.escenario)
			if c.errorEsperado != "" {
				respuesta := decodificarError(t, w)
				assert.Equal(t, c.errorEsperado, respuesta.Error, "escenario BDD %s", c.escenario)
				if c.campoEsperado != "" {
					assert.Equal(t, c.campoEsperado, respuesta.Campo, "escenario BDD %s", c.escenario)
				}
			} else {
				assert.Contains(t, w.Body.String(), `"defectos":[`, "escenario BDD %s: el listado debe ser arreglo, nunca null", c.escenario)
			}
			servicio.AssertExpectations(t)
		})
	}
}

// TestEditarDefectoBordesHandler cubre por HTTP los escenarios de borde del
// PATCH de US-18 (E-11, E-12, E-14, E-16, E-18, E-19). Escenarios:
// docs/features/US-18-defectos-por-sprint.feature.
// Guardas de regresión: el handler solo traduce; la lógica de límites vive en
// el service (ver TestActualizarDefectoDescripcionYSeveridad).
func TestEditarDefectoBordesHandler(t *testing.T) {
	t.Parallel()

	texto := func(v string) *string { return &v }
	sprintResolucion := uint(2)

	t.Run("E-11 [GUARDA DE REGRESION] descripcion de exactamente 1000 caracteres responde 200", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDeDefectos(t)
		descripcion := strings.Repeat("a", 1000)
		editado := defectoDeEjemplo()
		editado.Descripcion = descripcion
		servicio.On("Actualizar", mock.Anything, uint(1), uint(1), domain.ActualizacionDefecto{
			Descripcion: texto(descripcion),
		}).Return(editado, nil)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/defectos/1",
			`{"descripcion":"`+descripcion+`"}`)

		require.Equal(t, http.StatusOK, w.Code, "escenario BDD E-11")
		assert.Equal(t, descripcion, decodificarDefecto(t, w).Descripcion, "escenario BDD E-11")
		servicio.AssertExpectations(t)
	})

	t.Run("E-12 [GUARDA DE REGRESION] descripcion de 1000 runas multibyte responde 200", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDeDefectos(t)
		descripcion := strings.Repeat("ñ", 1000)
		editado := defectoDeEjemplo()
		editado.Descripcion = descripcion
		servicio.On("Actualizar", mock.Anything, uint(1), uint(1), domain.ActualizacionDefecto{
			Descripcion: texto(descripcion),
		}).Return(editado, nil)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/defectos/1",
			`{"descripcion":"`+descripcion+`"}`)

		require.Equal(t, http.StatusOK, w.Code, "escenario BDD E-12")
		assert.Equal(t, descripcion, decodificarDefecto(t, w).Descripcion, "escenario BDD E-12")
		servicio.AssertExpectations(t)
	})

	t.Run("E-14 [GUARDA DE REGRESION] descripcion solo con espacios responde 422 campo descripcion", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDeDefectos(t)
		servicio.On("Actualizar", mock.Anything, uint(1), uint(1), domain.ActualizacionDefecto{
			Descripcion: texto("   "),
		}).Return(nil, domain.ErrDescripcionDefectoObligatoria)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/defectos/1", `{"descripcion":"   "}`)

		require.Equal(t, http.StatusUnprocessableEntity, w.Code, "escenario BDD E-14")
		respuesta := decodificarError(t, w)
		assert.Equal(t, codigoValidacion, respuesta.Error, "escenario BDD E-14")
		assert.Equal(t, "descripcion", respuesta.Campo, "escenario BDD E-14")
		servicio.AssertExpectations(t)
	})

	t.Run("E-16 [GUARDA DE REGRESION] editar severidad conserva estado y sprint de resolucion", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDeDefectos(t)
		severidad := domain.SeveridadBlocker
		editado := defectoDeEjemplo()
		editado.Severidad = severidad
		editado.Estado = domain.EstadoDefectoResuelto
		resolucion := sprintResolucion
		editado.SprintResolucionID = &resolucion
		servicio.On("Actualizar", mock.Anything, uint(1), uint(1), domain.ActualizacionDefecto{
			Severidad: &severidad,
		}).Return(editado, nil)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/defectos/1", `{"severidad":"blocker"}`)

		require.Equal(t, http.StatusOK, w.Code, "escenario BDD E-16")
		respuesta := decodificarDefecto(t, w)
		assert.Equal(t, domain.SeveridadBlocker, respuesta.Severidad, "escenario BDD E-16")
		assert.Equal(t, domain.EstadoDefectoResuelto, respuesta.Estado, "escenario BDD E-16: el estado no cambia")
		require.NotNil(t, respuesta.SprintResolucionID, "escenario BDD E-16")
		assert.Equal(t, sprintResolucion, *respuesta.SprintResolucionID, "escenario BDD E-16: el sprint de resolucion no cambia")
		servicio.AssertExpectations(t)
	})

	t.Run("E-18 [GUARDA DE REGRESION] defecto inexistente responde 404 DEFECTO_NO_ENCONTRADO", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDeDefectos(t)
		servicio.On("Actualizar", mock.Anything, uint(1), uint(999), mock.Anything).Return(nil, domain.ErrDefectoNoEncontrado)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/defectos/999", `{"severidad":"major"}`)

		require.Equal(t, http.StatusNotFound, w.Code, "escenario BDD E-18")
		assert.Equal(t, codigoDefectoNoEncontrado, decodificarError(t, w).Error, "escenario BDD E-18")
		servicio.AssertExpectations(t)
	})

	t.Run("E-19 [GUARDA DE REGRESION] defecto de otro proyecto responde 404 DEFECTO_NO_ENCONTRADO", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDeDefectos(t)
		servicio.On("Actualizar", mock.Anything, uint(1), uint(7), mock.Anything).Return(nil, domain.ErrDefectoNoEncontrado)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/defectos/7", `{"descripcion":"Intento"}`)

		require.Equal(t, http.StatusNotFound, w.Code, "escenario BDD E-19")
		assert.Equal(t, codigoDefectoNoEncontrado, decodificarError(t, w).Error, "escenario BDD E-19")
		servicio.AssertExpectations(t)
	})
}
