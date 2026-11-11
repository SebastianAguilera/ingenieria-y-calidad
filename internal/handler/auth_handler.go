package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	"ingenieria-y-calidad/internal/domain"
)

// ServicioAuth es el puerto que el handler de login necesita del service.
type ServicioAuth interface {
	Login(ctx context.Context, email, password string) (string, *domain.Usuario, error)
}

// RegistrarRutasAuth stub de la fase RED: no registra POST /api/auth/login.
func RegistrarRutasAuth(r *gin.Engine, servicio ServicioAuth) {}
