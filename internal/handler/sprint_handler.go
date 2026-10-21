package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ingenieria-y-calidad/internal/domain"
)

// ServicioSprints es el puerto que el handler necesita del service de Sprints.
// Se define en el consumidor para que los tests inyecten un mock.
type ServicioSprints interface {
	Crear(ctx context.Context, proyectoID uint, nuevo domain.NuevoSprint) (*domain.Sprint, error)
	Listar(ctx context.Context, proyectoID uint) ([]domain.Sprint, error)
	CambiarEstado(ctx context.Context, proyectoID, sprintID uint, destino domain.EstadoSprint) (*domain.Sprint, error)
}

// SprintHandler expone el ciclo de vida de los Sprints de un proyecto por HTTP.
type SprintHandler struct {
	servicio ServicioSprints
}

// NuevoSprintHandler construye el controlador con inyeccion de dependencias.
func NuevoSprintHandler(servicio ServicioSprints) *SprintHandler {
	return &SprintHandler{servicio: servicio}
}

// RegistrarRutasSprints registra las rutas de US-03 bajo el grupo /api.
func RegistrarRutasSprints(r *gin.Engine, servicio ServicioSprints) {
	h := NuevoSprintHandler(servicio)
	api := r.Group("/api")
	proyectos := api.Group("/proyectos")
	{
		proyectos.POST("/:id/sprints", h.Crear)
		proyectos.GET("/:id/sprints", h.Listar)
		proyectos.PATCH("/:id/sprints/:sprintId/estado", h.CambiarEstado)
	}
}

// Crear maneja POST /api/proyectos/:id/sprints.
func (h *SprintHandler) Crear(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}

	var dto altaSprintDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		responderJSONInvalido(c)
		return
	}

	fechaInicio, ok := parsearFechaSprint(c, dto.FechaInicio, "fecha_inicio")
	if !ok {
		return
	}

	fechaFin, ok := parsearFechaSprint(c, dto.FechaFin, "fecha_fin")
	if !ok {
		return
	}

	sprint, err := h.servicio.Crear(c.Request.Context(), proyectoID, domain.NuevoSprint{
		Identificador: dto.Identificador,
		SprintGoal:    dto.SprintGoal,
		FechaInicio:   fechaInicio,
		FechaFin:      fechaFin,
	})
	if err != nil {
		traducirError(c, err)
		return
	}

	c.JSON(http.StatusCreated, construirRespuestaSprint(sprint))
}

// Listar maneja GET /api/proyectos/:id/sprints.
func (h *SprintHandler) Listar(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}

	sprints, err := h.servicio.Listar(c.Request.Context(), proyectoID)
	if err != nil {
		traducirError(c, err)
		return
	}

	respuestas := make([]respuestaSprint, 0, len(sprints))
	for _, sprint := range sprints {
		respuestas = append(respuestas, construirRespuestaSprint(&sprint))
	}

	c.JSON(http.StatusOK, respuestaListadoSprints{
		Total:   len(respuestas),
		Sprints: respuestas,
	})
}

// CambiarEstado maneja PATCH /api/proyectos/:id/sprints/:sprintId/estado.
func (h *SprintHandler) CambiarEstado(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}

	sprintID, ok := parsearID(c, "sprintId")
	if !ok {
		return
	}

	var dto cambioEstadoSprintDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		responderJSONInvalido(c)
		return
	}

	sprint, err := h.servicio.CambiarEstado(c.Request.Context(), proyectoID, sprintID, dto.Estado)
	if err != nil {
		traducirError(c, err)
		return
	}

	c.JSON(http.StatusOK, construirRespuestaSprint(sprint))
}

// parsearFechaSprint exige una fecha calendario "YYYY-MM-DD" no vacia. El
// handler la convierte a time.Time porque el formato es parte del contrato de
// la API (mismo criterio que US-01).
func parsearFechaSprint(c *gin.Context, valor, campo string) (time.Time, bool) {
	fecha, err := time.Parse(formatoFecha, strings.TrimSpace(valor))
	if err != nil {
		responderFechaInvalida(c, campo)
		return time.Time{}, false
	}
	return fecha, true
}
