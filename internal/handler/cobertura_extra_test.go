package handler

import (
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func TestProyectoHandlerRechazaIDsInvalidos(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		metodo string
		ruta   string
		cuerpo any
	}{
		{"detalle", http.MethodGet, "/api/proyectos/abc", nil},
		{"actualizar", http.MethodPut, "/api/proyectos/abc", map[string]any{"nombre": "X", "fecha_inicio": "2026-09-28"}},
		{"cambiar estado", http.MethodPatch, "/api/proyectos/abc/estado", map[string]any{"estado": "Activo"}},
		{"eliminar", http.MethodDelete, "/api/proyectos/abc", nil},
		{"agregar integrante", http.MethodPost, "/api/proyectos/abc/usuarios", map[string]any{"usuario_id": 1, "rol": "Constructor de Producto"}},
		{"quitar integrante", http.MethodDelete, "/api/proyectos/abc/usuarios/1", nil},
	}

	for _, c := range casos {
		c := c
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			router, _ := nuevoRouterDePrueba(t)

			w := ejecutar(t, router, c.metodo, c.ruta, c.cuerpo)

			require.Equal(t, http.StatusBadRequest, w.Code, "cuerpo: %s", w.Body.String())
			assert.Equal(t, codigoParametroInvalido, decodificarError(t, w).Error)
		})
	}
}

func TestActualizarProyectoValidacionesExtras(t *testing.T) {
	t.Parallel()

	t.Run("cuerpo JSON invalido", func(t *testing.T) {
		t.Parallel()
		router, _ := nuevoRouterDePrueba(t)

		w := ejecutar(t, router, http.MethodPut, "/api/proyectos/1", "{")

		require.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, codigoJSONInvalido, decodificarError(t, w).Error)
	})

	t.Run("fecha de inicio invalida", func(t *testing.T) {
		t.Parallel()
		router, _ := nuevoRouterDePrueba(t)

		w := ejecutar(t, router, http.MethodPut, "/api/proyectos/1", map[string]any{
			"nombre":       "Proyecto",
			"fecha_inicio": "2026-13-01",
		})

		require.Equal(t, http.StatusBadRequest, w.Code)
		respuesta := decodificarError(t, w)
		assert.Equal(t, codigoParametroInvalido, respuesta.Error)
		assert.Equal(t, "fecha_inicio", respuesta.Campo)
	})

	t.Run("error no controlado del servicio", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDePrueba(t)
		servicio.On("Actualizar", mock.Anything, uint(1), mock.Anything).Return(nil, errors.New("boom"))

		w := ejecutar(t, router, http.MethodPut, "/api/proyectos/1", map[string]any{
			"nombre":       "Proyecto",
			"fecha_inicio": "2026-09-28",
		})

		require.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, codigoErrorInterno, decodificarError(t, w).Error)
		servicio.AssertExpectations(t)
	})
}

func TestCrearProyectoErrorDelServicio(t *testing.T) {
	t.Parallel()

	router, servicio := nuevoRouterDePrueba(t)
	servicio.On("Crear", mock.Anything, mock.Anything).Return(nil, errors.New("boom"))

	w := ejecutar(t, router, http.MethodPost, "/api/proyectos", map[string]any{
		"nombre":       "Proyecto",
		"fecha_inicio": "2026-09-28",
	})

	require.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, codigoErrorInterno, decodificarError(t, w).Error)
	servicio.AssertExpectations(t)
}

func TestCrearProyectoValidacionesExtras(t *testing.T) {
	t.Parallel()

	t.Run("fecha de fin invalida", func(t *testing.T) {
		t.Parallel()
		router, _ := nuevoRouterDePrueba(t)

		w := ejecutar(t, router, http.MethodPost, "/api/proyectos", map[string]any{
			"nombre":       "Proyecto",
			"fecha_inicio": "2026-09-28",
			"fecha_fin":    "18-12-2026",
		})

		require.Equal(t, http.StatusBadRequest, w.Code)
		respuesta := decodificarError(t, w)
		assert.Equal(t, codigoParametroInvalido, respuesta.Error)
		assert.Equal(t, "fecha_fin", respuesta.Campo)
	})

	t.Run("nombre demasiado largo informado por el servicio", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDePrueba(t)
		servicio.On("Crear", mock.Anything, mock.Anything).Return(nil, domain.ErrNombreDemasiadoLargo)

		w := ejecutar(t, router, http.MethodPost, "/api/proyectos", map[string]any{
			"nombre":       "Proyecto",
			"fecha_inicio": "2026-09-28",
		})

		require.Equal(t, http.StatusUnprocessableEntity, w.Code)
		respuesta := decodificarError(t, w)
		assert.Equal(t, codigoValidacion, respuesta.Error)
		assert.Equal(t, "nombre", respuesta.Campo)
		servicio.AssertExpectations(t)
	})
}

func TestCrearSprintFechaInicioObligatoriaInformadaPorElServicio(t *testing.T) {
	t.Parallel()

	router, servicio := nuevoRouterDeSprints(t)
	servicio.On("Crear", mock.Anything, uint(1), mock.Anything).Return(nil, domain.ErrSprintFechaInicioObligatoria)

	w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/sprints", map[string]any{
		"identificador": "Sprint 1",
		"sprint_goal":   "Goal",
		"fecha_inicio":  "2026-10-06",
		"fecha_fin":     "2026-10-20",
	})

	require.Equal(t, http.StatusUnprocessableEntity, w.Code)
	respuesta := decodificarError(t, w)
	assert.Equal(t, codigoValidacion, respuesta.Error)
	assert.Equal(t, "fecha_inicio", respuesta.Campo)
	servicio.AssertExpectations(t)
}

func TestAgregarIntegranteValidacionesExtras(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		cuerpo      any
		codigo      int
		codigoError string
		campo       string
	}{
		{"cuerpo JSON invalido", "{", http.StatusBadRequest, codigoJSONInvalido, ""},
		{"usuario_id cero", map[string]any{"usuario_id": 0, "rol": "Constructor de Producto"}, http.StatusBadRequest, codigoParametroInvalido, ""},
		{"rol invalido", map[string]any{"usuario_id": 1, "rol": "Scrum Master"}, http.StatusUnprocessableEntity, codigoValidacion, "rol"},
	}

	for _, c := range casos {
		c := c
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			router, _ := nuevoRouterDePrueba(t)

			w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/usuarios", c.cuerpo)

			require.Equal(t, c.codigo, w.Code, "cuerpo: %s", w.Body.String())
			respuesta := decodificarError(t, w)
			assert.Equal(t, c.codigoError, respuesta.Error)
			assert.Equal(t, c.campo, respuesta.Campo)
		})
	}
}

func TestHistoriaHandlerErroresDeInfraestructura(t *testing.T) {
	t.Parallel()

	t.Run("error del servicio al listar", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDeHistorias(t)
		servicio.On("Listar", mock.Anything, uint(1)).Return(nil, errors.New("boom"))

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/1/historias", nil)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, codigoErrorInterno, decodificarError(t, w).Error)
		servicio.AssertExpectations(t)
	})

	t.Run("proyecto invalido al listar", func(t *testing.T) {
		t.Parallel()
		router, _ := nuevoRouterDeHistorias(t)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/abc/historias", nil)

		require.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, codigoParametroInvalido, decodificarError(t, w).Error)
	})

	t.Run("proyecto invalido al actualizar", func(t *testing.T) {
		t.Parallel()
		router, _ := nuevoRouterDeHistorias(t)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/abc/historias/1", map[string]any{"titulo": "Nuevo"})

		require.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, codigoParametroInvalido, decodificarError(t, w).Error)
	})
}

func TestSprintHandlerCambiarEstadoProyectoInvalido(t *testing.T) {
	t.Parallel()

	router, _ := nuevoRouterDeSprints(t)

	w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/abc/sprints/1/estado", map[string]any{"estado": "Activo"})

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, codigoParametroInvalido, decodificarError(t, w).Error)
}

func nuevoRouterSprintBacklog(t *testing.T) (*gin.Engine, *mockServicioSprintBacklog) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	servicio := &mockServicioSprintBacklog{}
	servicio.Mock.Test(t)
	router := gin.New()
	RegistrarRutasSprintBacklog(router, servicio)
	return router, servicio
}

func TestSprintBacklogHandlerValidaciones(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		metodo      string
		ruta        string
		cuerpo      any
		codigoError string
	}{
		{"proyecto invalido al asignar", http.MethodPost, "/api/proyectos/abc/sprints/1/historias", map[string]any{"historia_id": 1}, codigoParametroInvalido},
		{"sprint invalido al asignar", http.MethodPost, "/api/proyectos/1/sprints/abc/historias", map[string]any{"historia_id": 1}, codigoParametroInvalido},
		{"cuerpo invalido al asignar", http.MethodPost, "/api/proyectos/1/sprints/1/historias", "{", codigoJSONInvalido},
		{"proyecto invalido al desasignar", http.MethodDelete, "/api/proyectos/abc/sprints/1/historias/1", nil, codigoParametroInvalido},
		{"sprint invalido al desasignar", http.MethodDelete, "/api/proyectos/1/sprints/abc/historias/1", nil, codigoParametroInvalido},
		{"historia invalida al desasignar", http.MethodDelete, "/api/proyectos/1/sprints/1/historias/abc", nil, codigoParametroInvalido},
	}

	for _, c := range casos {
		c := c
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			router, _ := nuevoRouterSprintBacklog(t)

			w := ejecutar(t, router, c.metodo, c.ruta, c.cuerpo)

			require.Equal(t, http.StatusBadRequest, w.Code, "cuerpo: %s", w.Body.String())
			assert.Equal(t, c.codigoError, decodificarError(t, w).Error)
		})
	}

	t.Run("asigna a traves del registro de rutas", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterSprintBacklog(t)
		servicio.On("AsignarASprint", mock.Anything, uint(1), uint(10), uint(100)).Return(nil)

		w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/sprints/10/historias", map[string]any{"historia_id": 100})

		require.Equal(t, http.StatusNoContent, w.Code)
		servicio.AssertExpectations(t)
	})

	t.Run("desasigna a traves del registro de rutas", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterSprintBacklog(t)
		servicio.On("DesasignarDeSprint", mock.Anything, uint(1), uint(10), uint(100)).Return(nil)

		w := ejecutar(t, router, http.MethodDelete, "/api/proyectos/1/sprints/10/historias/100", nil)

		require.Equal(t, http.StatusNoContent, w.Code)
		servicio.AssertExpectations(t)
	})
}

func TestUsuarioHandlerErrores(t *testing.T) {
	t.Parallel()

	t.Run("cuerpo JSON invalido al crear", func(t *testing.T) {
		t.Parallel()
		router, _ := nuevoRouterUsuarios(t)

		w := ejecutarUsuarios(t, router, http.MethodPost, "/api/usuarios", "{")

		require.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, codigoJSONInvalido, decodificarError(t, w).Error)
	})

	t.Run("error del servicio al listar", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterUsuarios(t)
		servicio.On("Listar", mock.Anything).Return(nil, errors.New("boom"))

		w := ejecutarUsuarios(t, router, http.MethodGet, "/api/usuarios", nil)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, codigoErrorInterno, decodificarError(t, w).Error)
		servicio.AssertExpectations(t)
	})
}
