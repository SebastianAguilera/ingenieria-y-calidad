package handler

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

// secretoDePruebaMiddleware es el secreto con el que el middleware de prueba
// valida los tokens: equivale a JWT_SECRET en produccion.
const secretoDePruebaMiddleware = "secreto-de-prueba-us10"

func claimsVigentes() jwt.MapClaims {
	ahora := time.Now()
	return jwt.MapClaims{
		"sub":   "1",
		"email": "ana@example.com",
		"iat":   ahora.Unix(),
		"exp":   ahora.Add(24 * time.Hour).Unix(),
	}
}

// tokenFirmadoCon emite un JWT HS256 real con el secreto indicado.
func tokenFirmadoCon(t *testing.T, claims jwt.MapClaims, secreto string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	firmado, err := token.SignedString([]byte(secreto))
	require.NoError(t, err)
	return firmado
}

// tokenConAlgoritmo arma un JWT a mano con el algoritmo indicado y una firma
// invalida: el middleware debe rechazarlo por el algoritmo antes que por la
// firma (R-06, CL-04).
func tokenConAlgoritmo(alg string) string {
	cabecera := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"` + alg + `","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"1","email":"ana@example.com"}`))
	return cabecera + "." + payload + ".firma-invalida"
}

// nuevoRouterProtegido devuelve un router con el middleware aplicado sobre una
// ruta protegida y un contador de ejecuciones que permite verificar que la
// logica del negocio no se alcanza cuando el token no sirve (E-11).
func nuevoRouterProtegido(t *testing.T, secreto string) (*gin.Engine, *int) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	ejecuciones := 0
	r := gin.New()
	r.Use(Autenticar(secreto))
	r.GET("/api/proyectos", func(c *gin.Context) {
		ejecuciones++
		c.JSON(http.StatusOK, []string{})
	})
	return r, &ejecuciones
}

// pedirProtegido llama a GET /api/proyectos con el header Authorization
// indicado; un valor vacio significa no enviar el header.
func pedirProtegido(t *testing.T, router *gin.Engine, authorization string) *httptest.ResponseRecorder {
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

// TestE10TokenValidoPasaElMiddleware cubre E-10, CA-6 y R-06: con el Bearer
// emitido por el login el middleware deja pasar la solicitud y la ruta responde
// con su resultado normal.
func TestE10TokenValidoPasaElMiddleware(t *testing.T) {
	t.Parallel()

	router, ejecuciones := nuevoRouterProtegido(t, secretoDePruebaMiddleware)
	token := tokenFirmadoCon(t, claimsVigentes(), secretoDePruebaMiddleware)

	w := pedirProtegido(t, router, "Bearer "+token)

	require.Equal(t, http.StatusOK, w.Code, "E-10 / CA-6")
	assert.Equal(t, 1, *ejecuciones, "la ruta protegida debe ejecutarse")
}

// TestE11SinHeaderAuthorizationResponde401 cubre E-11, CA-4, R-06 y CL-05:
// default-deny sobre /api/*: sin header la solicitud no llega a la logica del
// listado.
func TestE11SinHeaderAuthorizationResponde401(t *testing.T) {
	t.Parallel()

	router, ejecuciones := nuevoRouterProtegido(t, secretoDePruebaMiddleware)

	w := pedirProtegido(t, router, "")

	require.Equal(t, http.StatusUnauthorized, w.Code, "E-11 / CA-4")

	var errResp respuestaError
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &errResp))
	assert.Equal(t, "TOKEN_REQUERIDO", errResp.Error)
	assert.Equal(t, domain.ErrTokenRequerido.Error(), errResp.Mensaje)
	assert.Equal(t, 0, *ejecuciones, "la logica de listado no se ejecuta")
}

// TestE12AuthorizationSinPrefijoBearerResponde401 cubre E-12, R-06 y CL-05:
// sin el prefijo "Bearer " el resultado es identico al header ausente y el
// token ni siquiera se parsea.
func TestE12AuthorizationSinPrefijoBearerResponde401(t *testing.T) {
	t.Parallel()

	router, ejecuciones := nuevoRouterProtegido(t, secretoDePruebaMiddleware)
	token := tokenFirmadoCon(t, claimsVigentes(), secretoDePruebaMiddleware)

	w := pedirProtegido(t, router, token)

	require.Equal(t, http.StatusUnauthorized, w.Code, "E-12 / CL-05")

	var errResp respuestaError
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &errResp))
	assert.Equal(t, "TOKEN_REQUERIDO", errResp.Error)
	assert.Equal(t, 0, *ejecuciones)
}

// TestE13TokenRechazadoResponde401TokenInvalido cubre E-13, CA-5, R-06 y
// CL-04/CL-05: con el prefijo "Bearer " presente, cualquier fallo de parseo,
// firma, vigencia o algoritmo responde TOKEN_INVALIDO sin filtrar el detalle
// interno.
func TestE13TokenRechazadoResponde401TokenInvalido(t *testing.T) {
	casos := []struct {
		nombre string
		token  string
	}{
		{
			nombre: "firmado con otra clave",
			token:  tokenFirmadoCon(t, claimsVigentes(), "otra-clave-distinta"),
		},
		{
			nombre: "expirado",
			token: tokenFirmadoCon(t, jwt.MapClaims{
				"sub":   "1",
				"email": "ana@example.com",
				"iat":   time.Now().Add(-2 * time.Hour).Unix(),
				"exp":   time.Now().Add(-time.Hour).Unix(),
			}, secretoDePruebaMiddleware),
		},
		{
			nombre: "con alg none",
			token:  tokenConAlgoritmo("none"),
		},
		{
			nombre: "con alg distinto de HS256",
			token:  tokenConAlgoritmo("RS256"),
		},
		{
			nombre: "que no es un JWT (Bearer basura)",
			token:  "basura",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()

			router, ejecuciones := nuevoRouterProtegido(t, secretoDePruebaMiddleware)

			w := pedirProtegido(t, router, "Bearer "+c.token)

			require.Equal(t, http.StatusUnauthorized, w.Code, "E-13 / CA-5")

			var errResp respuestaError
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &errResp))
			assert.Equal(t, "TOKEN_INVALIDO", errResp.Error)
			assert.Equal(t, domain.ErrTokenInvalido.Error(), errResp.Mensaje,
				"la respuesta solo expone el mensaje generico")
			assert.NotContains(t, errResp.Mensaje, "signature")
			assert.NotContains(t, errResp.Mensaje, "expired")
			assert.NotContains(t, errResp.Mensaje, "malformed")
			assert.Equal(t, 0, *ejecuciones, "el token invalido no alcanza la ruta")
		})
	}
}
