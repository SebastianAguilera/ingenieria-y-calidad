package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

// secretoDePruebaRouter equivale a JWT_SECRET para el wiring de US-10.
const secretoDePruebaRouter = "secreto-de-prueba-router"

// secuenciaDeTokens garantiza que cada login emita un token distinto (E-16).
var secuenciaDeTokens int64

// servicioAuthPrueba implementa handler.ServicioAuth emitiendo tokens HS256
// firmados con el secreto de prueba, igual que el servicio real.
type servicioAuthPrueba struct{}

func (servicioAuthPrueba) Login(_ context.Context, email, _ string) (string, *domain.Usuario, error) {
	ahora := time.Now()
	claims := jwt.MapClaims{
		"sub":   "1",
		"email": email,
		"iat":   ahora.Unix(),
		"exp":   ahora.Add(24 * time.Hour).Unix(),
		"jti":   atomic.AddInt64(&secuenciaDeTokens, 1),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	firmado, err := token.SignedString([]byte(secretoDePruebaRouter))
	if err != nil {
		return "", nil, err
	}
	return firmado, &domain.Usuario{
		ID:       1,
		Nombre:   "Ana Pérez",
		Email:    email,
		CreadoEn: ahora,
	}, nil
}

// Registrar emite el mismo token que Login: el doble de wiring de US-F09
// verifica que la ruta sea publica y responda 200 sin Authorization.
func (servicioAuthPrueba) Registrar(ctx context.Context, _, email, password string) (string, *domain.Usuario, error) {
	if password == "" {
		return "", nil, domain.ErrPasswordVacio
	}
	return servicioAuthPrueba{}.Login(ctx, email, password)
}

// servicioProyectosVacio implementa handler.ServicioProyectos devolviendo un
// listado vacio: permite verificar que con un token valido la ruta responde
// 200 con su resultado normal (E-10 / CA-6).
type servicioProyectosVacio struct{}

func (servicioProyectosVacio) Crear(context.Context, uint, domain.NuevoProyecto) (*domain.Proyecto, error) {
	return nil, domain.ErrProyectoNoEncontrado
}

func (servicioProyectosVacio) Listar(context.Context, uint) ([]domain.Proyecto, error) {
	return []domain.Proyecto{}, nil
}

func (servicioProyectosVacio) ObtenerPorID(context.Context, uint) (*domain.Proyecto, error) {
	return nil, domain.ErrProyectoNoEncontrado
}

func (servicioProyectosVacio) Actualizar(context.Context, uint, domain.ActualizacionProyecto) (*domain.Proyecto, error) {
	return nil, domain.ErrProyectoNoEncontrado
}

func (servicioProyectosVacio) CambiarEstado(context.Context, uint, domain.EstadoProyecto) (*domain.Proyecto, error) {
	return nil, domain.ErrProyectoNoEncontrado
}

func (servicioProyectosVacio) Eliminar(context.Context, uint) error {
	return domain.ErrProyectoNoEncontrado
}

func (servicioProyectosVacio) AgregarUsuario(context.Context, uint, domain.AsociacionUsuario) (*domain.Usuario, error) {
	return nil, domain.ErrProyectoNoEncontrado
}

func (servicioProyectosVacio) QuitarUsuario(context.Context, uint, uint) error {
	return domain.ErrProyectoNoEncontrado
}

// PerteneceA siempre true: este doble representa al equipo completo, para que
// los tests de US-10 lleguen al handler con un token valido.
func (servicioProyectosVacio) PerteneceA(context.Context, uint, uint) (bool, error) {
	return true, nil
}

// routerDeAutenticacion arma el router de produccion con los servicios de
// US-10 inyectados: lista publica explicita (GET /health y
// POST /api/auth/login) y middleware sobre el resto de /api/*.
func routerDeAutenticacion(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	return setupRouter(
		servicioProyectosVacio{},
		servicioHistoriasQueFalla{},
		servicioSprintsQueFalla{},
		nil, nil, nil, nil,
		servicioAuthPrueba{},
		secretoDePruebaRouter,
	)
}

// loginEnElRouter hace un login real contra el router y devuelve el token.
func loginEnElRouter(t *testing.T, router *gin.Engine) string {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, "/api/auth/login",
		bytes.NewReader([]byte(`{"email":"ana@example.com","password":"Secreto123"}`)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "el login debe ser publico: %s", w.Body.String())

	var respuesta struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta))
	require.NotEmpty(t, respuesta.Token)
	return respuesta.Token
}

// getProtegido llama a GET /api/proyectos con el header Authorization indicado;
// un valor vacio significa no enviar el header.
func getProtegido(t *testing.T, router *gin.Engine, authorization string) *httptest.ResponseRecorder {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, "/api/proyectos", nil)
	require.NoError(t, err)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// TestE14RutasPublicasSinToken cubre E-14 y CL-09: solo GET /health y
// POST /api/auth/login son publicas; responden normal y nunca 401 por falta
// de token.
func TestE14RutasPublicasSinToken(t *testing.T) {
	router := routerDeAutenticacion(t)

	wHealth := httptest.NewRecorder()
	reqHealth, err := http.NewRequest(http.MethodGet, "/health", nil)
	require.NoError(t, err)
	router.ServeHTTP(wHealth, reqHealth)

	wLogin := httptest.NewRecorder()
	reqLogin, err := http.NewRequest(http.MethodPost, "/api/auth/login",
		bytes.NewReader([]byte(`{"email":"ana@example.com","password":"Secreto123"}`)))
	require.NoError(t, err)
	reqLogin.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wLogin, reqLogin)

	assert.Equal(t, http.StatusOK, wHealth.Code, "E-14: /health sin Authorization")
	assert.NotContains(t, wHealth.Body.String(), "TOKEN_REQUERIDO")

	assert.Equal(t, http.StatusOK, wLogin.Code, "E-14: login sin Authorization")
	assert.NotContains(t, wLogin.Body.String(), "TOKEN_REQUERIDO")

	// El listado de proyectos, en cambio, sigue protegido.
	wProyectos := getProtegido(t, router, "")
	assert.Equal(t, http.StatusUnauthorized, wProyectos.Code, "CA-4")
}

// TestRutasProtegidasDelRouter cubre E-10, E-11 y E-13 a nivel de wiring: la
// lista publica explicita deja pasar /health y el login, el middleware rechaza
// a las rutas /api/* sin token o con un token que no sirve, y con el token del
// login la ruta responde con su listado normal (CA-6).
func TestRutasProtegidasDelRouter(t *testing.T) {
	tokenValido := loginEnElRouter(t, routerDeAutenticacion(t))

	casos := []struct {
		nombre         string
		authorization  string
		statusEsperado int
		errorEsperado  string
		escenario      string
	}{
		{
			nombre:         "sin Authorization",
			authorization:  "",
			statusEsperado: http.StatusUnauthorized,
			errorEsperado:  "TOKEN_REQUERIDO",
			escenario:      "E-11 / CA-4",
		},
		{
			nombre:         "token firmado con otra clave",
			authorization:  "Bearer " + tokenFirmadoEnElRouter(t, "otra-clave"),
			statusEsperado: http.StatusUnauthorized,
			errorEsperado:  "TOKEN_INVALIDO",
			escenario:      "E-13 / CA-5",
		},
		{
			nombre:         "token expirado",
			authorization:  "Bearer " + tokenVencidoEnElRouter(t),
			statusEsperado: http.StatusUnauthorized,
			errorEsperado:  "TOKEN_INVALIDO",
			escenario:      "E-13 / CA-5",
		},
		{
			nombre:         "token basura",
			authorization:  "Bearer basura",
			statusEsperado: http.StatusUnauthorized,
			errorEsperado:  "TOKEN_INVALIDO",
			escenario:      "E-13 / CA-5",
		},
		{
			nombre:         "token valido del login",
			authorization:  "Bearer " + tokenValido,
			statusEsperado: http.StatusOK,
			escenario:      "E-10 / CA-6",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			router := routerDeAutenticacion(t)

			w := getProtegido(t, router, c.authorization)

			require.Equal(t, c.statusEsperado, w.Code, c.escenario)
			if c.errorEsperado == "" {
				var listado struct {
					Total     int   `json:"total"`
					Proyectos []any `json:"proyectos"`
				}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listado),
					"la ruta responde su listado normal: %s", w.Body.String())
				assert.Zero(t, listado.Total)
				return
			}

			var errResp struct {
				Error   string `json:"error"`
				Mensaje string `json:"mensaje"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &errResp))
			assert.Equal(t, c.errorEsperado, errResp.Error, c.escenario)
		})
	}
}

// TestE16DosSesionesSimultaneas cubre E-16, R-07 y CL-08: sin logout
// server-side ni blacklist, dos logins producen dos tokens que siguen validos
// hasta su expiracion.
func TestE16DosSesionesSimultaneas(t *testing.T) {
	router := routerDeAutenticacion(t)

	tokenUno := loginEnElRouter(t, router)
	tokenDos := loginEnElRouter(t, router)
	require.NotEqual(t, tokenUno, tokenDos, "cada sesion recibe su propio token")

	wUno := getProtegido(t, router, "Bearer "+tokenUno)
	wDos := getProtegido(t, router, "Bearer "+tokenDos)

	assert.Equal(t, http.StatusOK, wUno.Code, "E-16: la primera sesion sigue valida")
	assert.Equal(t, http.StatusOK, wDos.Code, "E-16: la segunda sesion sigue valida")
}

// TestRegistroPublicoSinToken cubre E-09 y RB-09 de US-F09: el alta de usuario
// es una de las rutas publicas, asi que un request sin Authorization nunca
// recibe TOKEN_REQUERIDO.
func TestRegistroPublicoSinToken(t *testing.T) {
	router := routerDeAutenticacion(t)

	req, err := http.NewRequest(http.MethodPost, "/api/auth/registro",
		bytes.NewReader([]byte(`{"nombre":"Ana Pérez","email":"ana@example.com","password":"Secreto123"}`)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotContains(t, w.Body.String(), "TOKEN_REQUERIDO", "E-09 / RB-09")
	assert.Equal(t, http.StatusOK, w.Code, "E-09 / RB-09: el registro es publico y responde 200")
}

// tokenFirmadoEnElRouter emite un token HS256 con un secreto distinto al de la
// aplicacion (firma ajena, E-13).
func tokenFirmadoEnElRouter(t *testing.T, secreto string) string {
	t.Helper()
	ahora := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "1",
		"iat": ahora.Unix(),
		"exp": ahora.Add(24 * time.Hour).Unix(),
		"jti": atomic.AddInt64(&secuenciaDeTokens, 1),
	})
	firmado, err := token.SignedString([]byte(secreto))
	require.NoError(t, err)
	return firmado
}

// tokenVencidoEnElRouter emite un token firmado con el secreto correcto pero
// vencido (E-13).
func tokenVencidoEnElRouter(t *testing.T) string {
	t.Helper()
	ahora := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "1",
		"iat": ahora.Add(-2 * time.Hour).Unix(),
		"exp": ahora.Add(-time.Hour).Unix(),
		"jti": atomic.AddInt64(&secuenciaDeTokens, 1),
	})
	firmado, err := token.SignedString([]byte(secretoDePruebaRouter))
	require.NoError(t, err)
	return firmado
}
