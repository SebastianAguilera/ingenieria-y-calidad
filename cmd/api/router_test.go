package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

// servicioQueFalla implements handler.ServicioProyectos returning always
// ErrProyectoNoEncontrado. Its purpose is to make it verifiable that the router
// actually injects a usable service: with a nil service the four routes that
// reach it panic, gin.Recovery turns the panic into a 500, and a test that
// only checks "not 404" would pass while the wiring is broken.
type servicioQueFalla struct{}

func (servicioQueFalla) Crear(context.Context, domain.NuevoProyecto) (*domain.Proyecto, error) {
	return nil, domain.ErrProyectoNoEncontrado
}

func (servicioQueFalla) Listar(context.Context) ([]domain.Proyecto, error) {
	return nil, domain.ErrProyectoNoEncontrado
}

func (servicioQueFalla) ObtenerPorID(context.Context, uint) (*domain.Proyecto, error) {
	return nil, domain.ErrProyectoNoEncontrado
}

func (servicioQueFalla) Actualizar(context.Context, uint, domain.ActualizacionProyecto) (*domain.Proyecto, error) {
	return nil, domain.ErrProyectoNoEncontrado
}

func (servicioQueFalla) CambiarEstado(context.Context, uint, domain.EstadoProyecto) (*domain.Proyecto, error) {
	return nil, domain.ErrProyectoNoEncontrado
}

func (servicioQueFalla) Eliminar(context.Context, uint) error {
	return domain.ErrProyectoNoEncontrado
}

func (servicioQueFalla) AgregarUsuario(context.Context, uint, domain.UsuarioInput) (*domain.Usuario, error) {
	return nil, domain.ErrProyectoNoEncontrado
}

func (servicioQueFalla) QuitarUsuario(context.Context, uint, uint) error {
	return domain.ErrProyectoNoEncontrado
}

// TestRutasDeProyectosAlcanzanElServicio verifies that the eight routes of
// US-01 are registered and that the router injects a service the handlers can
// use. Each request carries a body that passes the syntactic validation layer,
// so reaching the service and receiving its ErrProyectoNoEncontrado back as a
// 404 PROYECTO_NO_ENCONTRADO is the only possible outcome. A missing route
// would give 404 with an empty body, a collision 405, and a nil service a 500
// coming from the recovery middleware.
func TestRutasDeProyectosAlcanzanElServicio(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter(servicioQueFalla{})

	casos := []struct {
		nombre    string
		metodo    string
		ruta      string
		cuerpo    string
		conCuerpo bool
	}{
		{
			nombre:    "alta de proyecto",
			metodo:    http.MethodPost,
			ruta:      "/api/proyectos",
			cuerpo:    `{"nombre":"Proyecto","fecha_inicio":"2026-03-02"}`,
			conCuerpo: true,
		},
		{
			nombre: "listado de proyectos",
			metodo: http.MethodGet,
			ruta:   "/api/proyectos",
		},
		{
			nombre: "detalle de proyecto",
			metodo: http.MethodGet,
			ruta:   "/api/proyectos/1",
		},
		{
			nombre:    "actualizacion de proyecto",
			metodo:    http.MethodPut,
			ruta:      "/api/proyectos/1",
			cuerpo:    `{"nombre":"Proyecto","fecha_inicio":"2026-03-02"}`,
			conCuerpo: true,
		},
		{
			nombre:    "cambio de estado",
			metodo:    http.MethodPatch,
			ruta:      "/api/proyectos/1/estado",
			cuerpo:    `{"estado":"Cerrado"}`,
			conCuerpo: true,
		},
		{
			nombre: "baja de proyecto",
			metodo: http.MethodDelete,
			ruta:   "/api/proyectos/1",
		},
		{
			nombre:    "alta de usuario",
			metodo:    http.MethodPost,
			ruta:      "/api/proyectos/1/usuarios",
			cuerpo:    `{"nombre":"Ada","email":"ada@utn.edu.ar"}`,
			conCuerpo: true,
		},
		{
			nombre: "baja de usuario",
			metodo: http.MethodDelete,
			ruta:   "/api/proyectos/1/usuarios/1",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			var lector *bytes.Reader
			if caso.conCuerpo {
				lector = bytes.NewReader([]byte(caso.cuerpo))
			} else {
				lector = bytes.NewReader(nil)
			}

			w := httptest.NewRecorder()
			req, err := http.NewRequest(caso.metodo, caso.ruta, lector)
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code,
				"la ruta no debe colisionar con otra: %s %s", caso.metodo, caso.ruta)

			var cuerpo struct {
				Error   string `json:"error"`
				Mensaje string `json:"mensaje"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cuerpo),
				"la respuesta debe ser un cuerpo de error de la API: %s", w.Body.String())

			assert.Equal(t, "PROYECTO_NO_ENCONTRADO", cuerpo.Error,
				"el handler debe alcanzar el servicio inyectado: %s %s", caso.metodo, caso.ruta)
			assert.Equal(t, domain.ErrProyectoNoEncontrado.Error(), cuerpo.Mensaje)
		})
	}
}

// TestHealthNoColisionaConLasRutasDeProyectos guards the other assertion of
// the route registration plan: /health must remain independent of the /api
// group, so it keeps answering 200 even with a service that always fails.
func TestHealthNoColisionaConLasRutasDeProyectos(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter(servicioQueFalla{})

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/health", nil)
	require.NoError(t, err)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), "PROYECTO_NO_ENCONTRADO")
}
