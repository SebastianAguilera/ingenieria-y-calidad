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

// nuevoRouterConFrontendKey reproduce el orden de setupRouter: las rutas
// publicas se registran antes del middleware y las protegidas despues. Asi se
// verifica que /health siga accesible sin la clave (health check de Render).
func nuevoRouterConFrontendKey(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.Use(ExigirFrontendKey())
	r.GET("/api/protegida", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func pedirConCabeceras(t *testing.T, router *gin.Engine, metodo, ruta string, cabeceras map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(metodo, ruta, nil)
	for clave, valor := range cabeceras {
		req.Header.Set(clave, valor)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// TestExigirFrontendKey cubre los escenarios de la clave compartida entre el
// proxy de Vercel y la API: la validacion solo se activa cuando FRONTEND_KEY
// esta definida (D5), y /health nunca la exige.
func TestExigirFrontendKey(t *testing.T) {
	t.Run("sin FRONTEND_KEY configurada no bloquea", func(t *testing.T) {
		t.Setenv("FRONTEND_KEY", "")
		router := nuevoRouterConFrontendKey(t)

		w := pedirConCabeceras(t, router, http.MethodGet, "/api/protegida", nil)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("sin cabecera devuelve 401", func(t *testing.T) {
		t.Setenv("FRONTEND_KEY", "clave-de-prueba")
		router := nuevoRouterConFrontendKey(t)

		w := pedirConCabeceras(t, router, http.MethodGet, "/api/protegida", nil)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		var respuesta respuestaError
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta))
		assert.Equal(t, codigoFrontendKeyInvalida, respuesta.Error)
	})

	t.Run("clave incorrecta devuelve 401", func(t *testing.T) {
		t.Setenv("FRONTEND_KEY", "clave-de-prueba")
		router := nuevoRouterConFrontendKey(t)

		w := pedirConCabeceras(t, router, http.MethodGet, "/api/protegida", map[string]string{
			cabeceraFrontendKey: "otra-clave",
		})

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("clave correcta deja pasar", func(t *testing.T) {
		t.Setenv("FRONTEND_KEY", "clave-de-prueba")
		router := nuevoRouterConFrontendKey(t)

		w := pedirConCabeceras(t, router, http.MethodGet, "/api/protegida", map[string]string{
			cabeceraFrontendKey: "clave-de-prueba",
		})

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("health queda publico aunque FRONTEND_KEY este configurada", func(t *testing.T) {
		t.Setenv("FRONTEND_KEY", "clave-de-prueba")
		router := nuevoRouterConFrontendKey(t)

		w := pedirConCabeceras(t, router, http.MethodGet, "/health", nil)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
