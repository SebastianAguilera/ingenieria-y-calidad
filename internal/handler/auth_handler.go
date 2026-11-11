package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"ingenieria-y-calidad/internal/domain"
)

// ServicioAuth es el puerto que el handler de login necesita del service.
type ServicioAuth interface {
	Login(ctx context.Context, email, password string) (string, *domain.Usuario, error)
}

// loginDTO es el cuerpo de POST /api/auth/login.
type loginDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// respuestaLoginDTO es la salida del login: el token de la sesion y el usuario
// sin password_hash (CA-7).
type respuestaLoginDTO struct {
	Token   string              `json:"token"`
	Usuario respuestaUsuarioDTO `json:"usuario"`
}

// AuthHandler expone POST /api/auth/login.
type AuthHandler struct {
	servicio ServicioAuth
}

// NuevoAuthHandler construye el handler con inyeccion del servicio de login.
func NuevoAuthHandler(servicio ServicioAuth) *AuthHandler {
	return &AuthHandler{servicio: servicio}
}

// RegistrarRutasAuth registra las rutas de autenticacion bajo /api.
func RegistrarRutasAuth(r *gin.Engine, servicio ServicioAuth) {
	auth := NuevoAuthHandler(servicio)
	api := r.Group("/api")
	{
		api.POST("/auth/login", auth.Login)
	}
}

// Login atiende POST /api/auth/login. R-08 fija el orden de la validacion:
// JSON (400) -> campos (422) -> credenciales (401) -> exito (200).
func (h *AuthHandler) Login(c *gin.Context) {
	var dto loginDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		responderJSONInvalido(c)
		return
	}

	if domain.NormalizarEmail(dto.Email) == "" {
		traducirError(c, domain.ErrUsuarioEmailInvalido)
		return
	}
	if dto.Password == "" {
		traducirError(c, domain.ErrPasswordVacio)
		return
	}

	token, usuario, err := h.servicio.Login(c.Request.Context(), dto.Email, dto.Password)
	if err != nil {
		traducirError(c, err)
		return
	}
	if usuario == nil {
		traducirError(c, domain.ErrCredencialesInvalidas)
		return
	}

	c.JSON(http.StatusOK, respuestaLoginDTO{
		Token:   token,
		Usuario: construirRespuestaUsuarioDTO(usuario),
	})
}
