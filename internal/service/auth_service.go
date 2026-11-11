package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"ingenieria-y-calidad/internal/domain"
)

// AuthService implementa las reglas de negocio de US-10: valida la contrasena
// contra el hash bcrypt persistido y emite el JWT de la sesion.
type AuthService struct {
	usuarios   domain.UsuarioRepository
	secreto    string
	expiracion time.Duration
}

// NewAuthService inyecta el repositorio de usuarios, JWT_SECRET y JWT_EXPIRY.
func NewAuthService(usuarios domain.UsuarioRepository, secreto string, expiracion time.Duration) *AuthService {
	return &AuthService{usuarios: usuarios, secreto: secreto, expiracion: expiracion}
}

// Login valida las credenciales y devuelve el token firmado junto al usuario
// autenticado. R-04: email inexistente, usuario sin hash y contrasena mala
// responden exactamente el mismo error, para no revelar si el email existe.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, *domain.Usuario, error) {
	normalizado := domain.NormalizarEmail(email)

	usuario, err := s.usuarios.ObtenerPorEmail(ctx, normalizado)
	if err != nil {
		if errors.Is(err, domain.ErrUsuarioNoEncontrado) {
			return "", nil, domain.ErrCredencialesInvalidas
		}
		return "", nil, fmt.Errorf("buscando usuario para el login: %w", err)
	}
	if usuario == nil {
		return "", nil, fmt.Errorf("buscando usuario para el login: el repositorio no devolvio usuario para %q", normalizado)
	}
	if usuario.PasswordHash == "" {
		return "", nil, domain.ErrCredencialesInvalidas
	}

	// R-01: la comparacion siempre es bcrypt, nunca igualdad de strings. El
	// detalle de la falla (hash corrupto, por ejemplo) solo va al log.
	if err := bcrypt.CompareHashAndPassword([]byte(usuario.PasswordHash), []byte(password)); err != nil {
		log.Printf("login fallido para %s: %v", normalizado, err)
		return "", nil, domain.ErrCredencialesInvalidas
	}

	token, err := s.emitirToken(usuario)
	if err != nil {
		return "", nil, fmt.Errorf("emitiendo el token de sesion: %w", err)
	}
	return token, usuario, nil
}

// Registrar da de alta un usuario con su contrasena en hash bcrypt y devuelve
// el token de sesion para el auto-login (US-F09). Stub de TDD RED: la
// implementacion completa llega en el commit GREEN.
func (s *AuthService) Registrar(ctx context.Context, nombre, email, password string) (string, *domain.Usuario, error) {
	return "", nil, nil
}

// emitirToken firma el JWT con HS256: sub (id en texto), email, iat y
// exp = iat + JWT_EXPIRY.
func (s *AuthService) emitirToken(usuario *domain.Usuario) (string, error) {
	ahora := time.Now()
	claims := jwt.MapClaims{
		"sub":   fmt.Sprintf("%d", usuario.ID),
		"email": usuario.Email,
		"iat":   ahora.Unix(),
		"exp":   ahora.Add(s.expiracion).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	firmado, err := token.SignedString([]byte(s.secreto))
	if err != nil {
		return "", err
	}
	return firmado, nil
}
