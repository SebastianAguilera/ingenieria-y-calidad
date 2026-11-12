package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

// dobleAcceso implementa ServicioAcceso para los tests del guard: registra
// cuantas veces se lo invoca y con que argumentos.
type dobleAcceso struct {
	pertenece  bool
	err        error
	llamadas   int
	proyectoID uint
	usuarioID  uint
}

func (d *dobleAcceso) PerteneceA(_ context.Context, proyectoID, usuarioID uint) (bool, error) {
	d.llamadas++
	d.proyectoID = proyectoID
	d.usuarioID = usuarioID
	return d.pertenece, d.err
}

// nuevoRouterConGuard arma el guard de membresia delante de rutas reales del
// proyecto (con y sin :id), con el middleware de identidad opcional: simula el
// orden de setupRouter (identidad -> guard -> handler).
func nuevoRouterConGuard(t *testing.T, acceso ServicioAcceso, conIdentidad bool, identidad uint) (*gin.Engine, *int) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	ejecuciones := 0
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if conIdentidad {
			c.Set(claveContextoUsuario, identidad)
		}
		c.Next()
	})
	r.Use(ExigirAccesoAProyecto(acceso))

	// Las rutas con :id arrancan con la misma validacion sintactica que el
	// handler real, para poder verificar que el guard no intercepta un id
	// malformado (RB-06 / E-12).
	conID := func(c *gin.Context) {
		if _, ok := parsearID(c, "id"); !ok {
			return
		}
		ejecuciones++
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
	sinID := func(c *gin.Context) {
		ejecuciones++
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}

	r.GET("/api/proyectos", sinID)
	r.POST("/api/proyectos", sinID)
	r.GET("/api/proyectos/:id", conID)
	r.PUT("/api/proyectos/:id", conID)
	r.PATCH("/api/proyectos/:id/estado", conID)
	r.DELETE("/api/proyectos/:id", conID)
	r.GET("/api/proyectos/:id/historias", conID)
	r.GET("/api/proyectos/:id/sprints", conID)
	r.POST("/api/proyectos/:id/usuarios", conID)
	r.DELETE("/api/proyectos/:id/usuarios/:usuarioId", conID)
	r.GET("/api/usuarios/:id", conID)
	return r, &ejecuciones
}

// pedirRuta ejecuta la peticion contra el router de prueba.
func pedirRuta(t *testing.T, router *gin.Engine, metodo, ruta string) *httptest.ResponseRecorder {
	t.Helper()
	req, err := http.NewRequest(metodo, ruta, nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// cuerpoDeError extrae el codigo y el mensaje de una respuesta de error.
func cuerpoDeError(t *testing.T, w *httptest.ResponseRecorder) (string, string) {
	t.Helper()
	var cuerpo struct {
		Error   string `json:"error"`
		Mensaje string `json:"mensaje"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cuerpo),
		"la respuesta debe ser un cuerpo de error de la API: %s", w.Body.String())
	return cuerpo.Error, cuerpo.Mensaje
}

// TestAccesoE05ProyectoAjenoResponde404 cubre E-05, E-06, E-07, E-08, E-09,
// CA-5 y CA-6 (US-F10): quien no es miembro recibe el mismo 404 que un
// proyecto inexistente en el detalle y en cualquier ruta anidada, sin que el
// handler del recurso se ejecute.
func TestAccesoE05ProyectoAjenoResponde404(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre  string
		metodo  string
		ruta    string
		escenar string
	}{
		{nombre: "detalle del proyecto", metodo: http.MethodGet, ruta: "/api/proyectos/1", escenar: "E-05"},
		{nombre: "modificacion del proyecto", metodo: http.MethodPut, ruta: "/api/proyectos/1", escenar: "E-06"},
		{nombre: "cambio de estado", metodo: http.MethodPatch, ruta: "/api/proyectos/1/estado", escenar: "E-06"},
		{nombre: "baja del proyecto", metodo: http.MethodDelete, ruta: "/api/proyectos/1", escenar: "E-06"},
		{nombre: "historias anidadas", metodo: http.MethodGet, ruta: "/api/proyectos/1/historias", escenar: "E-07"},
		{nombre: "sprints anidados", metodo: http.MethodGet, ruta: "/api/proyectos/1/sprints", escenar: "E-07"},
		{nombre: "auto-asociacion", metodo: http.MethodPost, ruta: "/api/proyectos/1/usuarios", escenar: "E-08"},
		{nombre: "desvinculacion de otro", metodo: http.MethodDelete, ruta: "/api/proyectos/1/usuarios/1", escenar: "E-09"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()

			acceso := &dobleAcceso{pertenece: false}
			router, ejecuciones := nuevoRouterConGuard(t, acceso, true, 7)

			w := pedirRuta(t, router, c.metodo, c.ruta)

			require.Equal(t, http.StatusNotFound, w.Code, "%s / CA-5", c.escenar)
			codigo, mensaje := cuerpoDeError(t, w)
			assert.Equal(t, "PROYECTO_NO_ENCONTRADO", codigo)
			assert.Equal(t, domain.ErrProyectoNoEncontrado.Error(), mensaje,
				"la respuesta debe ser identica a la de un proyecto inexistente")
			assert.Equal(t, 0, *ejecuciones, "el handler del recurso no debe ejecutarse")
			assert.Equal(t, 1, acceso.llamadas, "el guard consulta la membresia una vez")
			assert.Equal(t, uint(1), acceso.proyectoID)
			assert.Equal(t, uint(7), acceso.usuarioID)
		})
	}
}

// TestAccesoE07MiembroPasaElGuard cubre E-07, E-10, E-13 y CA-6 (US-F10):
// siendo miembro, el guard deja pasar y el handler del recurso responde
// normalmente, con la identidad tomada del contexto.
func TestAccesoE07MiembroPasaElGuard(t *testing.T) {
	t.Parallel()

	acceso := &dobleAcceso{pertenece: true}
	router, ejecuciones := nuevoRouterConGuard(t, acceso, true, 3)

	w := pedirRuta(t, router, http.MethodGet, "/api/proyectos/7/historias")

	require.Equal(t, http.StatusOK, w.Code, "E-07 / CA-6")
	assert.Equal(t, 1, *ejecuciones)
	assert.Equal(t, 1, acceso.llamadas)
	assert.Equal(t, uint(7), acceso.proyectoID)
	assert.Equal(t, uint(3), acceso.usuarioID)
}

// TestAccesoE11RutasSinIdDeProyectoNoPasanPorElGuard cubre E-11 y CA-7
// (US-F10): el guard solo actua sobre rutas con :id de proyecto; el listado,
// el alta y cualquier otro recurso con :id (usuarios) no consultan membresia.
func TestAccesoE11RutasSinIdDeProyectoNoPasanPorElGuard(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		metodo string
		ruta   string
	}{
		{nombre: "listado de proyectos", metodo: http.MethodGet, ruta: "/api/proyectos"},
		{nombre: "alta de proyectos", metodo: http.MethodPost, ruta: "/api/proyectos"},
		{nombre: "detalle de usuario (no es proyecto)", metodo: http.MethodGet, ruta: "/api/usuarios/9"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()

			acceso := &dobleAcceso{pertenece: false}
			router, ejecuciones := nuevoRouterConGuard(t, acceso, true, 3)

			w := pedirRuta(t, router, c.metodo, c.ruta)

			require.Equal(t, http.StatusOK, w.Code, "E-11 / CA-7")
			assert.Equal(t, 1, *ejecuciones)
			assert.Equal(t, 0, acceso.llamadas, "el guard no debe intervenir")
		})
	}
}

// TestAccesoE12IdMalformadoLoResuelveElHandler cubre E-12, CA-7 y RB-06
// (US-F10): un :id no numerico no lo intercepta el guard: llega al handler que
// responde 400 PARAMETRO_INVALIDO.
func TestAccesoE12IdMalformadoLoResuelveElHandler(t *testing.T) {
	t.Parallel()

	acceso := &dobleAcceso{pertenece: false}
	router, ejecuciones := nuevoRouterConGuard(t, acceso, true, 3)

	w := pedirRuta(t, router, http.MethodGet, "/api/proyectos/abc")

	require.Equal(t, http.StatusBadRequest, w.Code, "E-12 / CA-7")
	codigo, _ := cuerpoDeError(t, w)
	assert.Equal(t, "PARAMETRO_INVALIDO", codigo)
	assert.Equal(t, 0, *ejecuciones, "el id invalido ni siquiera llega al handler")
	assert.Equal(t, 0, acceso.llamadas, "el guard no consulta membresia con un id malformado")
}

// TestAccesoSinIdentidadResponde401 cubre CA-1 (US-F10): si por algun motivo
// el contexto no trae el usuario autenticado, el guard responde 401 TOKEN_INVALIDO
// en lugar de consultar la membresia con un id inventado.
func TestAccesoSinIdentidadResponde401(t *testing.T) {
	t.Parallel()

	acceso := &dobleAcceso{pertenece: true}
	router, ejecuciones := nuevoRouterConGuard(t, acceso, false, 0)

	w := pedirRuta(t, router, http.MethodGet, "/api/proyectos/1")

	require.Equal(t, http.StatusUnauthorized, w.Code, "CA-1")
	codigo, _ := cuerpoDeError(t, w)
	assert.Equal(t, "TOKEN_INVALIDO", codigo)
	assert.Equal(t, 0, *ejecuciones)
	assert.Equal(t, 0, acceso.llamadas, "sin identidad no se consulta la membresia")
}

// TestAccesoErrorDePersistenciaResponde500 cubre la condicion de error del
// guard (US-F10): si la consulta de membresia falla, la respuesta es un error
// interno y la ruta no se ejecuta.
func TestAccesoErrorDePersistenciaResponde500(t *testing.T) {
	t.Parallel()

	acceso := &dobleAcceso{pertenece: false, err: fmt.Errorf("connection refused")}
	router, ejecuciones := nuevoRouterConGuard(t, acceso, true, 3)

	w := pedirRuta(t, router, http.MethodGet, "/api/proyectos/1")

	require.Equal(t, http.StatusInternalServerError, w.Code)
	codigo, _ := cuerpoDeError(t, w)
	assert.Equal(t, "ERROR_INTERNO", codigo)
	assert.Equal(t, 0, *ejecuciones)
	assert.Equal(t, 1, acceso.llamadas)
}
