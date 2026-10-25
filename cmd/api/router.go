package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ingenieria-y-calidad/internal/handler"
)

// setupRouter construye el router de la aplicacion. Vive fuera de main para que
// los tests ejerciten exactamente las mismas rutas que se sirven en produccion.
func setupRouter(servicioProyectos handler.ServicioProyectos, servicioHistorias handler.ServicioHistorias, servicioSprints handler.ServicioSprints, servicioUsuarios handler.ServicioUsuarios, servicioSprintBacklog handler.ServicioSprintBacklog) *gin.Engine {
	r := gin.Default()
	r.GET("/health", healthHandler)
	handler.RegistrarRutasProyectos(r, servicioProyectos)
	handler.RegistrarRutasHistorias(r, servicioHistorias)
	handler.RegistrarRutasSprints(r, servicioSprints)
	if servicioUsuarios != nil {
		handler.RegistrarRutasUsuarios(r, servicioUsuarios)
	}
	if servicioSprintBacklog != nil {
		handler.RegistrarRutasSprintBacklog(r, servicioSprintBacklog)
	}
	return r
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Software Metrics & Estimation Engine API running",
	})
}
