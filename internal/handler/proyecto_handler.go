package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ingenieria-y-calidad/internal/domain"
)

// ServicioProyectos es el puerto que el handler necesita del service. Se define
// en el consumidor para que los tests puedan inyectar un mock.
type ServicioProyectos interface {
	Crear(ctx context.Context, nuevo domain.NuevoProyecto) (*domain.Proyecto, error)
	Listar(ctx context.Context) ([]domain.Proyecto, error)
	ObtenerPorID(ctx context.Context, id uint) (*domain.Proyecto, error)
	Actualizar(ctx context.Context, id uint, actualizacion domain.ActualizacionProyecto) (*domain.Proyecto, error)
	CambiarEstado(ctx context.Context, id uint, estado domain.EstadoProyecto) (*domain.Proyecto, error)
	Eliminar(ctx context.Context, id uint) error
	AgregarIntegrante(ctx context.Context, proyectoID uint, entrada domain.IntegranteInput) (*domain.Integrante, error)
	QuitarIntegrante(ctx context.Context, proyectoID, integranteID uint) error
}

// ProyectoHandler expone el modulo de gestion de proyectos por HTTP.
type ProyectoHandler struct {
	servicio ServicioProyectos
}

// NuevoProyectoHandler construye el controlador con inyeccion de dependencias.
func NuevoProyectoHandler(servicio ServicioProyectos) *ProyectoHandler {
	return &ProyectoHandler{servicio: servicio}
}

// RegistrarRutasProyectos registra las ocho rutas de US-01 bajo el grupo /api.
func RegistrarRutasProyectos(r *gin.Engine, servicio ServicioProyectos) {
	h := NuevoProyectoHandler(servicio)
	api := r.Group("/api")
	proyectos := api.Group("/proyectos")
	{
		proyectos.POST("", h.Crear)
		proyectos.GET("", h.Listar)
		proyectos.GET("/:id", h.ObtenerPorID)
		proyectos.PUT("/:id", h.Actualizar)
		proyectos.PATCH("/:id/estado", h.CambiarEstado)
		proyectos.DELETE("/:id", h.Eliminar)
		proyectos.POST("/:id/integrantes", h.AgregarIntegrante)
		proyectos.DELETE("/:id/integrantes/:integranteId", h.QuitarIntegrante)
	}
}

// Crear maneja POST /api/proyectos.
func (h *ProyectoHandler) Crear(c *gin.Context) {
	var dto altaProyectoDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		responderJSONInvalido(c)
		return
	}

	fechaInicio, err := parsearFechaObligatoria(dto.FechaInicio)
	if err != nil {
		responderFechaInvalida(c, "fecha_inicio")
		return
	}

	fechaFin, err := parsearFechaOpcional(dto.FechaFin)
	if err != nil {
		responderFechaInvalida(c, "fecha_fin")
		return
	}

	integrantes := convertirEntradas(dto.Integrantes)

	proyecto, err := h.servicio.Crear(c.Request.Context(), domain.NuevoProyecto{
		Nombre:      dto.Nombre,
		FechaInicio: fechaInicio,
		FechaFin:    fechaFin,
		Integrantes: integrantes,
	})
	if err != nil {
		traducirError(c, err)
		return
	}

	c.JSON(http.StatusCreated, construirRespuestaProyecto(proyecto, true))
}

// Listar maneja GET /api/proyectos.
func (h *ProyectoHandler) Listar(c *gin.Context) {
	proyectos, err := h.servicio.Listar(c.Request.Context())
	if err != nil {
		traducirError(c, err)
		return
	}

	respuestas := make([]respuestaProyecto, 0, len(proyectos))
	for _, proyecto := range proyectos {
		respuestas = append(respuestas, construirRespuestaProyecto(&proyecto, false))
	}

	c.JSON(http.StatusOK, respuestaListadoProyectos{
		Total:     len(respuestas),
		Proyectos: respuestas,
	})
}

// ObtenerPorID maneja GET /api/proyectos/:id.
func (h *ProyectoHandler) ObtenerPorID(c *gin.Context) {
	id, ok := parsearID(c, "id")
	if !ok {
		return
	}

	proyecto, err := h.servicio.ObtenerPorID(c.Request.Context(), id)
	if err != nil {
		traducirError(c, err)
		return
	}

	c.JSON(http.StatusOK, construirRespuestaProyecto(proyecto, true))
}

// Actualizar maneja PUT /api/proyectos/:id.
func (h *ProyectoHandler) Actualizar(c *gin.Context) {
	id, ok := parsearID(c, "id")
	if !ok {
		return
	}

	var dto actualizacionProyectoDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		responderJSONInvalido(c)
		return
	}

	fechaInicio, err := parsearFechaObligatoria(dto.FechaInicio)
	if err != nil {
		responderFechaInvalida(c, "fecha_inicio")
		return
	}

	fechaFin, err := parsearFechaOpcional(dto.FechaFin)
	if err != nil {
		responderFechaInvalida(c, "fecha_fin")
		return
	}

	proyecto, err := h.servicio.Actualizar(c.Request.Context(), id, domain.ActualizacionProyecto{
		Nombre:      dto.Nombre,
		FechaInicio: fechaInicio,
		FechaFin:    fechaFin,
	})
	if err != nil {
		traducirError(c, err)
		return
	}

	c.JSON(http.StatusOK, construirRespuestaProyecto(proyecto, true))
}

// CambiarEstado maneja PATCH /api/proyectos/:id/estado.
func (h *ProyectoHandler) CambiarEstado(c *gin.Context) {
	id, ok := parsearID(c, "id")
	if !ok {
		return
	}

	var dto cambioEstadoDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		responderJSONInvalido(c)
		return
	}

	proyecto, err := h.servicio.CambiarEstado(c.Request.Context(), id, dto.Estado)
	if err != nil {
		traducirError(c, err)
		return
	}

	c.JSON(http.StatusOK, construirRespuestaProyecto(proyecto, true))
}

// Eliminar maneja DELETE /api/proyectos/:id.
func (h *ProyectoHandler) Eliminar(c *gin.Context) {
	id, ok := parsearID(c, "id")
	if !ok {
		return
	}

	if err := h.servicio.Eliminar(c.Request.Context(), id); err != nil {
		traducirError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// AgregarIntegrante maneja POST /api/proyectos/:id/integrantes.
func (h *ProyectoHandler) AgregarIntegrante(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}

	var dto altaIntegranteDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		responderJSONInvalido(c)
		return
	}

	integrante, err := h.servicio.AgregarIntegrante(c.Request.Context(), proyectoID, domain.IntegranteInput{
		Nombre: dto.Nombre,
		Email:  dto.Email,
	})
	if err != nil {
		traducirError(c, err)
		return
	}

	c.JSON(http.StatusCreated, construirRespuestaVinculo(integrante, proyectoID))
}

// QuitarIntegrante maneja DELETE /api/proyectos/:id/integrantes/:integranteId.
func (h *ProyectoHandler) QuitarIntegrante(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}

	integranteID, ok := parsearID(c, "integranteId")
	if !ok {
		return
	}

	if err := h.servicio.QuitarIntegrante(c.Request.Context(), proyectoID, integranteID); err != nil {
		traducirError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func parsearID(c *gin.Context, param string) (uint, bool) {
	valor := c.Param(param)
	id, err := strconv.ParseUint(valor, 10, 64)
	if err != nil || id == 0 {
		responderParametroInvalido(c, param)
		return 0, false
	}
	return uint(id), true
}

func parsearFecha(valor string) (time.Time, error) {
	return time.Parse(formatoFecha, valor)
}

// parsearFechaObligatoria distingue entre un campo ausente o vacio, que deja
// en manos del service la regla de negocio, y un campo presente con formato
// invalido, que es un error sintactico de la capa HTTP.
func parsearFechaObligatoria(valor string) (time.Time, error) {
	if strings.TrimSpace(valor) == "" {
		return time.Time{}, nil
	}
	return parsearFecha(valor)
}

func parsearFechaOpcional(valor *string) (*time.Time, error) {
	if valor == nil || *valor == "" {
		return nil, nil
	}
	fecha, err := time.Parse(formatoFecha, *valor)
	if err != nil {
		return nil, err
	}
	return &fecha, nil
}

func convertirEntradas(entradas []integracionDTO) []domain.IntegranteInput {
	if len(entradas) == 0 {
		return nil
	}
	convertidas := make([]domain.IntegranteInput, 0, len(entradas))
	for _, entrada := range entradas {
		convertidas = append(convertidas, domain.IntegranteInput{
			Nombre: entrada.Nombre,
			Email:  entrada.Email,
		})
	}
	return convertidas
}
