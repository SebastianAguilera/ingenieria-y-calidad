package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegistrarRutasMetricas monta los endpoints de solo lectura de US-20 bajo
// /api/proyectos. Las rutas quedan detrás del guard de membresía registrado en
// router.go.
func RegistrarRutasMetricas(router *gin.Engine, servicio ServicioMetricas) {
	proyectos := router.Group("/api/proyectos")
	{
		proyectos.GET("/:id/sprints/:sprintId/metricas", metricasDeSprintHandler(servicio))
		proyectos.GET("/:id/metricas", metricasDeProyectoHandler(servicio))
	}
}

func metricasDeSprintHandler(servicio ServicioMetricas) gin.HandlerFunc {
	return func(c *gin.Context) {
		proyectoID, ok := parsearID(c, "id")
		if !ok {
			return
		}
		sprintID, ok := parsearID(c, "sprintId")
		if !ok {
			return
		}

		metricas, err := servicio.CalcularPorSprint(c.Request.Context(), proyectoID, sprintID)
		if err != nil {
			traducirError(c, err)
			return
		}

		c.JSON(http.StatusOK, construirRespuestaMetricasSprint(sprintID, metricas))
	}
}

func metricasDeProyectoHandler(servicio ServicioMetricas) gin.HandlerFunc {
	return func(c *gin.Context) {
		proyectoID, ok := parsearID(c, "id")
		if !ok {
			return
		}

		metricas, err := servicio.CalcularPorProyecto(c.Request.Context(), proyectoID)
		if err != nil {
			traducirError(c, err)
			return
		}

		c.JSON(http.StatusOK, construirRespuestaMetricasProyecto(proyectoID, metricas))
	}
}
