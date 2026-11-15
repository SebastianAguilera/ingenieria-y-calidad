package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"ingenieria-y-calidad/internal/domain"
)

// ServicioDefectos es el contrato que el handler necesita del service de
// defectos. Lo implementa *service.DefectoService.
type ServicioDefectos interface {
	Crear(ctx context.Context, proyectoID uint, nuevo domain.NuevoDefecto) (*domain.Defecto, error)
	Actualizar(ctx context.Context, proyectoID, defectoID uint, actualizacion domain.ActualizacionDefecto) (*domain.Defecto, error)
	Listar(ctx context.Context, proyectoID uint) ([]domain.Defecto, error)
	ListarPorSprint(ctx context.Context, proyectoID, sprintID uint, tipo domain.TipoListadoDefecto) ([]domain.Defecto, error)
}

// RegistrarRutasDefectos monta las rutas de US-09 y el listado por Sprint de
// US-18 bajo /api/proyectos.
func RegistrarRutasDefectos(router *gin.Engine, servicio ServicioDefectos) {
	proyectos := router.Group("/api/proyectos")
	{
		proyectos.POST("/:id/defectos", crearDefectoHandler(servicio))
		proyectos.GET("/:id/defectos", listarDefectosHandler(servicio))
		proyectos.PATCH("/:id/defectos/:defectoId", actualizarDefectoHandler(servicio))
		proyectos.GET("/:id/sprints/:sprintId/defectos", listarDefectosPorSprintHandler(servicio))
	}
}

func listarDefectosPorSprintHandler(servicio ServicioDefectos) gin.HandlerFunc {
	return func(c *gin.Context) {
		proyectoID, ok := parsearID(c, "id")
		if !ok {
			return
		}
		sprintID, ok := parsearID(c, "sprintId")
		if !ok {
			return
		}

		tipo := domain.TipoListadoDefecto(c.Query("tipo"))
		defectos, err := servicio.ListarPorSprint(c.Request.Context(), proyectoID, sprintID, tipo)
		if err != nil {
			traducirError(c, err)
			return
		}

		respuesta := respuestaListadoDefectos{
			Total:    len(defectos),
			Defectos: make([]respuestaDefecto, len(defectos)),
		}
		for idx, defecto := range defectos {
			respuesta.Defectos[idx] = construirRespuestaDefecto(&defecto)
		}

		c.JSON(http.StatusOK, respuesta)
	}
}

func crearDefectoHandler(servicio ServicioDefectos) gin.HandlerFunc {
	return func(c *gin.Context) {
		proyectoID, ok := parsearID(c, "id")
		if !ok {
			return
		}

		var cuerpo altaDefectoDTO
		if err := c.ShouldBindJSON(&cuerpo); err != nil {
			responderJSONInvalido(c)
			return
		}

		defecto, err := servicio.Crear(c.Request.Context(), proyectoID, domain.NuevoDefecto{
			Descripcion:       cuerpo.Descripcion,
			Severidad:         cuerpo.Severidad,
			HistoriaID:        cuerpo.HistoriaID,
			SprintDeteccionID: cuerpo.SprintDeteccionID,
			ReportadoPor:      cuerpo.ReportadoPor,
		})
		if err != nil {
			traducirError(c, err)
			return
		}

		c.JSON(http.StatusCreated, construirRespuestaDefecto(defecto))
	}
}

func listarDefectosHandler(servicio ServicioDefectos) gin.HandlerFunc {
	return func(c *gin.Context) {
		proyectoID, ok := parsearID(c, "id")
		if !ok {
			return
		}

		defectos, err := servicio.Listar(c.Request.Context(), proyectoID)
		if err != nil {
			traducirError(c, err)
			return
		}

		respuesta := respuestaListadoDefectos{
			Total:    len(defectos),
			Defectos: make([]respuestaDefecto, len(defectos)),
		}
		for idx, defecto := range defectos {
			respuesta.Defectos[idx] = construirRespuestaDefecto(&defecto)
		}

		c.JSON(http.StatusOK, respuesta)
	}
}

func actualizarDefectoHandler(servicio ServicioDefectos) gin.HandlerFunc {
	return func(c *gin.Context) {
		proyectoID, ok := parsearID(c, "id")
		if !ok {
			return
		}
		defectoID, ok := parsearID(c, "defectoId")
		if !ok {
			return
		}

		var cuerpo actualizacionDefectoDTO
		if err := c.ShouldBindJSON(&cuerpo); err != nil {
			responderJSONInvalido(c)
			return
		}

		defecto, err := servicio.Actualizar(c.Request.Context(), proyectoID, defectoID, domain.ActualizacionDefecto{
			Estado:                  cuerpo.Estado,
			SprintResolucionID:      cuerpo.SprintResolucionID,
			Descripcion:             cuerpo.Descripcion,
			Severidad:               cuerpo.Severidad,
			SprintDeteccionPresente: len(cuerpo.SprintDeteccionID) > 0,
		})
		if err != nil {
			traducirError(c, err)
			return
		}

		c.JSON(http.StatusOK, construirRespuestaDefecto(defecto))
	}
}
