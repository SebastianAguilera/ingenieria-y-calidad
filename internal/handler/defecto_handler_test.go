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
