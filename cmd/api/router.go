package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ingenieria-y-calidad/internal/handler"
)

// setupRouter construye el router de la aplicacion. Vive fuera de main para que
// los tests ejerciten exactamente las mismas rutas que se sirven en produccion.
// R-06: todo lo que no figura en la lista publica queda tras el middleware de
// autenticacion (default-deny).
func setupRouter(servicioProyectos handler.ServicioProyectos, servicioHistorias handler.ServicioHistorias, servicioSprints handler.ServicioSprints, servicioUsuarios handler.ServicioUsuarios, servicioSprintBacklog handler.ServicioSprintBacklog, servicioDefectos handler.ServicioDefectos, servicioCriterios handler.ServicioAcceptanceCriteria, servicioAuth handler.ServicioAuth, jwtSecret string) *gin.Engine {
	r := gin.Default()
	r.GET("/health", healthHandler)
	if servicioAuth != nil {
		handler.RegistrarRutasAuth(r, servicioAuth)
	}
	r.Use(handler.Autenticar(jwtSecret, "/health", "/api/auth/login", "/api/auth/registro"))
	// US-F10: el guard de membresia se registra justo despues del middleware de
	// autenticacion y antes de cualquier ruta de proyectos: en Gin las rutas
	// registradas antes del Use no reciben la cadena de middlewares.
	r.Use(handler.ExigirAccesoAProyecto(servicioProyectos))
	handler.RegistrarRutasProyectos(r, servicioProyectos)
	handler.RegistrarRutasHistorias(r, servicioHistorias)
	handler.RegistrarRutasSprints(r, servicioSprints)
	if servicioUsuarios != nil {
		handler.RegistrarRutasUsuarios(r, servicioUsuarios)
	}
	if servicioSprintBacklog != nil {
		handler.RegistrarRutasSprintBacklog(r, servicioSprintBacklog)
	}
	if servicioDefectos != nil {
		handler.RegistrarRutasDefectos(r, servicioDefectos)
	}
	if servicioCriterios != nil {
		handler.RegistrarRutasAcceptanceCriteria(r, servicioCriterios)
	}
	return r
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Software Metrics & Estimation Engine API running",
	})
}
