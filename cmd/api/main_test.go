package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheckHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter(nil)

	w := httptest.NewRecorder()
	req, httptestReq := http.NewRequest(http.MethodGet, "/health", nil)
	assert.NoError(t, httptestReq)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func TestRutasDeProyectosRegistradas(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter(nil)

	rutas := []struct {
		metodo string
		ruta   string
	}{
		{http.MethodPost, "/api/proyectos"},
		{http.MethodGet, "/api/proyectos"},
		{http.MethodGet, "/api/proyectos/1"},
		{http.MethodPut, "/api/proyectos/1"},
		{http.MethodPatch, "/api/proyectos/1/estado"},
		{http.MethodDelete, "/api/proyectos/1"},
		{http.MethodPost, "/api/proyectos/1/integrantes"},
		{http.MethodDelete, "/api/proyectos/1/integrantes/1"},
	}

	for _, r := range rutas {
		t.Run(r.metodo+" "+r.ruta, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, err := http.NewRequest(r.metodo, r.ruta, nil)
			assert.NoError(t, err)
			router.ServeHTTP(w, req)

			assert.NotEqual(t, http.StatusNotFound, w.Code, "la ruta debe estar registrada: %s %s", r.metodo, r.ruta)
			assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code)
		})
	}
}
