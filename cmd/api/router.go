package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ingenieria-y-calidad/internal/handler"
)

// setupRouter construye el router de la aplicacion. Vive fuera de main para que
// los tests ejerciten exactamente las mismas rutas que se sirven en produccion.
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
