package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"ingenieria-y-calidad/internal/domain"
)

// ServicioHistorias es el puerto que el handler necesita del service de
// historias. Se define en el consumidor para que los tests inyecten un mock.
type ServicioHistorias interface {
	Crear(ctx context.Context, proyectoID uint, nuevo domain.NuevaHistoria) (*domain.Historia, error)
	Listar(ctx context.Context, proyectoID uint) ([]domain.Historia, error)
	Actualizar(ctx context.Context, proyectoID, historiaID uint, actualizacion domain.ActualizacionHistoria) (*domain.Historia, error)
}

// HistoriaHandler expone el Product Backlog de un proyecto por HTTP.
type HistoriaHandler struct {
	servicio ServicioHistorias
}

// NuevoHistoriaHandler construye el controlador con inyeccion de dependencias.
func NuevoHistoriaHandler(servicio ServicioHistorias) *HistoriaHandler {
	return &HistoriaHandler{servicio: servicio}
}

// RegistrarRutasHistorias registra las rutas de US-02 bajo el grupo /api.
func RegistrarRutasHistorias(r *gin.Engine, servicio ServicioHistorias) {
	h := NuevoHistoriaHandler(servicio)
	api := r.Group("/api")
	proyectos := api.Group("/proyectos")
	{
		proyectos.POST("/:id/historias", h.Crear)
		proyectos.GET("/:id/historias", h.Listar)
		proyectos.PATCH("/:id/historias/:historiaId", h.Actualizar)
	}
}

// Crear maneja POST /api/proyectos/:id/historias.
func (h *HistoriaHandler) Crear(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}

	var dto altaHistoriaDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		responderJSONInvalido(c)
		return
	}

	historia, err := h.servicio.Crear(c.Request.Context(), proyectoID, domain.NuevaHistoria{
		Titulo:      dto.Titulo,
		Descripcion: dto.Descripcion,
		Prioridad:   dto.Prioridad,
		Estado:      dto.Estado,
	})
	if err != nil {
		traducirError(c, err)
		return
	}

	c.JSON(http.StatusCreated, construirRespuestaHistoria(historia))
}

// Listar maneja GET /api/proyectos/:id/historias.
func (h *HistoriaHandler) Listar(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}

	historias, err := h.servicio.Listar(c.Request.Context(), proyectoID)
	if err != nil {
		traducirError(c, err)
		return
	}

	respuestas := make([]respuestaHistoria, 0, len(historias))
	for _, historia := range historias {
		respuestas = append(respuestas, construirRespuestaHistoria(&historia))
	}

	c.JSON(http.StatusOK, respuestaListadoHistorias{
		Total:     len(respuestas),
		Historias: respuestas,
	})
}

// Actualizar maneja PATCH /api/proyectos/:id/historias/:historiaId.
func (h *HistoriaHandler) Actualizar(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}

	historiaID, ok := parsearID(c, "historiaId")
	if !ok {
		return
	}

	var dto actualizacionHistoriaDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		responderJSONInvalido(c)
		return
	}

	historia, err := h.servicio.Actualizar(c.Request.Context(), proyectoID, historiaID, domain.ActualizacionHistoria{
		Titulo:      dto.Titulo,
		Descripcion: dto.Descripcion,
		Prioridad:   dto.Prioridad,
		Estado:      dto.Estado,
	})
	if err != nil {
		traducirError(c, err)
		return
	}

	c.JSON(http.StatusOK, construirRespuestaHistoria(historia))
}
