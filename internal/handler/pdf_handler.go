package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegistrarRutasPDF(router *gin.Engine, servicio ServicioPDF) {
	proyectos := router.Group("/api/proyectos")
	{
		proyectos.GET("/:id/reporte.pdf", pdfProyectoHandler(servicio))
		proyectos.GET("/:id/sprints/:sprintId/reporte.pdf", pdfSprintHandler(servicio))
	}
}

func pdfProyectoHandler(servicio ServicioPDF) gin.HandlerFunc {
	return func(c *gin.Context) {
		proyectoID, ok := parsearID(c, "id")
		if !ok {
			return
		}

		pdfBytes, filename, err := servicio.GenerarPDFProyecto(c.Request.Context(), proyectoID)
		if err != nil {
			traducirError(c, err)
			return
		}

		servirPDF(c, pdfBytes, filename)
	}
}

func pdfSprintHandler(servicio ServicioPDF) gin.HandlerFunc {
	return func(c *gin.Context) {
		proyectoID, ok := parsearID(c, "id")
		if !ok {
			return
		}
		sprintID, ok := parsearID(c, "sprintId")
		if !ok {
			return
		}

		pdfBytes, filename, err := servicio.GenerarPDFSprint(c.Request.Context(), proyectoID, sprintID)
		if err != nil {
			traducirError(c, err)
			return
		}

		servirPDF(c, pdfBytes, filename)
	}
}

func servirPDF(c *gin.Context, pdfBytes []byte, filename string) {
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
