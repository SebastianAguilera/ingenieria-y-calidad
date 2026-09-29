package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

var (
	fechaInicioHTTP = time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)
	fechaFinHTTP    = time.Date(2026, 12, 18, 0, 0, 0, 0, time.UTC)
)

func nuevoRouterDePrueba(t *testing.T) (*gin.Engine, *mockServicioProyectos) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	servicio := nuevoMockServicioProyectos(t)
	router := gin.New()
	RegistrarRutasProyectos(router, servicio)
	return router, servicio
}

func ejecutar(t *testing.T, router *gin.Engine, metodo, ruta string, cuerpo any) *httptest.ResponseRecorder {
	t.Helper()

	var lector *bytes.Reader
	if cuerpo == nil {
		lector = bytes.NewReader(nil)
	} else if crudo, ok := cuerpo.(string); ok {
		lector = bytes.NewReader([]byte(crudo))
	} else {
		datos, err := json.Marshal(cuerpo)
		require.NoError(t, err)
		lector = bytes.NewReader(datos)
	}

	req := httptest.NewRequest(metodo, ruta, lector)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func decodificarError(t *testing.T, w *httptest.ResponseRecorder) respuestaError {
	t.Helper()
	var respuesta respuestaError
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta), "cuerpo: %s", w.Body.String())
	return respuesta
}

func TestHealthNoAfectado(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	RegistrarRutasProyectos(router, nuevoMockServicioProyectos(t))

	w := ejecutar(t, router, http.MethodGet, "/health", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func TestCrearProyecto(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		cuerpo      any
		preparar    func(s *mockServicioProyectos)
		codigo      int
		codigoError string
		mensaje     string
		campo       string
		escenario   string
	}{
		{
			nombre: "E-01 alta valida responde 201 con el proyecto persistido",
			cuerpo: map[string]any{
				"nombre":       "Software Metrics & Estimation",
				"fecha_inicio": "2026-09-28",
				"fecha_fin":    "2026-12-18",
				"integrantes": []map[string]string{
					{"nombre": "Aguilera Sebastián", "email": "sebas.aguilera@utn.edu.ar"},
				},
			},
			preparar: func(s *mockServicioProyectos) {
				s.On("Crear", mock.Anything, mock.Anything).Return(
					proyectoRespuesta(1, "Software Metrics & Estimation", &fechaFinHTTP, domain.EstadoActivo,
						[]domain.Integrante{{ID: 1, Nombre: "Aguilera Sebastián", Email: "sebas.aguilera@utn.edu.ar"}}),
					nil,
				)
			},
			codigo:    http.StatusCreated,
			escenario: "E-01",
		},
		{
			nombre: "E-03 alta sin integrantes responde 201 con lista vacia",
			cuerpo: map[string]any{
				"nombre":       "Proyecto sin equipo",
				"fecha_inicio": "2026-09-28",
				"integrantes":  []map[string]string{},
			},
			preparar: func(s *mockServicioProyectos) {
				s.On("Crear", mock.Anything, mock.Anything).Return(
					proyectoRespuesta(1, "Proyecto sin equipo", nil, domain.EstadoActivo, []domain.Integrante{}), nil)
			},
			codigo:    http.StatusCreated,
			escenario: "E-03",
		},
		{
			nombre:    "E-20 json malformado responde 400",
			cuerpo:    `{"nombre": "Proyecto",`,
			codigo:    http.StatusBadRequest,
			escenario: "E-20",
		},
		{
			nombre:    "E-20 body vacio responde 400",
			codigo:    http.StatusBadRequest,
			escenario: "E-20",
		},
		{
			nombre: "E-22 fecha en formato dd/mm/aaaa responde 400",
			cuerpo: map[string]any{
				"nombre":       "Proyecto",
				"fecha_inicio": "28/09/2026",
			},
			codigo:      http.StatusBadRequest,
			codigoError: "PARAMETRO_INVALIDO",
			campo:       "fecha_inicio",
			escenario:   "E-22",
		},
		{
			nombre: "E-11 fecha de fin anterior a la de inicio responde 422",
			cuerpo: map[string]any{
				"nombre":       "Proyecto invertido",
				"fecha_inicio": "2026-09-28",
				"fecha_fin":    "2026-09-01",
			},
			preparar: func(s *mockServicioProyectos) {
				s.On("Crear", mock.Anything, mock.Anything).Return(nil, domain.ErrFechaFinAnteriorAlInicio)
			},
			codigo:      http.StatusUnprocessableEntity,
			codigoError: "VALIDACION",
			mensaje:     "la fecha de fin no puede ser anterior a la fecha de inicio",
			campo:       "fecha_fin",
			escenario:   "E-11",
		},
		{
			nombre: "E-08 nombre vacio responde 422",
			cuerpo: map[string]any{
				"nombre":       "",
				"fecha_inicio": "2026-09-28",
			},
			preparar: func(s *mockServicioProyectos) {
				s.On("Crear", mock.Anything, mock.Anything).Return(nil, domain.ErrNombreObligatorio)
			},
			codigo:      http.StatusUnprocessableEntity,
			codigoError: "VALIDACION",
			mensaje:     "el nombre del proyecto es obligatorio",
			campo:       "nombre",
			escenario:   "E-08",
		},
		{
			nombre: "E-10 sin fecha de inicio responde 422",
			cuerpo: map[string]any{"nombre": "Proyecto sin inicio"},
			preparar: func(s *mockServicioProyectos) {
				s.On("Crear", mock.Anything, mock.Anything).Return(nil, domain.ErrFechaInicioObligatoria)
			},
			codigo:      http.StatusUnprocessableEntity,
			codigoError: "VALIDACION",
			mensaje:     "la fecha de inicio del proyecto es obligatoria",
			campo:       "fecha_inicio",
			escenario:   "E-10",
		},
		{
			nombre: "E-17 integrante duplicado responde 409",
			cuerpo: map[string]any{
				"nombre":       "Proyecto con duplicado",
				"fecha_inicio": "2026-09-28",
				"integrantes": []map[string]string{
					{"nombre": "Ada Lovelace", "email": "ada@utn.edu.ar"},
					{"nombre": "ada lovelace", "email": "ADA@utn.edu.ar"},
				},
			},
			preparar: func(s *mockServicioProyectos) {
				s.On("Crear", mock.Anything, mock.Anything).Return(nil, domain.ErrIntegranteYaAsociado)
			},
			codigo:      http.StatusConflict,
			codigoError: "INTEGRANTE_DUPLICADO",
			mensaje:     "el integrante ya forma parte del proyecto",
			campo:       "email",
			escenario:   "E-17",
		},
		{
			nombre: "E-49 error inesperado responde 500",
			cuerpo: map[string]any{
				"nombre":       "Proyecto",
				"fecha_inicio": "2026-09-28",
			},
			preparar: func(s *mockServicioProyectos) {
				s.On("Crear", mock.Anything, mock.Anything).Return(nil, errors.New("connection reset by peer"))
			},
			codigo:      http.StatusInternalServerError,
			codigoError: "ERROR_INTERNO",
			mensaje:     "ocurrió un error inesperado al procesar la solicitud",
			escenario:   "E-49",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			router, servicio := nuevoRouterDePrueba(t)
			if c.preparar != nil {
				c.preparar(servicio)
			}

			w := ejecutar(t, router, http.MethodPost, "/api/proyectos", c.cuerpo)

			require.Equal(t, c.codigo, w.Code, "escenario BDD %s cuerpo: %s", c.escenario, w.Body.String())
			if c.codigoError != "" {
				respuesta := decodificarError(t, w)
				assert.Equal(t, c.codigoError, respuesta.Error, "escenario BDD %s", c.escenario)
				if c.mensaje != "" {
					assert.Equal(t, c.mensaje, respuesta.Mensaje)
				}
				assert.Equal(t, c.campo, respuesta.Campo, "escenario BDD %s", c.escenario)
			}
		})
	}
}

func TestListarProyectos(t *testing.T) {
	t.Parallel()

	t.Run("E-21 sin proyectos responde 200 con total 0 y lista vacia", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDePrueba(t)
		servicio.On("Listar", mock.Anything).Return(nil, nil)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos", nil)

		require.Equal(t, http.StatusOK, w.Code)
		var respuesta struct {
			Total     int                 `json:"total"`
			Proyectos []respuestaProyecto `json:"proyectos"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta))
		assert.Equal(t, 0, respuesta.Total)
		require.NotNil(t, respuesta.Proyectos, "el array nunca debe serializarse como null")
		assert.Empty(t, respuesta.Proyectos)
	})

	t.Run("E-22 lista con proyectos y cantidad_integrantes", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDePrueba(t)
		servicio.On("Listar", mock.Anything).Return([]domain.Proyecto{
			*proyectoRespuesta(1, "Proyecto A", &fechaFinHTTP, domain.EstadoActivo, []domain.Integrante{
				{ID: 1, Nombre: "Ada", Email: "ada@utn.edu.ar"},
			}),
			*proyectoRespuesta(2, "Proyecto B", nil, domain.EstadoActivo, nil),
		}, nil)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos", nil)

		require.Equal(t, http.StatusOK, w.Code)
		var respuesta struct {
			Total     int                 `json:"total"`
			Proyectos []respuestaProyecto `json:"proyectos"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta))
		require.Equal(t, 2, respuesta.Total)
		require.Len(t, respuesta.Proyectos, 2)
		assert.Equal(t, 1, respuesta.Proyectos[0].CantidadIntegrantes)
		assert.Equal(t, 0, respuesta.Proyectos[1].CantidadIntegrantes)
		assert.Nil(t, respuesta.Proyectos[1].FechaFin, "la fecha de fin ausente se serializa como null")
	})
}

func TestObtenerProyectoPorID(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		ruta        string
		preparar    func(s *mockServicioProyectos)
		codigo      int
		codigoError string
		escenario   string
	}{
		{
			nombre: "E-05 detalle con integrantes",
			ruta:   "/api/proyectos/1",
			preparar: func(s *mockServicioProyectos) {
				s.On("ObtenerPorID", mock.Anything, uint(1)).Return(
					proyectoRespuesta(1, "Proyecto", &fechaFinHTTP, domain.EstadoActivo, []domain.Integrante{
						{ID: 1, Nombre: "Ada", Email: "ada@utn.edu.ar"},
					}), nil)
			},
			codigo:    http.StatusOK,
			escenario: "E-05",
		},
		{
			nombre: "E-13 proyecto inexistente responde 404",
			ruta:   "/api/proyectos/99",
			preparar: func(s *mockServicioProyectos) {
				s.On("ObtenerPorID", mock.Anything, uint(99)).Return(nil, domain.ErrProyectoNoEncontrado)
			},
			codigo:      http.StatusNotFound,
			codigoError: "PROYECTO_NO_ENCONTRADO",
			escenario:   "E-13",
		},
		{
			nombre:    "E-12 id no numerico responde 400",
			ruta:      "/api/proyectos/abc",
			codigo:    http.StatusBadRequest,
			escenario: "E-12",
		},
		{
			nombre:    "E-12 id cero responde 400",
			ruta:      "/api/proyectos/0",
			codigo:    http.StatusBadRequest,
			escenario: "E-12",
		},
		{
			nombre:    "E-12 id negativo responde 400",
			ruta:      "/api/proyectos/-1",
			codigo:    http.StatusBadRequest,
			escenario: "E-12",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			router, servicio := nuevoRouterDePrueba(t)
			if c.preparar != nil {
				c.preparar(servicio)
			}

			w := ejecutar(t, router, http.MethodGet, c.ruta, nil)

			require.Equal(t, c.codigo, w.Code, "escenario BDD %s", c.escenario)
			if c.codigoError != "" {
				assert.Equal(t, c.codigoError, decodificarError(t, w).Error, "escenario BDD %s", c.escenario)
			}
		})
	}
}

func TestActualizarProyecto(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		cuerpo      any
		preparar    func(s *mockServicioProyectos)
		codigo      int
		codigoError string
		mensaje     string
		escenario   string
	}{
		{
			nombre: "E-10 modificacion valida responde 200",
			cuerpo: map[string]any{
				"nombre":       "Nombre nuevo",
				"fecha_inicio": "2026-09-28",
				"fecha_fin":    "2026-12-20",
			},
			preparar: func(s *mockServicioProyectos) {
				s.On("Actualizar", mock.Anything, uint(1), mock.Anything).Return(
					proyectoRespuesta(1, "Nombre nuevo", &fechaFinHTTP, domain.EstadoActivo, nil), nil)
			},
			codigo:    http.StatusOK,
			escenario: "E-10",
		},
		{
			nombre: "E-32 proyecto cerrado responde 409",
			cuerpo: map[string]any{
				"nombre":       "Intento",
				"fecha_inicio": "2026-09-28",
			},
			preparar: func(s *mockServicioProyectos) {
				s.On("Actualizar", mock.Anything, uint(1), mock.Anything).Return(nil, domain.ErrProyectoCerrado)
			},
			codigo:      http.StatusConflict,
			codigoError: "PROYECTO_CERRADO",
			mensaje:     "el proyecto está cerrado: solo se permite cambiar su estado o reabrirlo",
			escenario:   "E-32",
		},
		{
			nombre: "E-30 proyecto inexistente responde 404",
			cuerpo: map[string]any{
				"nombre":       "Intento",
				"fecha_inicio": "2026-09-28",
			},
			preparar: func(s *mockServicioProyectos) {
				s.On("Actualizar", mock.Anything, uint(1), mock.Anything).Return(nil, domain.ErrProyectoNoEncontrado)
			},
			codigo:      http.StatusNotFound,
			codigoError: "PROYECTO_NO_ENCONTRADO",
			escenario:   "E-30",
		},
		{
			nombre: "E-23 fecha de fin con formato invalido responde 400",
			cuerpo: map[string]any{
				"nombre":       "Proyecto",
				"fecha_inicio": "2026-09-28",
				"fecha_fin":    "18-12-2026",
			},
			codigo:      http.StatusBadRequest,
			codigoError: "PARAMETRO_INVALIDO",
			escenario:   "E-23",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			router, servicio := nuevoRouterDePrueba(t)
			if c.preparar != nil {
				c.preparar(servicio)
			}

			w := ejecutar(t, router, http.MethodPut, "/api/proyectos/1", c.cuerpo)

			require.Equal(t, c.codigo, w.Code, "escenario BDD %s cuerpo: %s", c.escenario, w.Body.String())
			if c.codigoError != "" {
				respuesta := decodificarError(t, w)
				assert.Equal(t, c.codigoError, respuesta.Error, "escenario BDD %s", c.escenario)
				if c.mensaje != "" {
					assert.Equal(t, c.mensaje, respuesta.Mensaje)
				}
			}
		})
	}
}

func TestCambiarEstadoProyecto(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		cuerpo      any
		preparar    func(s *mockServicioProyectos)
		codigo      int
		codigoError string
		mensaje     string
		escenario   string
	}{
		{
			nombre: "E-33 cierre responde 200",
			cuerpo: map[string]string{"estado": "Cerrado"},
			preparar: func(s *mockServicioProyectos) {
				s.On("CambiarEstado", mock.Anything, uint(1), domain.EstadoCerrado).Return(
					proyectoRespuesta(1, "Proyecto", &fechaFinHTTP, domain.EstadoCerrado, nil), nil)
			},
			codigo:    http.StatusOK,
			escenario: "E-33",
		},
		{
			nombre: "E-34 reapertura responde 200",
			cuerpo: map[string]string{"estado": "Activo"},
			preparar: func(s *mockServicioProyectos) {
				s.On("CambiarEstado", mock.Anything, uint(1), domain.EstadoActivo).Return(
					proyectoRespuesta(1, "Proyecto", &fechaFinHTTP, domain.EstadoActivo, nil), nil)
			},
			codigo:    http.StatusOK,
			escenario: "E-34",
		},
		{
			nombre: "E-36 estado invalido responde 400",
			cuerpo: map[string]string{"estado": "activo"},
			preparar: func(s *mockServicioProyectos) {
				s.On("CambiarEstado", mock.Anything, uint(1), domain.EstadoProyecto("activo")).
					Return(nil, domain.ErrEstadoInvalido)
			},
			codigo:      http.StatusBadRequest,
			codigoError: "ESTADO_INVALIDO",
			mensaje:     `el estado del proyecto debe ser "Activo" o "Cerrado"`,
			escenario:   "E-36",
		},
		{
			nombre: "E-37 cerrar sin fecha de fin responde 422",
			cuerpo: map[string]string{"estado": "Cerrado"},
			preparar: func(s *mockServicioProyectos) {
				s.On("CambiarEstado", mock.Anything, uint(1), domain.EstadoCerrado).
					Return(nil, domain.ErrFechaFinRequerida)
			},
			codigo:      http.StatusUnprocessableEntity,
			codigoError: "VALIDACION",
			mensaje:     "no se puede cerrar un proyecto sin fecha de fin: informá la fecha de fin antes de cerrarlo",
			escenario:   "E-37",
		},
		{
			nombre:    "E-20 body vacio responde 400",
			codigo:    http.StatusBadRequest,
			escenario: "E-20",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			router, servicio := nuevoRouterDePrueba(t)
			if c.preparar != nil {
				c.preparar(servicio)
			}

			w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/estado", c.cuerpo)

			require.Equal(t, c.codigo, w.Code, "escenario BDD %s cuerpo: %s", c.escenario, w.Body.String())
			if c.codigoError != "" {
				respuesta := decodificarError(t, w)
				assert.Equal(t, c.codigoError, respuesta.Error, "escenario BDD %s", c.escenario)
				if c.mensaje != "" {
					assert.Equal(t, c.mensaje, respuesta.Mensaje)
				}
			}
		})
	}
}

func TestEliminarProyecto(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		preparar    func(s *mockServicioProyectos)
		codigo      int
		codigoError string
		escenario   string
	}{
		{
			nombre: "E-51 baja logica responde 204",
			preparar: func(s *mockServicioProyectos) {
				s.On("Eliminar", mock.Anything, uint(1)).Return(nil)
			},
			codigo:    http.StatusNoContent,
			escenario: "E-51",
		},
		{
			nombre: "E-52 proyecto con historial responde 409",
			preparar: func(s *mockServicioProyectos) {
				s.On("Eliminar", mock.Anything, uint(1)).Return(domain.ErrProyectoConHistorial)
			},
			codigo:      http.StatusConflict,
			codigoError: "PROYECTO_CON_HISTORIAL",
			escenario:   "E-52",
		},
		{
			nombre: "E-50 proyecto dado de baja responde 404",
			preparar: func(s *mockServicioProyectos) {
				s.On("Eliminar", mock.Anything, uint(1)).Return(domain.ErrProyectoNoEncontrado)
			},
			codigo:      http.StatusNotFound,
			codigoError: "PROYECTO_NO_ENCONTRADO",
			escenario:   "E-50",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			router, servicio := nuevoRouterDePrueba(t)
			c.preparar(servicio)

			w := ejecutar(t, router, http.MethodDelete, "/api/proyectos/1", nil)

			require.Equal(t, c.codigo, w.Code, "escenario BDD %s", c.escenario)
			if c.codigo == http.StatusNoContent {
				assert.Empty(t, w.Body.String(), "204 no debe devolver cuerpo")
			}
			if c.codigoError != "" {
				assert.Equal(t, c.codigoError, decodificarError(t, w).Error, "escenario BDD %s", c.escenario)
			}
		})
	}
}

func TestAgregarIntegrante(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		cuerpo      any
		preparar    func(s *mockServicioProyectos)
		codigo      int
		codigoError string
		mensaje     string
		campo       string
		escenario   string
	}{
		{
			nombre: "E-40 alta de integrante responde 201 con el proyecto_id",
			cuerpo: map[string]string{"nombre": "Jazmín Pérez", "email": "jazmin.perez@utn.edu.ar"},
			preparar: func(s *mockServicioProyectos) {
				s.On("AgregarIntegrante", mock.Anything, uint(1), mock.Anything).Return(
					&domain.Integrante{ID: 7, Nombre: "Jazmín Pérez", Email: "jazmin.perez@utn.edu.ar"}, nil)
			},
			codigo:    http.StatusCreated,
			escenario: "E-40",
		},
		{
			nombre: "E-44 integrante ya asociado responde 409",
			cuerpo: map[string]string{"nombre": "Ada", "email": "ada@utn.edu.ar"},
			preparar: func(s *mockServicioProyectos) {
				s.On("AgregarIntegrante", mock.Anything, uint(1), mock.Anything).
					Return(nil, domain.ErrIntegranteYaAsociado)
			},
			codigo:      http.StatusConflict,
			codigoError: "INTEGRANTE_DUPLICADO",
			escenario:   "E-44",
		},
		{
			nombre: "E-53 proyecto dado de baja responde 404",
			cuerpo: map[string]string{"nombre": "Ada", "email": "ada@utn.edu.ar"},
			preparar: func(s *mockServicioProyectos) {
				s.On("AgregarIntegrante", mock.Anything, uint(1), mock.Anything).
					Return(nil, domain.ErrProyectoNoEncontrado)
			},
			codigo:      http.StatusNotFound,
			codigoError: "PROYECTO_NO_ENCONTRADO",
			escenario:   "E-53",
		},
		{
			nombre: "E-48 proyecto cerrado responde 409",
			cuerpo: map[string]string{"nombre": "Ada", "email": "ada@utn.edu.ar"},
			preparar: func(s *mockServicioProyectos) {
				s.On("AgregarIntegrante", mock.Anything, uint(1), mock.Anything).
					Return(nil, domain.ErrProyectoCerrado)
			},
			codigo:      http.StatusConflict,
			codigoError: "PROYECTO_CERRADO",
			escenario:   "E-48",
		},
		{
			nombre: "E-42 email con formato Nombre <correo> responde 422",
			cuerpo: map[string]string{"nombre": "Ada", "email": "Ada Lovelace <ada@utn.edu.ar>"},
			preparar: func(s *mockServicioProyectos) {
				s.On("AgregarIntegrante", mock.Anything, uint(1), mock.Anything).
					Return(nil, domain.ErrIntegranteEmailInvalido)
			},
			codigo:      http.StatusUnprocessableEntity,
			codigoError: "VALIDACION",
			campo:       "email",
			escenario:   "E-42",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			router, servicio := nuevoRouterDePrueba(t)
			c.preparar(servicio)

			w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/integrantes", c.cuerpo)

			require.Equal(t, c.codigo, w.Code, "escenario BDD %s cuerpo: %s", c.escenario, w.Body.String())
			if c.codigo == http.StatusCreated {
				var respuesta struct {
					ID         uint `json:"id"`
					ProyectoID uint `json:"proyecto_id"`
				}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta))
				assert.Equal(t, uint(7), respuesta.ID)
				assert.Equal(t, uint(1), respuesta.ProyectoID)
			}
			if c.codigoError != "" {
				assert.Equal(t, c.codigoError, decodificarError(t, w).Error, "escenario BDD %s", c.escenario)
			}
		})
	}
}

func TestQuitarIntegrante(t *testing.T) {
	t.Parallel()

	t.Run("E-46 desvinculacion responde 204", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDePrueba(t)
		servicio.On("QuitarIntegrante", mock.Anything, uint(1), uint(4)).Return(nil)

		w := ejecutar(t, router, http.MethodDelete, "/api/proyectos/1/integrantes/4", nil)

		require.Equal(t, http.StatusNoContent, w.Code)
		assert.Empty(t, w.Body.String(), "204 no debe devolver cuerpo")
		servicio.AssertExpectations(t)
	})

	t.Run("E-47 integrante no asociado responde 404", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDePrueba(t)
		servicio.On("QuitarIntegrante", mock.Anything, uint(1), uint(9)).Return(domain.ErrIntegranteNoAsociado)

		w := ejecutar(t, router, http.MethodDelete, "/api/proyectos/1/integrantes/9", nil)

		require.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "INTEGRANTE_NO_ENCONTRADO", decodificarError(t, w).Error)
	})

	t.Run("E-12 integranteId no numerico responde 400", func(t *testing.T) {
		t.Parallel()
		router, _ := nuevoRouterDePrueba(t)

		w := ejecutar(t, router, http.MethodDelete, "/api/proyectos/1/integrantes/xyz", nil)

		require.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "PARAMETRO_INVALIDO", decodificarError(t, w).Error)
	})

	t.Run("E-48 proyecto cerrado responde 409", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDePrueba(t)
		servicio.On("QuitarIntegrante", mock.Anything, uint(1), uint(4)).Return(domain.ErrProyectoCerrado)

		w := ejecutar(t, router, http.MethodDelete, "/api/proyectos/1/integrantes/4", nil)

		require.Equal(t, http.StatusConflict, w.Code)
		assert.Equal(t, "PROYECTO_CERRADO", decodificarError(t, w).Error)
	})
}

func TestContextoCanceladoSePropaga(t *testing.T) {
	t.Parallel()

	router, servicio := nuevoRouterDePrueba(t)
	servicio.On("Listar", mock.Anything).Return(nil, context.Canceled)

	w := ejecutar(t, router, http.MethodGet, "/api/proyectos", nil)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestIntegrantesAnidadosNoExponenProyectoIdCero protege el contrato: dentro de
// un proyecto el vinculo es implicito, por lo que emitir "proyecto_id": 0 es ruido
// que confunde al cliente. El smoke test end-to-end detecto la fuga.
func TestIntegrantesAnidadosNoExponenProyectoIdCero(t *testing.T) {
	t.Parallel()

	router, servicio := nuevoRouterDePrueba(t)
	servicio.On("ObtenerPorID", mock.Anything, uint(1)).Return(
		proyectoRespuesta(1, "Proyecto", nil, domain.EstadoActivo, []domain.Integrante{
			{ID: 7, Nombre: "Ada", Email: "ada@utn.edu.ar"},
		}), nil)

	w := ejecutar(t, router, http.MethodGet, "/api/proyectos/1", nil)
	require.Equal(t, http.StatusOK, w.Code)

	var mapa map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &mapa))
	integrantes, ok := mapa["integrantes"].([]any)
	require.True(t, ok)
	require.Len(t, integrantes, 1)

	integrante := integrantes[0].(map[string]any)
	assert.NotContains(t, integrante, "proyecto_id",
		"el vinculo al proyecto es implicito en la respuesta anidada")
	assert.Equal(t, "ada@utn.edu.ar", integrante["email"])
}

// TestAltaIntegranteSiInformaProyectoId comprueba el caso contrario: en la
// respuesta del alta del vinculo el proyecto si es informacion relevante.
func TestAltaIntegranteSiInformaProyectoId(t *testing.T) {
	t.Parallel()

	router, servicio := nuevoRouterDePrueba(t)
	servicio.On("AgregarIntegrante", mock.Anything, uint(1),
		domain.IntegranteInput{Nombre: "Carla", Email: "carla@utn.edu.ar"}).
		Return(&domain.Integrante{ID: 9, Nombre: "Carla", Email: "carla@utn.edu.ar"}, nil)

	w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/integrantes",
		altaIntegranteDTO{Nombre: "Carla", Email: "carla@utn.edu.ar"})
	require.Equal(t, http.StatusCreated, w.Code)

	var mapa map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &mapa))
	assert.Equal(t, float64(1), mapa["proyecto_id"])
	assert.Equal(t, float64(9), mapa["id"])
}

// TestListadoDevuelveIntegrantesComoArregloVacio evita que el listado exponga
// "integrantes": null, que obliga al cliente a distinguir null de lista vacia.
func TestListadoDevuelveIntegrantesComoArregloVacio(t *testing.T) {
	t.Parallel()

	router, servicio := nuevoRouterDePrueba(t)
	servicio.On("Listar", mock.Anything).Return([]domain.Proyecto{
		*proyectoRespuesta(1, "Proyecto", nil, domain.EstadoActivo, nil),
	}, nil)

	w := ejecutar(t, router, http.MethodGet, "/api/proyectos", nil)
	require.Equal(t, http.StatusOK, w.Code)

	var mapa struct {
		Total     int `json:"total"`
		Proyectos []struct {
			Integrantes []respuestaIntegrante `json:"integrantes"`
		} `json:"proyectos"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &mapa))
	assert.Equal(t, 1, mapa.Total)
	require.Len(t, mapa.Proyectos, 1)
	assert.NotNil(t, mapa.Proyectos[0].Integrantes,
		"el listado debe devolver un arreglo vacio, no null")
	assert.Empty(t, mapa.Proyectos[0].Integrantes)
}

func proyectoRespuesta(
	id uint,
	nombre string,
	fechaFin *time.Time,
	estado domain.EstadoProyecto,
	integrantes []domain.Integrante,
) *domain.Proyecto {
	if integrantes == nil {
		integrantes = []domain.Integrante{}
	}
	return &domain.Proyecto{
		ID:            id,
		Nombre:        nombre,
		FechaInicio:   fechaInicioHTTP,
		FechaFin:      fechaFin,
		Estado:        estado,
		Integrantes:   integrantes,
		CreadoEn:      fechaInicioHTTP,
		ActualizadoEn: fechaInicioHTTP,
	}
}
