package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/config"
)

const origenFrontendDePrueba = "https://ingenieria-y-calidad.vercel.app"

func routerConConfiguracion(t *testing.T, cfg config.Config) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	return setupRouterConConfig(
		servicioProyectosVacio{},
		servicioHistoriasQueFalla{},
		servicioSprintsQueFalla{},
		nil, nil, nil, nil,
		nil, nil, nil, nil,
		servicioAuthPrueba{},
		cfg,
	)
}

func configuracionDePrueba() config.Config {
	return config.Config{
		AppEnv:              "test",
		JWTSecret:           secretoDePruebaRouter,
		AuthRateLimitMax:    20,
		AuthRateLimitWindow: time.Minute,
	}
}

func loginCORS(t *testing.T, router *gin.Engine, origin, ip string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		bytes.NewReader([]byte(`{"email":"ana@example.com","password":"Secreto123"}`)))
	req.Header.Set("Content-Type", "application/json")
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	req.RemoteAddr = ip + ":12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestRouterCORSYRateLimit(t *testing.T) {
	t.Run("preflight permitido devuelve 204 con CORS", func(t *testing.T) {
		cfg := configuracionDePrueba()
		cfg.AppEnv = "production"
		cfg.CORSAllowedOrigins = []string{origenFrontendDePrueba}
		router := routerConConfiguracion(t, cfg)

		req := httptest.NewRequest(http.MethodOptions, "/api/auth/login", nil)
		req.Header.Set("Origin", origenFrontendDePrueba)
		req.Header.Set("Access-Control-Request-Method", "POST")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, origenFrontendDePrueba, w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("preflight con origen no permitido devuelve 403", func(t *testing.T) {
		cfg := configuracionDePrueba()
		cfg.AppEnv = "production"
		cfg.CORSAllowedOrigins = []string{origenFrontendDePrueba}
		router := routerConConfiguracion(t, cfg)

		req := httptest.NewRequest(http.MethodOptions, "/api/auth/login", nil)
		req.Header.Set("Origin", "https://intruso.example")
		req.Header.Set("Access-Control-Request-Method", "POST")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusForbidden, w.Code)
		var respuesta struct {
			Error string `json:"error"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta))
		assert.Equal(t, "ORIGEN_NO_PERMITIDO", respuesta.Error)
	})

	t.Run("login sin Origin se procesa en produccion", func(t *testing.T) {
		cfg := configuracionDePrueba()
		cfg.AppEnv = "production"
		cfg.CORSAllowedOrigins = []string{origenFrontendDePrueba}
		router := routerConConfiguracion(t, cfg)

		w := loginCORS(t, router, "", "192.0.2.20")

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("login con origen permitido recibe CORS", func(t *testing.T) {
		cfg := configuracionDePrueba()
		cfg.AppEnv = "production"
		cfg.CORSAllowedOrigins = []string{origenFrontendDePrueba}
		router := routerConConfiguracion(t, cfg)

		w := loginCORS(t, router, origenFrontendDePrueba, "192.0.2.21")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, origenFrontendDePrueba, w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("mas de 20 intentos desde la misma IP devuelve 429", func(t *testing.T) {
		cfg := configuracionDePrueba()
		router := routerConConfiguracion(t, cfg)

		for i := 0; i < 20; i++ {
			w := loginCORS(t, router, "", "192.0.2.22")
			require.Equal(t, http.StatusOK, w.Code, "intento %d", i+1)
		}

		w := loginCORS(t, router, "", "192.0.2.22")

		require.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.NotEmpty(t, w.Header().Get("Retry-After"))
	})
}
