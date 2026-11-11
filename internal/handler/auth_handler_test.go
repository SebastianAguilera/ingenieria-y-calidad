package handler

import (
	"bytes"
	"context"
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

// mockServicioAuth es el doble de ServicioAuth usado por los escenarios de
// login (E-01 a E-09, E-18).
type mockServicioAuth struct {
	mock.Mock
}

func (m *mockServicioAuth) Login(ctx context.Context, email, password string) (string, *domain.Usuario, error) {
	ret := m.Called(ctx, email, password)
	token, _ := ret.Get(0).(string)
	var usuario *domain.Usuario
	if ret.Get(1) != nil {
		usuario = ret.Get(1).(*domain.Usuario)
	}
	return token, usuario, ret.Error(2)
}

func nuevoRouterAuth(t *testing.T, servicio ServicioAuth) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegistrarRutasAuth(r, servicio)
	return r
}

// ejecutarLogin arma un POST /api/auth/login con el content type indicado, de
// modo que se puedan cubrir el caso JSON (E-01) y el body que no es JSON
// (E-08, E-09).
func ejecutarLogin(t *testing.T, router *gin.Engine, contentType, cuerpo string) *httptest.ResponseRecorder {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte(cuerpo)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func loginJSON(email, password string) string {
	return `{"email":"` + email + `","password":"` + password + `"}`
}

// usuarioAutenticado es el usuario que el servicio devuelve en un login
// exitoso, con el hash cargado para demostrar que la capa HTTP lo descarta.
func usuarioAutenticado() *domain.Usuario {
	return &domain.Usuario{
		ID:           1,
		Nombre:       "Ana Pérez",
		Email:        "ana@example.com",
		CreadoEn:     time.Date(2026, time.November, 10, 12, 0, 0, 0, time.UTC),
		PasswordHash: "$2a$10$abcdefghijklmnopqrstuvABCDEFGHIJKLMNOPQRSTUVWXYZ01234",
	}
}

type respuestaLoginPrueba struct {
	Token   string `json:"token"`
	Usuario struct {
		ID       uint   `json:"id"`
		Nombre   string `json:"nombre"`
		Email    string `json:"email"`
		CreadoEn string `json:"creado_en"`
	} `json:"usuario"`
}

// TestLoginE01DevuelveTokenYUsuarioSinPasswordHash cubre E-01, CA-1 y CA-7 en
// la capa HTTP: 200 con token, usuario con sus cuatro campos y ninguna
// traza de password_hash en el cuerpo.
func TestLoginE01DevuelveTokenYUsuarioSinPasswordHash(t *testing.T) {
	t.Parallel()

	servicio := new(mockServicioAuth)
	servicio.On("Login", mock.Anything, "ana@example.com", "Secreto123").
		Return("jwt-de-ana", usuarioAutenticado(), nil)
	router := nuevoRouterAuth(t, servicio)

	w := ejecutarLogin(t, router, "application/json", loginJSON("ana@example.com", "Secreto123"))

	require.Equal(t, http.StatusOK, w.Code, "E-01 / CA-1")

	var respuesta respuestaLoginPrueba
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta))
	assert.Equal(t, "jwt-de-ana", respuesta.Token)
	assert.Equal(t, uint(1), respuesta.Usuario.ID)
	assert.Equal(t, "Ana Pérez", respuesta.Usuario.Nombre)
	assert.Equal(t, "ana@example.com", respuesta.Usuario.Email)
	assert.NotEmpty(t, respuesta.Usuario.CreadoEn, "el usuario incluye creado_en")

	cuerpo := w.Body.String()
	assert.NotContains(t, cuerpo, "password_hash", "CA-7")
	assert.NotContains(t, cuerpo, usuarioAutenticado().PasswordHash, "CA-7")
}

// TestLoginE03Responde401SinToken cubre E-03 y CA-2 a nivel HTTP: el error
// del servicio se traduce en 401 CREDENCIALES_INVALIDAS y la respuesta no
// incluye token.
func TestLoginE03Responde401SinToken(t *testing.T) {
	t.Parallel()

	servicio := new(mockServicioAuth)
	servicio.On("Login", mock.Anything, "ana@example.com", "OtraClave").
		Return("", nil, domain.ErrCredencialesInvalidas)
	router := nuevoRouterAuth(t, servicio)

	w := ejecutarLogin(t, router, "application/json", loginJSON("ana@example.com", "OtraClave"))

	require.Equal(t, http.StatusUnauthorized, w.Code, "E-03 / CA-2")

	var errResp respuestaError
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &errResp))
	assert.Equal(t, "CREDENCIALES_INVALIDAS", errResp.Error)
	assert.Equal(t, domain.ErrCredencialesInvalidas.Error(), errResp.Mensaje)
	assert.NotContains(t, w.Body.String(), "token",
		"la respuesta de error no lleva campo token")
}

// TestLoginE07CampoVacioResponde422 cubre E-07, R-08 y CL-02: el campo vacio
// se rechaza con VALIDACION antes de llegar al servicio, por lo que la
// contrasena jamas se compara contra un hash.
func TestLoginE07CampoVacioResponde422(t *testing.T) {
	casos := []struct {
		nombre   string
		email    string
		password string
		campo    string
		mensaje  string
	}{
		{
			nombre:   "email vacio",
			email:    "",
			password: "Secreto123",
			campo:    "email",
		},
		{
			nombre:   "password vacio",
			email:    "ana@example.com",
			password: "",
			campo:    "password",
			mensaje:  domain.ErrPasswordVacio.Error(),
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()

			servicio := new(mockServicioAuth)
			router := nuevoRouterAuth(t, servicio)

			w := ejecutarLogin(t, router, "application/json", loginJSON(c.email, c.password))

			require.Equal(t, http.StatusUnprocessableEntity, w.Code, "E-07 / CL-02")

			var errResp respuestaError
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &errResp))
			assert.Equal(t, "VALIDACION", errResp.Error)
			assert.Equal(t, c.campo, errResp.Campo)
			if c.mensaje != "" {
				assert.Equal(t, c.mensaje, errResp.Mensaje)
			}
			servicio.AssertNotCalled(t, "Login", mock.Anything, mock.Anything, mock.Anything,
				"la validacion de campos ocurre antes de cualquier comparacion")
		})
	}
}

// TestLoginE08BodyNoJsonResponde400 cubre E-08 y R-08: un body que no es
// JSON se rechaza con JSON_INVALIDO antes de validar campos o credenciales.
func TestLoginE08BodyNoJsonResponde400(t *testing.T) {
	t.Parallel()

	servicio := new(mockServicioAuth)
	router := nuevoRouterAuth(t, servicio)

	w := ejecutarLogin(t, router, "application/x-www-form-urlencoded",
		"email=ana@example.com&password=Secreto123")

	require.Equal(t, http.StatusBadRequest, w.Code, "E-08")

	var errResp respuestaError
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &errResp))
	assert.Equal(t, "JSON_INVALIDO", errResp.Error)
	servicio.AssertNotCalled(t, "Login", mock.Anything, mock.Anything, mock.Anything)
}

// TestLoginE09PrecedenciaDeErrores cubre E-09 y R-08: el orden de validacion
// es JSON (400) -> campos (422) -> credenciales (401) -> exito (200) y manda
// el primer problema detectado.
func TestLoginE09PrecedenciaDeErrores(t *testing.T) {
	casos := []struct {
		nombre         string
		contentType    string
		cuerpo         string
		statusEsperado int
		errorEsperado  string
	}{
		{
			nombre:         "body no JSON y campos vacios manda el 400",
			contentType:    "application/x-www-form-urlencoded",
			cuerpo:         "email=&password=",
			statusEsperado: http.StatusBadRequest,
			errorEsperado:  "JSON_INVALIDO",
		},
		{
			nombre:         "campos vacios y password incorrecta manda el 422",
			contentType:    "application/json",
			cuerpo:         `{"email":"","password":"OtraClave"}`,
			statusEsperado: http.StatusUnprocessableEntity,
			errorEsperado:  "VALIDACION",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()

			servicio := new(mockServicioAuth)
			servicio.On("Login", mock.Anything, mock.Anything, mock.Anything).
				Return("", nil, domain.ErrCredencialesInvalidas)
			router := nuevoRouterAuth(t, servicio)

			w := ejecutarLogin(t, router, c.contentType, c.cuerpo)

			require.Equal(t, c.statusEsperado, w.Code, "E-09 / R-08")

			var errResp respuestaError
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &errResp))
			assert.Equal(t, c.errorEsperado, errResp.Error)
			servicio.AssertNotCalled(t, "Login", mock.Anything, mock.Anything, mock.Anything,
				"el primer problema detectado gana y corta el resto de la validacion")
		})
	}
}

// TestE18NingunEndpointDevuelvePasswordHash cubre E-18 y CA-7: ni el login ni
// el listado de usuarios exponen la cadena password_hash ni el hash almacenado,
// aunque el objeto devuelto por el servicio lo traiga cargado.
func TestE18NingunEndpointDevuelvePasswordHash(t *testing.T) {
	t.Parallel()

	hash := usuarioAutenticado().PasswordHash

	auth := new(mockServicioAuth)
	auth.On("Login", mock.Anything, "ana@example.com", "Secreto123").
		Return("jwt-de-ana", usuarioAutenticado(), nil)
	routerAuth := nuevoRouterAuth(t, auth)

	usuarios := new(mockServicioUsuarios)
	usuarios.On("Listar", mock.Anything).
		Return([]domain.Usuario{*usuarioAutenticado()}, nil)
	gin.SetMode(gin.TestMode)
	routerUsuarios := gin.New()
	RegistrarRutasUsuarios(routerUsuarios, usuarios)

	wLogin := ejecutarLogin(t, routerAuth, "application/json", loginJSON("ana@example.com", "Secreto123"))
	wListado := ejecutarUsuarios(t, routerUsuarios, http.MethodGet, "/api/usuarios", nil)

	require.Equal(t, http.StatusOK, wLogin.Code)
	require.Equal(t, http.StatusOK, wListado.Code)

	for _, cuerpo := range []string{wLogin.Body.String(), wListado.Body.String()} {
		assert.NotContains(t, cuerpo, "password_hash", "E-18 / CA-7")
		assert.NotContains(t, cuerpo, hash, "E-18 / CA-7")
	}
}
