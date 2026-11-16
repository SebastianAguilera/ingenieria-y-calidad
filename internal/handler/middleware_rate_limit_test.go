package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func routerConRateLimit(limite gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(limite)
	r.POST("/api/auth/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.POST("/api/auth/registro", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.GET("/api/proyectos", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func pedirRateLimit(t *testing.T, router *gin.Engine, metodo, ruta, ip string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(metodo, ruta, nil)
	req.RemoteAddr = ip + ":12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestLimitarAuth(t *testing.T) {
	t.Run("permite hasta el maximo y luego responde 429", func(t *testing.T) {
		router := routerConRateLimit(LimitarAuth(2, time.Minute))

		primera := pedirRateLimit(t, router, http.MethodPost, "/api/auth/login", "192.0.2.10")
		segunda := pedirRateLimit(t, router, http.MethodPost, "/api/auth/login", "192.0.2.10")
		tercera := pedirRateLimit(t, router, http.MethodPost, "/api/auth/login", "192.0.2.10")

		assert.Equal(t, http.StatusOK, primera.Code)
		assert.Equal(t, http.StatusOK, segunda.Code)
		require.Equal(t, http.StatusTooManyRequests, tercera.Code)
		assert.Equal(t, "60", tercera.Header().Get("Retry-After"))

		var respuesta respuestaError
		require.NoError(t, json.Unmarshal(tercera.Body.Bytes(), &respuesta))
		assert.Equal(t, codigoDemasiadasSolicitudes, respuesta.Error)
	})

	t.Run("cada IP tiene su propio limite", func(t *testing.T) {
		router := routerConRateLimit(LimitarAuth(1, time.Minute))

		primera := pedirRateLimit(t, router, http.MethodPost, "/api/auth/login", "192.0.2.11")
		segunda := pedirRateLimit(t, router, http.MethodPost, "/api/auth/login", "192.0.2.12")

		assert.Equal(t, http.StatusOK, primera.Code)
		assert.Equal(t, http.StatusOK, segunda.Code)
	})

	t.Run("cada endpoint tiene su propio limite", func(t *testing.T) {
		router := routerConRateLimit(LimitarAuth(1, time.Minute))

		login := pedirRateLimit(t, router, http.MethodPost, "/api/auth/login", "192.0.2.13")
		registro := pedirRateLimit(t, router, http.MethodPost, "/api/auth/registro", "192.0.2.13")

		assert.Equal(t, http.StatusOK, login.Code)
		assert.Equal(t, http.StatusOK, registro.Code)
	})

	t.Run("las rutas fuera de auth no se limitan", func(t *testing.T) {
		router := routerConRateLimit(LimitarAuth(1, time.Minute))

		for i := 0; i < 3; i++ {
			w := pedirRateLimit(t, router, http.MethodGet, "/api/proyectos", "192.0.2.14")
			assert.Equal(t, http.StatusOK, w.Code)
		}
	})

	t.Run("al vencer la ventana vuelve a permitir", func(t *testing.T) {
		ahora := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
		router := routerConRateLimit(limitarAuthConReloj(1, time.Minute, func() time.Time {
			return ahora
		}))

		primera := pedirRateLimit(t, router, http.MethodPost, "/api/auth/login", "192.0.2.15")
		bloqueada := pedirRateLimit(t, router, http.MethodPost, "/api/auth/login", "192.0.2.15")

		ahora = ahora.Add(time.Minute)

		tercera := pedirRateLimit(t, router, http.MethodPost, "/api/auth/login", "192.0.2.15")

		assert.Equal(t, http.StatusOK, primera.Code)
		assert.Equal(t, http.StatusTooManyRequests, bloqueada.Code)
		assert.Equal(t, http.StatusOK, tercera.Code)
	})
}
