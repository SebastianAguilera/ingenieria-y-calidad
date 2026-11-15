package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"ingenieria-y-calidad/internal/domain"
)

// ServicioUsuarios es el puerto que el handler de usuarios necesita del service.
type ServicioUsuarios interface {
	Crear(ctx context.Context, entrada domain.UsuarioInput) (*domain.Usuario, error)
	Listar(ctx context.Context) ([]domain.Usuario, error)
	ObtenerPorID(ctx context.Context, id uint) (*domain.Usuario, error)
}

// UsuarioHandler expone el módulo de gestión independiente de usuarios por HTTP.
type UsuarioHandler struct {
	servicio ServicioUsuarios
}

// NuevoUsuarioHandler construye el controlador con inyección de dependencias.
func NuevoUsuarioHandler(servicio ServicioUsuarios) *UsuarioHandler {
	return &UsuarioHandler{servicio: servicio}
}

// RegistrarRutasUsuarios registra las rutas de US-08 bajo el grupo /api.
func RegistrarRutasUsuarios(r *gin.Engine, servicio ServicioUsuarios) {
	h := NuevoUsuarioHandler(servicio)
	api := r.Group("/api")
	usuarios := api.Group("/usuarios")
	{
		usuarios.POST("", h.Crear)
		usuarios.GET("", h.Listar)
	}
}

// Crear maneja POST /api/usuarios.
func (h *UsuarioHandler) Crear(c *gin.Context) {
	var dto altaUsuarioDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		responderJSONInvalido(c)
		return
	}

	usuario, err := h.servicio.Crear(c.Request.Context(), domain.UsuarioInput{
		Nombre: dto.Nombre,
		Email:  dto.Email,
	})
	if err != nil {
		traducirError(c, err)
		return
	}

	c.JSON(http.StatusCreated, construirRespuestaUsuarioDTO(usuario))
}

// Listar maneja GET /api/usuarios.
func (h *UsuarioHandler) Listar(c *gin.Context) {
	usuarios, err := h.servicio.Listar(c.Request.Context())
	if err != nil {
		traducirError(c, err)
		return
	}

	c.JSON(http.StatusOK, construirRespuestaUsuariosDTO(usuarios))
}
