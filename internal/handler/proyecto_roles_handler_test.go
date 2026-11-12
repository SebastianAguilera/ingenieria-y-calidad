package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

// Pruebas de US-11: permisos por rol. usuarioRouterDePrueba (1) es el usuario
// autenticado.

func TestMiembroNoPuedeEliminarNiGestionarElEquipo(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		metodo string
		ruta   string
		cuerpo any
	}{
		{"E-09 eliminar el proyecto", http.MethodDelete, "/api/proyectos/1", nil},
		{"E-10 agregar un integrante", http.MethodPost, "/api/proyectos/1/usuarios", map[string]any{"usuario_id": 7, "rol": "MIEMBRO"}},
		{"E-10 quitar a otro integrante", http.MethodDelete, "/api/proyectos/1/usuarios/4", nil},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			// Sin expectativas de Eliminar/AgregarUsuario/QuitarUsuario: si el guard
			// dejara pasar el request, el mock fallaria el test.
			router, _ := nuevoRouterDeProyectosConRol(t, usuarioRouterDePrueba, domain.RolMiembro)

			w := ejecutar(t, router, c.metodo, c.ruta, c.cuerpo)

			require.Equal(t, http.StatusForbidden, w.Code, "cuerpo: %s", w.Body.String())
			assert.Equal(t, "ACCESO_DENEGADO", decodificarError(t, w).Error)
		})
	}
}

func TestLiderEliminaYGestionaElEquipo(t *testing.T) {
	t.Parallel()

	t.Run("E-03 elimina el proyecto", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDeProyectosConRol(t, usuarioRouterDePrueba, domain.RolLider)
		servicio.On("Eliminar", mock.Anything, uint(1)).Return(nil)

		w := ejecutar(t, router, http.MethodDelete, "/api/proyectos/1", nil)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("E-04 quita a otro integrante", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDeProyectosConRol(t, usuarioRouterDePrueba, domain.RolLider)
		servicio.On("QuitarUsuario", mock.Anything, uint(1), uint(4)).Return(nil)

		w := ejecutar(t, router, http.MethodDelete, "/api/proyectos/1/usuarios/4", nil)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("E-02 agrega un integrante sin rol y lo deja al servicio como MIEMBRO", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDeProyectosConRol(t, usuarioRouterDePrueba, domain.RolLider)
		servicio.On("AgregarUsuario", mock.Anything, uint(1), domain.AsociacionUsuario{UsuarioID: 7}).
			Return(&domain.Usuario{ID: 7, Nombre: "Cris", Email: "cris@utn.edu.ar"}, nil)

		w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/usuarios", map[string]any{"usuario_id": 7})

		assert.Equal(t, http.StatusOK, w.Code, "cuerpo: %s", w.Body.String())
	})

	t.Run("E-14 rol invalido responde 422", func(t *testing.T) {
		t.Parallel()
		router, _ := nuevoRouterDeProyectosConRol(t, usuarioRouterDePrueba, domain.RolLider)

		w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/usuarios",
			map[string]any{"usuario_id": 7, "rol": "Scrum Master"})

		require.Equal(t, http.StatusUnprocessableEntity, w.Code)
		respuesta := decodificarError(t, w)
		assert.Equal(t, "VALIDACION", respuesta.Error)
		assert.Equal(t, "rol", respuesta.Campo)
	})
}

func TestMiembroPuedeQuitarseASiMismo(t *testing.T) {
	t.Parallel()

	router, servicio := nuevoRouterDeProyectosConRol(t, usuarioRouterDePrueba, domain.RolMiembro)
	servicio.On("QuitarUsuario", mock.Anything, uint(1), usuarioRouterDePrueba).Return(nil)

	w := ejecutar(t, router, http.MethodDelete, "/api/proyectos/1/usuarios/1", nil)

	assert.Equal(t, http.StatusNoContent, w.Code, "E-07: la auto-baja no exige LIDER")
}

func TestMiembroNoPuedeEliminarElProyectoAunqueSeaSuId(t *testing.T) {
	t.Parallel()

	// La excepcion de auto-baja es solo para DELETE .../usuarios/{su id}.
	router, _ := nuevoRouterDeProyectosConRol(t, usuarioRouterDePrueba, domain.RolMiembro)

	w := ejecutar(t, router, http.MethodDelete, "/api/proyectos/1", nil)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestQuitarAlUltimoLiderResponde409(t *testing.T) {
	t.Parallel()

	router, servicio := nuevoRouterDeProyectosConRol(t, usuarioRouterDePrueba, domain.RolLider)
	servicio.On("QuitarUsuario", mock.Anything, uint(1), usuarioRouterDePrueba).Return(domain.ErrUltimoLider)

	w := ejecutar(t, router, http.MethodDelete, "/api/proyectos/1/usuarios/1", nil)

	require.Equal(t, http.StatusConflict, w.Code, "E-12")
	assert.Equal(t, "ULTIMO_LIDER", decodificarError(t, w).Error)
}

func TestGuardDeLiderSinIdentidadResponde401(t *testing.T) {
	t.Parallel()

	router, _ := nuevoRouterDeProyectosConIdentidad(t, 0)

	w := ejecutar(t, router, http.MethodDelete, "/api/proyectos/1", nil)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "TOKEN_INVALIDO", decodificarError(t, w).Error)
}

func TestGuardDeLiderConIdInvalidoDejaQueElHandlerResponda400(t *testing.T) {
	t.Parallel()

	router, _ := nuevoRouterDeProyectosConRol(t, usuarioRouterDePrueba, domain.RolMiembro)

	w := ejecutar(t, router, http.MethodDelete, "/api/proyectos/abc", nil)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "PARAMETRO_INVALIDO", decodificarError(t, w).Error)
}

func TestGuardDeLiderPropagaFallosDePersistenciaComo500(t *testing.T) {
	t.Parallel()

	router, servicio := nuevoRouterDeProyectosConIdentidad(t, usuarioRouterDePrueba)
	servicio.ExpectedCalls = nil
	servicio.On("EsLider", mock.Anything, uint(1), usuarioRouterDePrueba).Return(false, errors.New("timeout"))

	w := ejecutar(t, router, http.MethodDelete, "/api/proyectos/1", nil)

	require.Equal(t, http.StatusInternalServerError, w.Code, "ante un fallo no se abre la operacion")
	assert.Equal(t, "ERROR_INTERNO", decodificarError(t, w).Error)
}

func TestDetalleIncluyeElRolActual(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		rol    domain.RolProyecto
	}{
		{"E-06 miembro", domain.RolMiembro},
		{"lider", domain.RolLider},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			router, servicio := nuevoRouterDeProyectosConRol(t, usuarioRouterDePrueba, c.rol)
			servicio.On("ObtenerPorID", mock.Anything, uint(1)).
				Return(&domain.Proyecto{ID: 1, Nombre: "Metrics", FechaInicio: fechaInicioHTTP, Estado: domain.EstadoActivo}, nil)

			w := ejecutar(t, router, http.MethodGet, "/api/proyectos/1", nil)

			require.Equal(t, http.StatusOK, w.Code)
			var cuerpo map[string]any
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cuerpo))
			assert.Equal(t, string(c.rol), cuerpo["rol_actual"])
		})
	}
}

func TestListadoNoIncluyeRolActual(t *testing.T) {
	t.Parallel()

	router, servicio := nuevoRouterDePrueba(t)
	servicio.On("Listar", mock.Anything, usuarioRouterDePrueba).
		Return([]domain.Proyecto{{ID: 1, Nombre: "Metrics", FechaInicio: fechaInicioHTTP, Estado: domain.EstadoActivo}}, nil)

	w := ejecutar(t, router, http.MethodGet, "/api/proyectos", nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), "rol_actual")
}

func TestDetalleSinIdentidadResponde401(t *testing.T) {
	t.Parallel()

	router, _ := nuevoRouterDeProyectosConIdentidad(t, 0)

	w := ejecutar(t, router, http.MethodGet, "/api/proyectos/1", nil)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMapeoDeErroresDeRoles(t *testing.T) {
	t.Parallel()

	casos := []struct {
		err    error
		status int
		codigo string
	}{
		{domain.ErrAccesoDenegado, http.StatusForbidden, "ACCESO_DENEGADO"},
		{domain.ErrUltimoLider, http.StatusConflict, "ULTIMO_LIDER"},
	}

	for _, c := range casos {
		t.Run(c.codigo, func(t *testing.T) {
			t.Parallel()
			mapeo, ok := buscarMapeo(c.err, contextoRaiz)
			require.True(t, ok)
			assert.Equal(t, c.status, mapeo.estadoHTTP)
			assert.Equal(t, c.codigo, mapeo.codigo)
			assert.NotEmpty(t, mapeo.mensaje)
		})
	}
}
