package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const origenDeProduccion = "https://ingenieria-y-calidad.vercel.app"

func routerConCORS(appEnv string, origenes []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS(appEnv, origenes))
	r.POST("/api/auth/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func pedirCORS(t *testing.T, router *gin.Engine, metodo string, cabeceras map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(metodo, "/api/auth/login", nil)
	for clave, valor := range cabeceras {
		req.Header.Set(clave, valor)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// TestCORSEnDesarrollo verifica que fuera de produccion cualquier origen puede
// leer la respuesta, que es lo que necesita el entorno local.
func TestCORSEnDesarrollo(t *testing.T) {
	t.Run("cualquier origen se refleja", func(t *testing.T) {
		router := routerConCORS("development", nil)

		w := pedirCORS(t, router, http.MethodPost, map[string]string{
			"Origin": "https://cualquier-origen.example",
		})

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "https://cualquier-origen.example", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "Origin", w.Header().Get("Vary"))
	})

	t.Run("sin Origin no bloquea y usa comodin", func(t *testing.T) {
		router := routerConCORS("development", nil)

		w := pedirCORS(t, router, http.MethodPost, nil)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})
}

// TestCORSEnProduccion verifica la allowlist exacta del frontend y que la
// ausencia de Origin no bloquee la solicitud: CORS solo limita lo que el
// navegador puede leer.
func TestCORSEnProduccion(t *testing.T) {
	origenes := []string{origenDeProduccion}

	t.Run("origen permitido recibe la cabecera exacta", func(t *testing.T) {
		router := routerConCORS("production", origenes)

		w := pedirCORS(t, router, http.MethodPost, map[string]string{"Origin": origenDeProduccion})

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, origenDeProduccion, w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "Origin", w.Header().Get("Vary"))
	})

	t.Run("origen no permitido devuelve 403", func(t *testing.T) {
		router := routerConCORS("production", origenes)

		w := pedirCORS(t, router, http.MethodPost, map[string]string{"Origin": "https://intruso.example"})

		require.Equal(t, http.StatusForbidden, w.Code)
		var respuesta respuestaError
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta))
		assert.Equal(t, codigoOrigenNoPermitido, respuesta.Error)
	})

	t.Run("sin Origin procesa la solicitud", func(t *testing.T) {
		router := routerConCORS("production", origenes)

		w := pedirCORS(t, router, http.MethodPost, nil)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestCORSPreflight(t *testing.T) {
	origenes := []string{origenDeProduccion}

	t.Run("origen permitido devuelve 204 con metodos y cabeceras", func(t *testing.T) {
		router := routerConCORS("production", origenes)

		w := pedirCORS(t, router, http.MethodOptions, map[string]string{
			"Origin":                        origenDeProduccion,
			"Access-Control-Request-Method": "POST",
		})

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, origenDeProduccion, w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
		assert.Equal(t, "600", w.Header().Get("Access-Control-Max-Age"))
	})

	t.Run("origen no permitido devuelve 403", func(t *testing.T) {
		router := routerConCORS("production", origenes)

		w := pedirCORS(t, router, http.MethodOptions, map[string]string{
			"Origin": "https://intruso.example",
		})

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("sin Origin devuelve 204 sin bloquear", func(t *testing.T) {
		router := routerConCORS("production", origenes)

		w := pedirCORS(t, router, http.MethodOptions, nil)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}
