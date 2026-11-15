package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"ingenieria-y-calidad/internal/domain"
)

type ServicioAcceptanceCriteria interface {
	Crear(ctx context.Context, proyectoID, historiaID uint, nuevo domain.NuevoAcceptanceCriterion) (*domain.AcceptanceCriterion, error)
	Listar(ctx context.Context, proyectoID, historiaID uint) ([]domain.AcceptanceCriterion, error)
	Actualizar(ctx context.Context, proyectoID, historiaID, criterioID uint, actualizacion domain.ActualizacionAcceptanceCriterion) (*domain.AcceptanceCriterion, error)
	Eliminar(ctx context.Context, proyectoID, historiaID, criterioID uint) error
}

type AcceptanceCriterionHandler struct {
	servicio ServicioAcceptanceCriteria
}

func NuevoAcceptanceCriterionHandler(servicio ServicioAcceptanceCriteria) *AcceptanceCriterionHandler {
	return &AcceptanceCriterionHandler{servicio: servicio}
}

func RegistrarRutasAcceptanceCriteria(r *gin.Engine, servicio ServicioAcceptanceCriteria) {
	h := NuevoAcceptanceCriterionHandler(servicio)
	api := r.Group("/api")
	proyectos := api.Group("/proyectos")
	{
		rutasHistoria := proyectos.Group("/:id/historias/:historiaId/criterios")
		{
			rutasHistoria.POST("", h.Crear)
			rutasHistoria.GET("", h.Listar)
			rutasHistoria.PATCH("/:criterioId", h.Actualizar)
			rutasHistoria.DELETE("/:criterioId", h.Eliminar)
		}
	}
}

func (h *AcceptanceCriterionHandler) Crear(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}
	historiaID, ok := parsearID(c, "historiaId")
	if !ok {
		return
	}

	var dto altaAcceptanceCriterionDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		responderJSONInvalido(c)
		return
	}

	criterio, err := h.servicio.Crear(c.Request.Context(), proyectoID, historiaID, domain.NuevoAcceptanceCriterion{
		Descripcion: dto.Descripcion,
	})
	if err != nil {
		traducirError(c, err)
		return
	}

	c.JSON(http.StatusCreated, construirRespuestaAcceptanceCriterion(criterio))
}

func (h *AcceptanceCriterionHandler) Listar(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}
	historiaID, ok := parsearID(c, "historiaId")
	if !ok {
		return
	}

	criterios, err := h.servicio.Listar(c.Request.Context(), proyectoID, historiaID)
	if err != nil {
		traducirError(c, err)
		return
	}

	respuestas := make([]respuestaAcceptanceCriterion, 0, len(criterios))
	for _, cr := range criterios {
		respuestas = append(respuestas, construirRespuestaAcceptanceCriterion(&cr))
	}

	c.JSON(http.StatusOK, respuestaListadoAcceptanceCriteria{
		Total:     len(respuestas),
		Criterios: respuestas,
	})
}

func (h *AcceptanceCriterionHandler) Actualizar(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}
	historiaID, ok := parsearID(c, "historiaId")
	if !ok {
		return
	}
	criterioID, ok := parsearID(c, "criterioId")
	if !ok {
		return
	}

	var dto actualizacionAcceptanceCriterionDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		responderJSONInvalido(c)
		return
	}

	isMet, err := parsearIsMet(dto.IsMet)
	if err != nil {
		traducirError(c, err)
		return
	}

	criterio, err := h.servicio.Actualizar(c.Request.Context(), proyectoID, historiaID, criterioID, domain.ActualizacionAcceptanceCriterion{
		Descripcion: dto.Descripcion,
		IsMet:       isMet,
	})
	if err != nil {
		traducirError(c, err)
		return
	}

	c.JSON(http.StatusOK, construirRespuestaAcceptanceCriterion(criterio))
}

func (h *AcceptanceCriterionHandler) Eliminar(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}
	historiaID, ok := parsearID(c, "historiaId")
	if !ok {
		return
	}
	criterioID, ok := parsearID(c, "criterioId")
	if !ok {
		return
	}

	if err := h.servicio.Eliminar(c.Request.Context(), proyectoID, historiaID, criterioID); err != nil {
		traducirError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
