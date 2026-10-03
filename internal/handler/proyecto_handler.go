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
	AgregarUsuario(ctx context.Context, proyectoID uint, entrada domain.UsuarioInput) (*domain.Usuario, error)
	QuitarUsuario(ctx context.Context, proyectoID, usuarioID uint) error
}

// ProyectoHandler expone el modulo de gestion de proyectos por HTTP.
type ProyectoHandler struct {
	servicio ServicioProyectos
}

// NuevoProyectoHandler construye el controlador con inyeccion de dependencias.
func NuevoProyectoHandler(servicio ServicioProyectos) *ProyectoHandler {
	return &ProyectoHandler{servicio: servicio}
}

// RegistrarRutasProyectos registra las rutas de US-01 bajo el grupo /api.
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
		proyectos.POST("/:id/usuarios", h.AgregarUsuario)
		proyectos.DELETE("/:id/usuarios/:usuarioId", h.QuitarUsuario)
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

	usuarios := convertirEntradas(dto.Usuarios)

	proyecto, err := h.servicio.Crear(c.Request.Context(), domain.NuevoProyecto{
		Nombre:      dto.Nombre,
		FechaInicio: fechaInicio,
		FechaFin:    fechaFin,
		Usuarios:    usuarios,
	})
	if err != nil {
		traducirErrorDeAltaConEquipo(c, err)
		return
	}

	c.JSON(http.StatusCreated, construirRespuestaProyecto(proyecto))
}

// Listar maneja GET /api/proyectos.
func (h *ProyectoHandler) Listar(c *gin.Context) {
	proyectos, err := h.servicio.Listar(c.Request.Context())
	if err != nil {
		traducirError(c, err)
		return
	}

	respuestas := make([]respuestaProyectoListado, 0, len(proyectos))
	for _, proyecto := range proyectos {
		respuestas = append(respuestas, construirRespuestaListado(&proyecto))
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

	c.JSON(http.StatusOK, construirRespuestaProyecto(proyecto))
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
		traducirErrorDeActualizacion(c, err)
		return
	}

	c.JSON(http.StatusOK, construirRespuestaProyecto(proyecto))
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
		traducirErrorDeCambioEstado(c, err)
		return
	}

	c.JSON(http.StatusOK, construirRespuestaProyecto(proyecto))
}

// Eliminar maneja DELETE /api/proyectos/:id.
func (h *ProyectoHandler) Eliminar(c *gin.Context) {
	id, ok := parsearID(c, "id")
	if !ok {
		return
	}

	err := h.servicio.Eliminar(c.Request.Context(), id)
	if err != nil {
		traducirErrorDeEliminacion(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// AgregarUsuario maneja POST /api/proyectos/:id/usuarios.
func (h *ProyectoHandler) AgregarUsuario(c *gin.Context) {
	id, ok := parsearID(c, "id")
	if !ok {
		return
	}

	var dto altaUsuarioDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		responderJSONInvalido(c)
		return
	}

	usuario, err := h.servicio.AgregarUsuario(c.Request.Context(), id, domain.UsuarioInput{
		Nombre: dto.Nombre,
		Email:  dto.Email,
	})
	if err != nil {
		traducirErrorDeAsociacion(c, err)
		return
	}

	c.JSON(http.StatusCreated, construirRespuestaVinculo(usuario, id))
}

// QuitarUsuario maneja DELETE /api/proyectos/:id/usuarios/:usuarioId.
func (h *ProyectoHandler) QuitarUsuario(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}

	usuarioID, ok := parsearID(c, "usuarioId")
	if !ok {
		return
	}

	err := h.servicio.QuitarUsuario(c.Request.Context(), proyectoID, usuarioID)
	if err != nil {
		traducirErrorDeDesvinculacion(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func convertirEntradas(dtos []usuarioDTO) []domain.UsuarioInput {
	entradas := make([]domain.UsuarioInput, len(dtos))
	for i, dto := range dtos {
		entradas[i] = domain.UsuarioInput{
			Nombre: dto.Nombre,
			Email:  dto.Email,
		}
	}
	return entradas
}

func parsearID(c *gin.Context, param string) (uint, bool) {
	valStr := c.Param(param)
	val, err := strconv.ParseUint(valStr, 10, 64)
	if err != nil || val == 0 {
		responderParametroInvalido(c)
		return 0, false
	}
	return uint(val), true
}

func parsearFechaObligatoria(s string) (time.Time, error) {
	if strings.TrimSpace(s) == "" {
		return time.Time{}, nil
	}
	return time.Parse(formatoFecha, strings.TrimSpace(s))
}

func parsearFechaOpcional(s *string) (*time.Time, error) {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil, nil
	}
	t, err := time.Parse(formatoFecha, strings.TrimSpace(*s))
	if err != nil {
		return nil, err
	}
	return &t, nil
}
