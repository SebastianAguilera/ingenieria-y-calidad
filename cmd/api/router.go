package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ingenieria-y-calidad/internal/handler"
)

func setupRouter(servicioProyectos handler.ServicioProyectos) *gin.Engine {
	r := gin.Default()
	r.GET("/health", healthHandler)
	handler.RegistrarRutasProyectos(r, servicioProyectos)
	return r
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Software Metrics & Estimation Engine API running",
	})
}
