package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SprintBacklogHandler expone la asignación de historias al Sprint Backlog.
type SprintBacklogHandler struct {
	servicio ServicioSprintBacklog
}

// NuevoSprintBacklogHandler construye el controlador con inyección de dependencias.
func NuevoSprintBacklogHandler(servicio ServicioSprintBacklog) *SprintBacklogHandler {
	return &SprintBacklogHandler{servicio: servicio}
}

// RegistrarRutasSprintBacklog registra las rutas de US-04 bajo el grupo /api.
func RegistrarRutasSprintBacklog(r *gin.Engine, servicio ServicioSprintBacklog) {
	h := NuevoSprintBacklogHandler(servicio)
	api := r.Group("/api")
	proyectos := api.Group("/proyectos")
	{
		proyectos.GET("/:id/sprints/:sprintId/historias", h.Listar)
		proyectos.POST("/:id/sprints/:sprintId/historias", h.Asignar)
		proyectos.DELETE("/:id/sprints/:sprintId/historias/:historiaId", h.Desasignar)
	}
}

// Listar maneja GET /api/proyectos/:id/sprints/:sprintId/historias.
func (h *SprintBacklogHandler) Listar(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}
	sprintID, ok := parsearID(c, "sprintId")
	if !ok {
		return
	}

	detalle, err := h.servicio.ListarSprintBacklog(c.Request.Context(), proyectoID, sprintID)
	if err != nil {
		traducirError(c, err)
		return
	}

	c.JSON(http.StatusOK, construirRespuestaSprintBacklog(detalle))
}

// Asignar maneja POST /api/proyectos/:id/sprints/:sprintId/historias.
func (h *SprintBacklogHandler) Asignar(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}
	sprintID, ok := parsearID(c, "sprintId")
	if !ok {
		return
	}

	var dto asignarHistoriaSprintDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		responderJSONInvalido(c)
		return
	}

	if err := h.servicio.AsignarASprint(c.Request.Context(), proyectoID, sprintID, dto.HistoriaID); err != nil {
		traducirError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// Desasignar maneja DELETE /api/proyectos/:id/sprints/:sprintId/historias/:historiaId.
func (h *SprintBacklogHandler) Desasignar(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}
	sprintID, ok := parsearID(c, "sprintId")
	if !ok {
		return
	}
	historiaID, ok := parsearID(c, "historiaId")
	if !ok {
		return
	}

	if err := h.servicio.DesasignarDeSprint(c.Request.Context(), proyectoID, sprintID, historiaID); err != nil {
		traducirError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
