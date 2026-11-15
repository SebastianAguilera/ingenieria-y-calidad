package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegistrarRutasReporte(router *gin.Engine, servicio ServicioReporte) {
	proyectos := router.Group("/api/proyectos")
	{
		proyectos.GET("/:id/reporte", reporteProyectoHandler(servicio))
		proyectos.GET("/:id/sprints/:sprintId/reporte", reporteSprintHandler(servicio))
	}
}

func reporteProyectoHandler(servicio ServicioReporte) gin.HandlerFunc {
	return func(c *gin.Context) {
		proyectoID, ok := parsearID(c, "id")
		if !ok {
			return
		}

		reporte, err := servicio.ObtenerReporteProyecto(c.Request.Context(), proyectoID)
		if err != nil {
			traducirError(c, err)
			return
		}

		c.JSON(http.StatusOK, construirRespuestaReporteProyecto(reporte))
	}
}

func reporteSprintHandler(servicio ServicioReporte) gin.HandlerFunc {
	return func(c *gin.Context) {
		proyectoID, ok := parsearID(c, "id")
		if !ok {
			return
		}
		sprintID, ok := parsearID(c, "sprintId")
		if !ok {
			return
		}

		reporte, err := servicio.ObtenerReporteSprint(c.Request.Context(), proyectoID, sprintID)
		if err != nil {
			traducirError(c, err)
			return
		}

		c.JSON(http.StatusOK, construirRespuestaReporteSprint(reporte))
	}
}
