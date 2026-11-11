package service

import (
	"context"
	"time"

	"ingenieria-y-calidad/internal/domain"
)

// AuthService emite el JWT de la sesion. Stub de la fase RED: todavia no
// implementa la validacion de credenciales ni la firma.
type AuthService struct{}

// NewAuthService inyecta el repositorio de usuarios, JWT_SECRET y JWT_EXPIRY.
func NewAuthService(usuarios domain.UsuarioRepository, secreto string, expiracion time.Duration) *AuthService {
	return &AuthService{}
}

// Login todavia no esta implementado: siempre devuelve sin token y sin error.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, *domain.Usuario, error) {
	return "", nil, nil
}
