package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

func (servicioQueFalla) Crear(context.Context, uint, domain.NuevoProyecto) (*domain.Proyecto, error) {
	return nil, domain.ErrProyectoNoEncontrado
}

func (servicioQueFalla) Listar(context.Context, uint) ([]domain.Proyecto, error) {
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

func (servicioQueFalla) ListarUsuarios(context.Context, uint) ([]domain.MiembroProyecto, error) {
	return nil, domain.ErrProyectoNoEncontrado
}

func (servicioQueFalla) AgregarUsuario(context.Context, uint, domain.AsociacionUsuario) (*domain.Usuario, error) {
	return nil, domain.ErrProyectoNoEncontrado
}

func (servicioQueFalla) CambiarRol(context.Context, uint, uint, domain.RolProyecto) error {
	return domain.ErrProyectoNoEncontrado
}

func (servicioQueFalla) QuitarUsuario(context.Context, uint, uint) error {
	return domain.ErrProyectoNoEncontrado
}

// PerteneceA devuelve siempre true para que los tests de enrutamiento lleguen
// al service: el guard de US-F10 se prueba aparte con servicioSinMembresia.
func (servicioQueFalla) PerteneceA(context.Context, uint, uint) (bool, error) {
	return true, nil
}

// EsLider devuelve siempre true: los tests de enrutamiento no deben chocar con
// el guard de permisos de US-11, que se prueba aparte con servicioMiembro.
func (servicioQueFalla) EsLider(context.Context, uint, uint) (bool, error) {
	return true, nil
}

func (servicioQueFalla) ObtenerRol(context.Context, uint, uint) (domain.RolProyecto, error) {
	return domain.RolLider, nil
}

// servicioMiembro es un integrante sin permisos de lider: cuenta las consultas
// de rol y marca si una operacion de lider llego al servicio.
type servicioMiembro struct {
	servicioQueFalla
	consultasDeRol int
	eliminado      bool
}

func (s *servicioMiembro) EsLider(context.Context, uint, uint) (bool, error) {
	s.consultasDeRol++
	return false, nil
}

func (s *servicioMiembro) Eliminar(context.Context, uint) error {
	s.eliminado = true
	return nil
}

func (s *servicioMiembro) CambiarRol(context.Context, uint, uint, domain.RolProyecto) error {
	s.eliminado = true
	return nil
}

// servicioSinMembresia niega la membresia de todo proyecto y marca lo que el
// guard debio impedir: el handler nunca debe alcanzarse.
type servicioSinMembresia struct {
	servicioQueFalla
	consultas         int
	proyectoAlcanzado bool
	listadoAlcanzado  bool
}

func (s *servicioSinMembresia) PerteneceA(context.Context, uint, uint) (bool, error) {
	s.consultas++
	return false, nil
}

func (s *servicioSinMembresia) ObtenerPorID(context.Context, uint) (*domain.Proyecto, error) {
	s.proyectoAlcanzado = true
	return nil, domain.ErrProyectoCerrado
}

func (s *servicioSinMembresia) Listar(context.Context, uint) ([]domain.Proyecto, error) {
	s.listadoAlcanzado = true
	return []domain.Proyecto{}, nil
}

// servicioHistoriasQueFalla implementa handler.ServicioHistorias devolviendo
// siempre ErrHistoriaNoEncontrada. Es un tipo aparte de servicioQueFalla porque
// Go no admite sobrecarga de metodos: Crear, Listar y Actualizar colisionan
// entre ambos puertos.
type servicioHistoriasQueFalla struct{}

func (servicioHistoriasQueFalla) Crear(context.Context, uint, domain.NuevaHistoria) (*domain.Historia, error) {
	return nil, domain.ErrHistoriaNoEncontrada
}

func (servicioHistoriasQueFalla) Listar(context.Context, uint) ([]domain.Historia, error) {
	return nil, domain.ErrHistoriaNoEncontrada
}

func (servicioHistoriasQueFalla) Actualizar(context.Context, uint, uint, domain.ActualizacionHistoria) (*domain.Historia, error) {
	return nil, domain.ErrHistoriaNoEncontrada
}

// servicioSprintsQueFalla implementa handler.ServicioSprints devolviendo
// siempre ErrSprintNoEncontrado. Es un tipo aparte porque Go no admite
// sobrecarga de metodos: Crear, Listar y CambiarEstado colisionan con los
// otros puertos.
type servicioSprintsQueFalla struct{}

func (servicioSprintsQueFalla) Crear(context.Context, uint, domain.NuevoSprint) (*domain.Sprint, error) {
	return nil, domain.ErrSprintNoEncontrado
}

func (servicioSprintsQueFalla) Listar(context.Context, uint) ([]domain.Sprint, error) {
	return nil, domain.ErrSprintNoEncontrado
}

func (servicioSprintsQueFalla) CambiarEstado(context.Context, uint, uint, domain.EstadoSprint) (*domain.Sprint, error) {
	return nil, domain.ErrSprintNoEncontrado
}

type servicioSprintBacklogQueFalla struct{}

func (servicioSprintBacklogQueFalla) AsignarASprint(context.Context, uint, uint, uint) error {
	return domain.ErrHistoriaYaAsignada
}

func (servicioSprintBacklogQueFalla) DesasignarDeSprint(context.Context, uint, uint, uint) error {
	return domain.ErrSprintCerrado
}

func (servicioSprintBacklogQueFalla) ListarSprintBacklog(context.Context, uint, uint) (*domain.SprintBacklogDetalle, error) {
	return &domain.SprintBacklogDetalle{
		Sprint: &domain.Sprint{ID: 1},
		Items:  []domain.SprintBacklogItemDetalle{},
	}, nil
}

type servicioDefectosQueFalla struct{}

func (servicioDefectosQueFalla) Crear(context.Context, uint, domain.NuevoDefecto) (*domain.Defecto, error) {
	return nil, domain.ErrDefectoNoEncontrado
}

func (servicioDefectosQueFalla) Actualizar(context.Context, uint, uint, domain.ActualizacionDefecto) (*domain.Defecto, error) {
	return nil, domain.ErrDefectoNoEncontrado
}

func (servicioDefectosQueFalla) Listar(context.Context, uint) ([]domain.Defecto, error) {
	return nil, domain.ErrDefectoNoEncontrado
}

type servicioCriteriosQueFalla struct{}

func (servicioCriteriosQueFalla) Crear(context.Context, uint, uint, domain.NuevoAcceptanceCriterion) (*domain.AcceptanceCriterion, error) {
	return nil, domain.ErrCriterioNoEncontrado
}

func (servicioCriteriosQueFalla) Listar(context.Context, uint, uint) ([]domain.AcceptanceCriterion, error) {
	return nil, domain.ErrCriterioNoEncontrado
}

func (servicioCriteriosQueFalla) Actualizar(context.Context, uint, uint, uint, domain.ActualizacionAcceptanceCriterion) (*domain.AcceptanceCriterion, error) {
	return nil, domain.ErrCriterioNoEncontrado
}

func (servicioCriteriosQueFalla) Eliminar(context.Context, uint, uint, uint) error {
	return domain.ErrCriterioNoEncontrado
}

// servicioPokerQueFalla implementa handler.ServicioPoker devolviendo siempre
// ErrSesionPokerNoEncontrada, de modo que una peticion que alcanza el servicio
// responde 404 SESION_POKER_NO_ENCONTRADA. Marca alcanzado para que el test del
// guard verifique que el servicio de poker nunca se ejecuta (E-15).
type servicioPokerQueFalla struct {
	alcanzado bool
}

func (s *servicioPokerQueFalla) CrearSesion(context.Context, uint, uint, uint) (*domain.DetalleSesionPoker, error) {
	s.alcanzado = true
	return nil, domain.ErrSesionPokerNoEncontrada
}

func (s *servicioPokerQueFalla) ListarPorHistoria(context.Context, uint, uint) ([]domain.SesionPoker, error) {
	s.alcanzado = true
	return nil, domain.ErrSesionPokerNoEncontrada
}

func (s *servicioPokerQueFalla) Detalle(context.Context, uint, uint, uint) (*domain.DetalleSesionPoker, error) {
	s.alcanzado = true
	return nil, domain.ErrSesionPokerNoEncontrada
}

func (s *servicioPokerQueFalla) Votar(context.Context, uint, uint, uint, float64) (*domain.ResultadoVotoPoker, error) {
	s.alcanzado = true
	return nil, domain.ErrSesionPokerNoEncontrada
}

func (s *servicioPokerQueFalla) Activas(context.Context, uint, uint) ([]domain.SesionPokerActiva, error) {
	s.alcanzado = true
	return nil, domain.ErrSesionPokerNoEncontrada
}

func (s *servicioPokerQueFalla) Revelar(context.Context, uint, uint, uint) (*domain.DetalleSesionPoker, error) {
	s.alcanzado = true
	return nil, domain.ErrSesionPokerNoEncontrada
}

func (s *servicioPokerQueFalla) NuevaRonda(context.Context, uint, uint, uint) (*domain.DetalleSesionPoker, error) {
	s.alcanzado = true
	return nil, domain.ErrSesionPokerNoEncontrada
}

func (s *servicioPokerQueFalla) Acordar(context.Context, uint, uint, uint, float64) (*domain.DetalleSesionPoker, error) {
	s.alcanzado = true
	return nil, domain.ErrSesionPokerNoEncontrada
}

func (s *servicioPokerQueFalla) Cancelar(context.Context, uint, uint, uint) (*domain.DetalleSesionPoker, error) {
	s.alcanzado = true
	return nil, domain.ErrSesionPokerNoEncontrada
}

// TestRutasDeProyectosAlcanzanElServicio verifies that the routes of
// US-01, US-08 and US-12 are registered and that the router injects a service the handlers can
// use. Each request carries a body that passes the syntactic validation layer,
// so reaching the service and receiving its ErrProyectoNoEncontrado back as a
// 404 PROYECTO_NO_ENCONTRADO is the only possible outcome. A missing route
// would give 404 with an empty body, a collision 405, and a nil service a 500
// coming from the recovery middleware.
func TestRutasDeProyectosAlcanzanElServicio(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter(servicioQueFalla{}, servicioHistoriasQueFalla{}, servicioSprintsQueFalla{}, nil, nil, nil, nil, nil, nil, secretoDePruebaRouter)

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
			cuerpo:    `{"email":"ada@utn.edu.ar","rol":"MIEMBRO"}`,
			conCuerpo: true,
		},
		{
			nombre:    "listado de integrantes",
			metodo:    http.MethodGet,
			ruta:      "/api/proyectos/1/usuarios",
			conCuerpo: false,
		},
		{
			nombre:    "cambio de rol",
			metodo:    http.MethodPatch,
			ruta:      "/api/proyectos/1/usuarios/1",
			cuerpo:    `{"rol":"LIDER"}`,
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
			req.Header.Set("Authorization", "Bearer "+tokenFirmadoEnElRouter(t, secretoDePruebaRouter))
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

// TestRutasDeHistoriasAlcanzanElServicio verifies that the three routes of
// US-02 are registered and that the router injects the backlog service. As in
// the projects suite, a request with a syntactically valid body reaches the
// service and its ErrHistoriaNoEncontrada comes back as 404
// HISTORIA_NO_ENCONTRADA. A nil service would surface as a 500.
func TestRutasDeHistoriasAlcanzanElServicio(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter(servicioQueFalla{}, servicioHistoriasQueFalla{}, servicioSprintsQueFalla{}, nil, nil, nil, nil, nil, nil, secretoDePruebaRouter)

	casos := []struct {
		nombre    string
		metodo    string
		ruta      string
		cuerpo    string
		conCuerpo bool
	}{
		{
			nombre:    "alta de historia",
			metodo:    http.MethodPost,
			ruta:      "/api/proyectos/1/historias",
			cuerpo:    `{"titulo":"Historia"}`,
			conCuerpo: true,
		},
		{
			nombre: "listado de historias",
			metodo: http.MethodGet,
			ruta:   "/api/proyectos/1/historias",
		},
		{
			nombre:    "modificacion de historia",
			metodo:    http.MethodPatch,
			ruta:      "/api/proyectos/1/historias/5",
			cuerpo:    `{"prioridad":"Alta"}`,
			conCuerpo: true,
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
			req.Header.Set("Authorization", "Bearer "+tokenFirmadoEnElRouter(t, secretoDePruebaRouter))
			router.ServeHTTP(w, req)

			assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code,
				"la ruta no debe colisionar con otra: %s %s", caso.metodo, caso.ruta)

			var cuerpo struct {
				Error   string `json:"error"`
				Mensaje string `json:"mensaje"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cuerpo),
				"la respuesta debe ser un cuerpo de error de la API: %s", w.Body.String())

			assert.Equal(t, "HISTORIA_NO_ENCONTRADA", cuerpo.Error,
				"el handler debe alcanzar el servicio inyectado: %s %s", caso.metodo, caso.ruta)
			assert.Equal(t, domain.ErrHistoriaNoEncontrada.Error(), cuerpo.Mensaje)
		})
	}
}

// TestRutasDeSprintsAlcanzanElServicio verifies that the three routes of
// US-03 are registered and that the router injects the sprints service. As in
// the previous suites, a request with a syntactically valid body reaches the
// service and its ErrSprintNoEncontrado comes back as 404 SPRINT_NO_ENCONTRADA.
// A nil service would surface as a 500.
func TestRutasDeSprintsAlcanzanElServicio(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter(servicioQueFalla{}, servicioHistoriasQueFalla{}, servicioSprintsQueFalla{}, nil, nil, nil, nil, nil, nil, secretoDePruebaRouter)

	casos := []struct {
		nombre    string
		metodo    string
		ruta      string
		cuerpo    string
		conCuerpo bool
	}{
		{
			nombre:    "alta de sprint",
			metodo:    http.MethodPost,
			ruta:      "/api/proyectos/1/sprints",
			cuerpo:    `{"identificador":"Sprint 1","sprint_goal":"MVP","fecha_inicio":"2026-10-06","fecha_fin":"2026-10-20"}`,
			conCuerpo: true,
		},
		{
			nombre: "listado de sprints",
			metodo: http.MethodGet,
			ruta:   "/api/proyectos/1/sprints",
		},
		{
			nombre:    "cambio de estado de sprint",
			metodo:    http.MethodPatch,
			ruta:      "/api/proyectos/1/sprints/1/estado",
			cuerpo:    `{"estado":"Activo"}`,
			conCuerpo: true,
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
			req.Header.Set("Authorization", "Bearer "+tokenFirmadoEnElRouter(t, secretoDePruebaRouter))
			router.ServeHTTP(w, req)

			assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code,
				"la ruta no debe colisionar con otra: %s %s", caso.metodo, caso.ruta)

			var cuerpo struct {
				Error   string `json:"error"`
				Mensaje string `json:"mensaje"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cuerpo),
				"la respuesta debe ser un cuerpo de error de la API: %s", w.Body.String())

			assert.Equal(t, "SPRINT_NO_ENCONTRADA", cuerpo.Error,
				"el handler debe alcanzar el servicio inyectado: %s %s", caso.metodo, caso.ruta)
			assert.Equal(t, domain.ErrSprintNoEncontrado.Error(), cuerpo.Mensaje)
		})
	}
}

// TestHealthNoColisionaConLasRutasDeProyectos guards the other assertion of
// the route registration plan: /health must remain independent of the /api
// group, so it keeps answering 200 even with a service that always fails.
func TestHealthNoColisionaConLasRutasDeProyectos(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter(servicioQueFalla{}, servicioHistoriasQueFalla{}, servicioSprintsQueFalla{}, nil, nil, nil, nil, nil, nil, secretoDePruebaRouter)

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/health", nil)
	require.NoError(t, err)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), "PROYECTO_NO_ENCONTRADO")
}

// TestRutasDeSprintBacklogAlcanzanElServicio verifica que las rutas de US-04 estén registradas.
func TestRutasDeSprintBacklogAlcanzanElServicio(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter(servicioQueFalla{}, servicioHistoriasQueFalla{}, servicioSprintsQueFalla{}, nil, servicioSprintBacklogQueFalla{}, nil, nil, nil, nil, secretoDePruebaRouter)

	casos := []struct {
		nombre    string
		metodo    string
		ruta      string
		cuerpo    string
		conCuerpo bool
	}{
		{
			nombre:    "asignar historia",
			metodo:    http.MethodPost,
			ruta:      "/api/proyectos/1/sprints/1/historias",
			cuerpo:    `{"historia_id":1}`,
			conCuerpo: true,
		},
		{
			nombre: "desasignar historia",
			metodo: http.MethodDelete,
			ruta:   "/api/proyectos/1/sprints/1/historias/1",
		},
		{
			nombre: "consultar sprint backlog",
			metodo: http.MethodGet,
			ruta:   "/api/proyectos/1/sprints/1/historias",
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
			req.Header.Set("Authorization", "Bearer "+tokenFirmadoEnElRouter(t, secretoDePruebaRouter))
			router.ServeHTTP(w, req)

			assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code)
			assert.NotEqual(t, http.StatusNotFound, w.Code, "la ruta debe estar registrada")
		})
	}
}

// TestRutasDeDefectosAlcanzanElServicio verifica que las rutas de US-09 estén
// registradas y que el router inyecte el servicio de defectos.
func TestRutasDeDefectosAlcanzanElServicio(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter(servicioQueFalla{}, servicioHistoriasQueFalla{}, servicioSprintsQueFalla{}, nil, nil, servicioDefectosQueFalla{}, nil, nil, nil, secretoDePruebaRouter)

	casos := []struct {
		nombre    string
		metodo    string
		ruta      string
		cuerpo    string
		conCuerpo bool
	}{
		{
			nombre:    "alta de defecto",
			metodo:    http.MethodPost,
			ruta:      "/api/proyectos/1/defectos",
			cuerpo:    `{"descripcion":"Error","severidad":"major","historia_id":1,"sprint_deteccion_id":1}`,
			conCuerpo: true,
		},
		{
			nombre: "listado de defectos",
			metodo: http.MethodGet,
			ruta:   "/api/proyectos/1/defectos",
		},
		{
			nombre:    "actualizacion de defecto",
			metodo:    http.MethodPatch,
			ruta:      "/api/proyectos/1/defectos/1",
			cuerpo:    `{"estado":"in_progress"}`,
			conCuerpo: true,
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
			req.Header.Set("Authorization", "Bearer "+tokenFirmadoEnElRouter(t, secretoDePruebaRouter))
			router.ServeHTTP(w, req)

			assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code)
			assert.Equal(t, http.StatusNotFound, w.Code, "la ruta debe alcanzar el servicio")

			var cuerpo struct {
				Error string `json:"error"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cuerpo))
			assert.Equal(t, "DEFECTO_NO_ENCONTRADO", cuerpo.Error)
		})
	}
}

// TestRutasDeCriteriosDeAceptacionAlcanzanElServicio verifica que las rutas de
// US-07 estén registradas y que el router inyecte el servicio de criterios.
func TestRutasDeCriteriosDeAceptacionAlcanzanElServicio(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter(servicioQueFalla{}, servicioHistoriasQueFalla{}, servicioSprintsQueFalla{}, nil, nil, nil, servicioCriteriosQueFalla{}, nil, nil, secretoDePruebaRouter)

	casos := []struct {
		nombre    string
		metodo    string
		ruta      string
		cuerpo    string
		conCuerpo bool
	}{
		{
			nombre:    "alta de criterio",
			metodo:    http.MethodPost,
			ruta:      "/api/proyectos/1/historias/1/criterios",
			cuerpo:    `{"descripcion":"El usuario ve el saldo"}`,
			conCuerpo: true,
		},
		{
			nombre: "listado de criterios",
			metodo: http.MethodGet,
			ruta:   "/api/proyectos/1/historias/1/criterios",
		},
		{
			nombre:    "actualizacion de criterio",
			metodo:    http.MethodPatch,
			ruta:      "/api/proyectos/1/historias/1/criterios/1",
			cuerpo:    `{"is_met":true}`,
			conCuerpo: true,
		},
		{
			nombre: "eliminacion de criterio",
			metodo: http.MethodDelete,
			ruta:   "/api/proyectos/1/historias/1/criterios/1",
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
			req.Header.Set("Authorization", "Bearer "+tokenFirmadoEnElRouter(t, secretoDePruebaRouter))
			router.ServeHTTP(w, req)

			assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code)
			assert.Equal(t, http.StatusNotFound, w.Code, "la ruta debe alcanzar el servicio")

			var cuerpo struct {
				Error string `json:"error"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cuerpo))
			assert.Equal(t, "CRITERIO_NO_ENCONTRADO", cuerpo.Error)
		})
	}
}

// TestRutasDePokerAlcanzanElServicio verifica que las cuatro rutas de US-14
// esten registradas y que el router inyecte el servicio de Planning Poker. Una
// peticion con token valido alcanza el servicio y su ErrSesionPokerNoEncontrada
// vuelve como 404 SESION_POKER_NO_ENCONTRADA; un servicio nil se veria como 500.
// El caso "activas" ademas prueba que la ruta estatica no colisiona con el
// parametro ":sesionId".
func TestRutasDePokerAlcanzanElServicio(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter(servicioQueFalla{}, servicioHistoriasQueFalla{}, servicioSprintsQueFalla{}, nil, nil, nil, nil, &servicioPokerQueFalla{}, nil, secretoDePruebaRouter)

	casos := []struct {
		nombre string
		metodo string
		ruta   string
		cuerpo string
	}{
		{
			nombre: "creacion de sesion",
			metodo: http.MethodPost,
			ruta:   "/api/proyectos/1/historias/5/poker-sessions",
		},
		{
			nombre: "historial de la historia",
			metodo: http.MethodGet,
			ruta:   "/api/proyectos/1/historias/5/poker-sessions",
		},
		{
			nombre: "sesiones activas del proyecto",
			metodo: http.MethodGet,
			ruta:   "/api/proyectos/1/poker-sessions/activas",
		},
		{
			nombre: "detalle de la sesion",
			metodo: http.MethodGet,
			ruta:   "/api/proyectos/1/poker-sessions/9",
		},
		{
			nombre: "voto de la sesion (US-15)",
			metodo: http.MethodPost,
			ruta:   "/api/proyectos/1/poker-sessions/9/votos",
			cuerpo: `{"puntos": 5}`,
		},
		{
			nombre: "revelar la sesion (US-16)",
			metodo: http.MethodPost,
			ruta:   "/api/proyectos/1/poker-sessions/9/revelar",
		},
		{
			nombre: "nueva ronda (US-16)",
			metodo: http.MethodPost,
			ruta:   "/api/proyectos/1/poker-sessions/9/rondas",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			w := pedirProtegidoConCuerpoEnRouter(t, router, caso.metodo, caso.ruta, caso.cuerpo)

			assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code,
				"la ruta no debe colisionar con otra: %s %s", caso.metodo, caso.ruta)

			var cuerpo struct {
				Error   string `json:"error"`
				Mensaje string `json:"mensaje"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cuerpo),
				"la respuesta debe ser un cuerpo de error de la API: %s", w.Body.String())

			assert.Equal(t, "SESION_POKER_NO_ENCONTRADA", cuerpo.Error,
				"el handler debe alcanzar el servicio inyectado: %s %s", caso.metodo, caso.ruta)
			assert.Equal(t, domain.ErrSesionPokerNoEncontrada.Error(), cuerpo.Mensaje)
		})
	}
}

// pedirProtegidoEnRouter ejecuta una peticion con un token valido del router.
func pedirProtegidoEnRouter(t *testing.T, router *gin.Engine, metodo, ruta string) *httptest.ResponseRecorder {
	t.Helper()
	return pedirProtegidoConCuerpoEnRouter(t, router, metodo, ruta, "")
}

// pedirProtegidoConCuerpoEnRouter es como pedirProtegidoEnRouter pero envia un
// cuerpo JSON, para las rutas que lo exigen (por ejemplo el voto de US-15).
func pedirProtegidoConCuerpoEnRouter(t *testing.T, router *gin.Engine, metodo, ruta, cuerpo string) *httptest.ResponseRecorder {
	t.Helper()
	var lector io.Reader
	if cuerpo != "" {
		lector = strings.NewReader(cuerpo)
	}
	req, err := http.NewRequest(metodo, ruta, lector)
	require.NoError(t, err)
	if cuerpo != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+tokenFirmadoEnElRouter(t, secretoDePruebaRouter))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// TestAccesoE05BloqueoDeMembresiaEnElRouter verifica que el guard de US-F10
// quedo registrado en setupRouter, justo despues del middleware de
// autenticacion y antes de las rutas de proyectos: quien no es miembro no
// llega a ningun handler del proyecto (E-05, E-07), el listado no pasa por el
// guard (E-11) y un id malformado lo resuelve el handler con 400 (E-12).
func TestAccesoE05BloqueoDeMembresiaEnElRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("E-05 detalle de un proyecto ajeno", func(t *testing.T) {
		acceso := &servicioSinMembresia{}
		router := setupRouter(acceso, servicioHistoriasQueFalla{}, servicioSprintsQueFalla{}, nil, nil, nil, nil, nil, nil, secretoDePruebaRouter)

		w := pedirProtegidoEnRouter(t, router, http.MethodGet, "/api/proyectos/1")

		require.Equal(t, http.StatusNotFound, w.Code)
		var cuerpo struct {
			Error   string `json:"error"`
			Mensaje string `json:"mensaje"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cuerpo))
		assert.Equal(t, "PROYECTO_NO_ENCONTRADO", cuerpo.Error)
		assert.Equal(t, domain.ErrProyectoNoEncontrado.Error(), cuerpo.Mensaje)
		assert.False(t, acceso.proyectoAlcanzado, "el handler del proyecto no debe ejecutarse")
		assert.Equal(t, 1, acceso.consultas, "el guard consulta la membresia una vez")
	})

	t.Run("E-07 ruta anidada de un proyecto ajeno", func(t *testing.T) {
		acceso := &servicioSinMembresia{}
		router := setupRouter(acceso, servicioHistoriasQueFalla{}, servicioSprintsQueFalla{}, nil, nil, nil, nil, nil, nil, secretoDePruebaRouter)

		w := pedirProtegidoEnRouter(t, router, http.MethodGet, "/api/proyectos/1/historias")

		require.Equal(t, http.StatusNotFound, w.Code)
		var cuerpo struct {
			Error string `json:"error"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cuerpo))
		assert.Equal(t, "PROYECTO_NO_ENCONTRADO", cuerpo.Error,
			"el guard responde antes que el servicio de historias")
		assert.Equal(t, 1, acceso.consultas)
	})

	t.Run("E-15 poker de un proyecto ajeno", func(t *testing.T) {
		casos := []struct {
			nombre string
			metodo string
			ruta   string
		}{
			{
				nombre: "el no miembro no crea",
				metodo: http.MethodPost,
				ruta:   "/api/proyectos/1/historias/5/poker-sessions",
			},
			{
				nombre: "el no miembro no consulta",
				metodo: http.MethodGet,
				ruta:   "/api/proyectos/1/poker-sessions/activas",
			},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				acceso := &servicioSinMembresia{}
				poker := &servicioPokerQueFalla{}
				router := setupRouter(acceso, servicioHistoriasQueFalla{}, servicioSprintsQueFalla{}, nil, nil, nil, nil, poker, nil, secretoDePruebaRouter)

				w := pedirProtegidoEnRouter(t, router, caso.metodo, caso.ruta)

				require.Equal(t, http.StatusNotFound, w.Code)
				var cuerpo struct {
					Error string `json:"error"`
				}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cuerpo))
				assert.Equal(t, "PROYECTO_NO_ENCONTRADO", cuerpo.Error,
					"el guard responde antes que el servicio de poker")
				assert.False(t, poker.alcanzado, "el servicio de poker no debe ejecutarse")
				assert.Equal(t, 1, acceso.consultas)
			})
		}
	})

	t.Run("E-11 el listado no pasa por el guard", func(t *testing.T) {
		acceso := &servicioSinMembresia{}
		router := setupRouter(acceso, servicioHistoriasQueFalla{}, servicioSprintsQueFalla{}, nil, nil, nil, nil, nil, nil, secretoDePruebaRouter)

		w := pedirProtegidoEnRouter(t, router, http.MethodGet, "/api/proyectos")

		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 0, acceso.consultas, "sin :id de proyecto el guard no interviene")
		assert.True(t, acceso.listadoAlcanzado, "el listado debe llegar a su servicio")
	})

	t.Run("E-12 id malformado lo resuelve el handler", func(t *testing.T) {
		acceso := &servicioSinMembresia{}
		router := setupRouter(acceso, servicioHistoriasQueFalla{}, servicioSprintsQueFalla{}, nil, nil, nil, nil, nil, nil, secretoDePruebaRouter)

		w := pedirProtegidoEnRouter(t, router, http.MethodGet, "/api/proyectos/abc")

		require.Equal(t, http.StatusBadRequest, w.Code)
		var cuerpo struct {
			Error string `json:"error"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cuerpo))
		assert.Equal(t, "PARAMETRO_INVALIDO", cuerpo.Error)
		assert.Equal(t, 0, acceso.consultas, "el guard no consulta membresia con un id malformado")
	})
}

// TestPermisosDeLiderEnElRouter verifica US-11 sobre las rutas reales: un
// miembro no lider recibe 403 (E-09), un no miembro sigue recibiendo 404 y el
// guard de lider ni siquiera se consulta (E-11), y la auto-baja de un miembro
// llega al servicio (E-07).
func TestPermisosDeLiderEnElRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("E-09 el miembro no lider no elimina el proyecto", func(t *testing.T) {
		miembro := &servicioMiembro{}
		router := setupRouter(miembro, servicioHistoriasQueFalla{}, servicioSprintsQueFalla{}, nil, nil, nil, nil, nil, nil, secretoDePruebaRouter)

		w := pedirProtegidoEnRouter(t, router, http.MethodDelete, "/api/proyectos/1")

		require.Equal(t, http.StatusForbidden, w.Code)
		var cuerpo struct {
			Error string `json:"error"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cuerpo))
		assert.Equal(t, "ACCESO_DENEGADO", cuerpo.Error)
		assert.False(t, miembro.eliminado, "la operacion de lider no debe alcanzar el servicio")
		assert.Equal(t, 1, miembro.consultasDeRol)
	})

	t.Run("E-11 el no miembro recibe 404 antes que 403", func(t *testing.T) {
		acceso := &servicioSinMembresia{}
		router := setupRouter(acceso, servicioHistoriasQueFalla{}, servicioSprintsQueFalla{}, nil, nil, nil, nil, nil, nil, secretoDePruebaRouter)

		w := pedirProtegidoEnRouter(t, router, http.MethodDelete, "/api/proyectos/1")

		require.Equal(t, http.StatusNotFound, w.Code)
		var cuerpo struct {
			Error string `json:"error"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cuerpo))
		assert.Equal(t, "PROYECTO_NO_ENCONTRADO", cuerpo.Error)
	})

	t.Run("E-05 el miembro trabaja el contenido del proyecto", func(t *testing.T) {
		miembro := &servicioMiembro{}
		router := setupRouter(miembro, servicioHistoriasQueFalla{}, servicioSprintsQueFalla{}, nil, nil, nil, nil, nil, nil, secretoDePruebaRouter)

		w := pedirProtegidoEnRouter(t, router, http.MethodGet, "/api/proyectos/1/historias")

		assert.NotEqual(t, http.StatusForbidden, w.Code, "RB-04: el contenido no exige LIDER")
		assert.Equal(t, 0, miembro.consultasDeRol, "el guard de lider no interviene en el contenido")
	})
}
