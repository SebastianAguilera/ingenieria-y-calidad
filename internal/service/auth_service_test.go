package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"ingenieria-y-calidad/internal/domain"
)

// Constantes de prueba de US-10: el secreto y la expiracion que inyecta el
// test al servicio son los mismos que en produccion llegan desde JWT_SECRET y
// JWT_EXPIRY.
const (
	secretoDePrueba    = "secreto-de-prueba-us10"
	expiracionDePrueba = 24 * time.Hour
	passwordDePrueba   = "Secreto123"
	emailDePrueba      = "ana@example.com"
	usuarioIDDePrueba  = uint(1)
)

// hashDePrueba genera un hash bcrypt real de la contrasena usada en los
// escenarios (E-01 a E-06, E-17): solo existe como hash, nunca en claro.
func hashDePrueba(t *testing.T) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(passwordDePrueba), bcrypt.DefaultCost)
	require.NoError(t, err)
	return string(hash)
}

func usuarioDePrueba(hash string) *domain.Usuario {
	return &domain.Usuario{
		ID:           usuarioIDDePrueba,
		Nombre:       "Ana Pérez",
		Email:        emailDePrueba,
		CreadoEn:     time.Date(2026, time.November, 10, 12, 0, 0, 0, time.UTC),
		PasswordHash: hash,
	}
}

func nuevoAuthServiceParaPruebas(usuarios domain.UsuarioRepository) *AuthService {
	return NewAuthService(usuarios, secretoDePrueba, expiracionDePrueba)
}

// verificarConSecreto valida un token emitido por el servicio con el mismo
// secreto que usa la aplicacion y devuelve los claims parseados.
func verificarConSecreto(t *testing.T, token string) (*jwt.Token, jwt.MapClaims) {
	t.Helper()
	parsed, err := jwt.Parse(token, func(tok *jwt.Token) (interface{}, error) {
		return []byte(secretoDePrueba), nil
	})
	require.NoError(t, err, "el token debe verificar con el secreto de configuracion")
	require.True(t, parsed.Valid)

	claims, ok := parsed.Claims.(jwt.MapClaims)
	require.True(t, ok)
	return parsed, claims
}

// TestLoginE01CredencialesValidasDevuelveTokenYUsuario cubre E-01, CA-1 y R-05:
// con un hash bcrypt valido el login devuelve un token verificable con el
// secreto y el usuario autenticado.
func TestLoginE01CredencialesValidasDevuelveTokenYUsuario(t *testing.T) {
	t.Parallel()

	repo := nuevoMockUsuarioRepository(t)
	repo.On("ObtenerPorEmail", mock.Anything, emailDePrueba).
		Return(usuarioDePrueba(hashDePrueba(t)), nil)

	token, usuario, err := nuevoAuthServiceParaPruebas(repo).
		Login(context.Background(), emailDePrueba, passwordDePrueba)

	require.NoError(t, err, "E-01")
	require.NotEmpty(t, token, "E-01: el login debe devolver un token")
	require.NotNil(t, usuario, "E-01: el login debe devolver el usuario")
	assert.Equal(t, usuarioIDDePrueba, usuario.ID)
	assert.Equal(t, emailDePrueba, usuario.Email)

	parsed, _ := verificarConSecreto(t, token)
	assert.Equal(t, "HS256", parsed.Header["alg"], "R-05: el algoritmo es HS256")
}

// TestLoginE02NormalizaElEmailAntesDeBuscar cubre E-02, R-03 y CL-01: con
// mayusculas y espacios el servicio normaliza (trim + minusculas) antes de
// llamar a ObtenerPorEmail y el resultado es el mismo que en E-01.
func TestLoginE02NormalizaElEmailAntesDeBuscar(t *testing.T) {
	t.Parallel()

	repo := nuevoMockUsuarioRepository(t)
	repo.On("ObtenerPorEmail", mock.Anything, emailDePrueba).
		Return(usuarioDePrueba(hashDePrueba(t)), nil)

	token, usuario, err := nuevoAuthServiceParaPruebas(repo).
		Login(context.Background(), "  ANA@EXAMPLE.COM  ", passwordDePrueba)

	require.NoError(t, err, "E-02")
	require.NotEmpty(t, token)
	require.NotNil(t, usuario)
	assert.Equal(t, emailDePrueba, usuario.Email, "el email devuelto queda normalizado")
	repo.AssertExpectations(t)
}

// TestLoginE03ContrasenaIncorrectaRespondeCredencialesInvalidas cubre E-03 y
// CA-2: 401 CREDENCIALES_INVALIDAS sin token y sin revelar el detalle.
func TestLoginE03ContrasenaIncorrectaRespondeCredencialesInvalidas(t *testing.T) {
	t.Parallel()

	repo := nuevoMockUsuarioRepository(t)
	repo.On("ObtenerPorEmail", mock.Anything, emailDePrueba).
		Return(usuarioDePrueba(hashDePrueba(t)), nil)

	token, usuario, err := nuevoAuthServiceParaPruebas(repo).
		Login(context.Background(), emailDePrueba, "OtraClave")

	assert.ErrorIs(t, err, domain.ErrCredencialesInvalidas, "E-03 / CA-2")
	assert.NotErrorIs(t, err, domain.ErrUsuarioNoEncontrado, "R-04")
	assert.Empty(t, token, "con credenciales invalidas no se emite token")
	assert.Nil(t, usuario)
}

// TestLoginE04EmailInexistenteRespondeElMismoErrorQueE03 cubre E-04 y CA-3
// (anti user-enumeration): el mensaje para un email inexistente debe ser
// identico byte a byte al de una contrasena incorrecta.
func TestLoginE04EmailInexistenteRespondeElMismoErrorQueE03(t *testing.T) {
	t.Parallel()

	// Caso A: contrasena incorrecta (E-03).
	repoConHash := nuevoMockUsuarioRepository(t)
	repoConHash.On("ObtenerPorEmail", mock.Anything, emailDePrueba).
		Return(usuarioDePrueba(hashDePrueba(t)), nil)
	_, _, errPasswordMala := nuevoAuthServiceParaPruebas(repoConHash).
		Login(context.Background(), emailDePrueba, "OtraClave")
	require.ErrorIs(t, errPasswordMala, domain.ErrCredencialesInvalidas, "E-03 / CA-2")

	// Caso B: email que no existe en la base (E-04).
	repoSinUsuario := nuevoMockUsuarioRepository(t)
	repoSinUsuario.On("ObtenerPorEmail", mock.Anything, "nadie@example.com").
		Return(nil, domain.ErrUsuarioNoEncontrado)
	token, usuario, errEmailInexistente := nuevoAuthServiceParaPruebas(repoSinUsuario).
		Login(context.Background(), "nadie@example.com", passwordDePrueba)

	require.ErrorIs(t, errEmailInexistente, domain.ErrCredencialesInvalidas, "E-04 / CA-3")
	assert.NotErrorIs(t, errEmailInexistente, domain.ErrUsuarioNoEncontrado,
		"en login no se usa ErrUsuarioNoEncontrado (R-04)")
	assert.Equal(t, errPasswordMala.Error(), errEmailInexistente.Error(),
		"CA-3: el mensaje debe ser exactamente el mismo que en E-03")
	assert.Empty(t, token)
	assert.Nil(t, usuario)
}

// TestLoginE05UsuarioSinHashRespondeElMismoError cubre E-05, R-09 y CL-03:
// usuarios sin password_hash (p. ej. seba antes del backfill) no llegan a
// bcrypt y devuelven el mismo 401 generico.
func TestLoginE05UsuarioSinHashRespondeElMismoError(t *testing.T) {
	t.Parallel()

	repo := nuevoMockUsuarioRepository(t)
	repo.On("ObtenerPorEmail", mock.Anything, emailDePrueba).
		Return(usuarioDePrueba(""), nil)

	token, usuario, err := nuevoAuthServiceParaPruebas(repo).
		Login(context.Background(), emailDePrueba, passwordDePrueba)

	require.ErrorIs(t, err, domain.ErrCredencialesInvalidas, "E-05 / R-09")
	assert.Equal(t, domain.ErrCredencialesInvalidas.Error(), err.Error(),
		"el mensaje debe ser identico al de E-03")
	assert.Empty(t, token)
	assert.Nil(t, usuario)
}

// TestLoginE06HashCorruptoRespondeElMismoError cubre E-06 y CL-06: un hash que
// bcrypt no puede parsear responde el 401 generico y no filtra el detalle
// interno de la falla.
func TestLoginE06HashCorruptoRespondeElMismoError(t *testing.T) {
	t.Parallel()

	repo := nuevoMockUsuarioRepository(t)
	repo.On("ObtenerPorEmail", mock.Anything, emailDePrueba).
		Return(usuarioDePrueba("esto-no-es-un-hash-bcrypt"), nil)

	token, _, err := nuevoAuthServiceParaPruebas(repo).
		Login(context.Background(), emailDePrueba, passwordDePrueba)

	require.ErrorIs(t, err, domain.ErrCredencialesInvalidas, "E-06 / CL-06")
	assert.Equal(t, domain.ErrCredencialesInvalidas.Error(), err.Error(),
		"no se revela el detalle interno del error")
	assert.Empty(t, token)
}

// TestLoginE15ClaimsDelToken cubre E-15 y R-05: HS256 fijo, sub como texto del
// id, email, y exp = iat + JWT_EXPIRY.
func TestLoginE15ClaimsDelToken(t *testing.T) {
	t.Parallel()

	repo := nuevoMockUsuarioRepository(t)
	repo.On("ObtenerPorEmail", mock.Anything, emailDePrueba).
		Return(usuarioDePrueba(hashDePrueba(t)), nil)

	token, _, err := nuevoAuthServiceParaPruebas(repo).
		Login(context.Background(), emailDePrueba, passwordDePrueba)
	require.NoError(t, err)

	parsed, claims := verificarConSecreto(t, token)
	assert.Equal(t, "HS256", parsed.Header["alg"], "R-05")
	assert.Equal(t, "1", claims["sub"], "R-05: sub es el id en formato texto")
	assert.Equal(t, emailDePrueba, claims["email"])

	iat, errIat := claims.GetIssuedAt()
	exp, errExp := claims.GetExpirationTime()
	require.NoError(t, errIat, "el token debe llevar iat")
	require.NoError(t, errExp, "el token debe llevar exp")
	require.NotNil(t, iat)
	require.NotNil(t, exp)
	assert.Equal(t, iat.Time.Add(expiracionDePrueba).Unix(), exp.Time.Unix(),
		"R-05: exp = iat + JWT_EXPIRY")
	assert.True(t, exp.Time.After(time.Now()), "el token recien emitido no esta vencido")
}

// TestLoginE17ComparaConBcrypt cubre E-17 y R-01 en la capa de servicio: la
// unica contrasena que valida es la que corresponde al hash bcrypt persistido,
// y la comparacion se hace con bcrypt y no con igualdad de strings.
func TestLoginE17ComparaConBcrypt(t *testing.T) {
	t.Parallel()

	hash := hashDePrueba(t)

	repoOk := nuevoMockUsuarioRepository(t)
	repoOk.On("ObtenerPorEmail", mock.Anything, emailDePrueba).
		Return(usuarioDePrueba(hash), nil)
	_, _, errOK := nuevoAuthServiceParaPruebas(repoOk).
		Login(context.Background(), emailDePrueba, passwordDePrueba)
	require.NoError(t, errOK, "E-17: el hash bcrypt valida la contrasena original")

	repoKo := nuevoMockUsuarioRepository(t)
	repoKo.On("ObtenerPorEmail", mock.Anything, emailDePrueba).
		Return(usuarioDePrueba(hash), nil)
	_, _, errKo := nuevoAuthServiceParaPruebas(repoKo).
		Login(context.Background(), emailDePrueba, passwordDePrueba+"x")
	assert.ErrorIs(t, errKo, domain.ErrCredencialesInvalidas)

	assert.NotEqual(t, hash, passwordDePrueba, "el hash nunca es la contrasena en claro")
}

// ---------------------------------------------------------------------------
// US-F09 - Registro de usuario con contrasena (TDD RED)
// ---------------------------------------------------------------------------

// TestRegistroE01CreaUsuarioConHashYDevuelveToken cubre E-02, CA-2, RB-03,
// RB-10 y RB-11: el alta valida, normaliza el email, persiste un hash bcrypt
// (nunca la contrasena en claro) y devuelve un token verificable con el mismo
// secreto que el login.
func TestRegistroE01CreaUsuarioConHashYDevuelveToken(t *testing.T) {
	t.Parallel()

	var persistido *domain.Usuario
	repo := nuevoMockUsuarioRepository(t)
	repo.On("Crear", mock.Anything, mock.AnythingOfType("*domain.Usuario")).
		Run(func(args mock.Arguments) {
			usuario := args.Get(1).(*domain.Usuario)
			usuario.ID = 7
			usuario.CreadoEn = time.Date(2026, time.November, 11, 12, 0, 0, 0, time.UTC)
			persistido = usuario
		}).
		Return(nil)

	token, usuario, err := nuevoAuthServiceParaPruebas(repo).
		Registrar(context.Background(), "  Ana Pérez  ", "  ANA@Example.com ", passwordDePrueba)

	require.NoError(t, err, "E-02 / CA-2")
	require.NotEmpty(t, token, "E-02 / CA-2: el registro devuelve el token de sesion")
	require.NotNil(t, usuario)

	// RB-10: nombre recortado y email normalizado (trim + minusculas).
	require.NotNil(t, persistido, "E-02: el usuario llega al repositorio")
	assert.Equal(t, "Ana Pérez", persistido.Nombre, "el nombre se recorta")
	assert.Equal(t, "ana@example.com", persistido.Email, "RB-10: email normalizado")

	// RB-11: la contrasena solo se persiste como hash bcrypt.
	require.NotEmpty(t, persistido.PasswordHash, "RB-11: el alta guarda el hash")
	assert.NotEqual(t, passwordDePrueba, persistido.PasswordHash, "RB-11")
	assert.NoError(t,
		bcrypt.CompareHashAndPassword([]byte(persistido.PasswordHash), []byte(passwordDePrueba)),
		"RB-11: el hash valida la contrasena original")

	// CA-2: el token cumple el mismo contrato que el login.
	parsed, claims := verificarConSecreto(t, token)
	require.True(t, parsed.Valid, "CA-2")
	assert.Equal(t, "7", claims["sub"], "CA-2: sub es el id en texto")
	assert.Equal(t, "ana@example.com", claims["email"], "CA-2")
	assert.NotNil(t, claims["exp"], "CA-2")

	assert.Equal(t, persistido, usuario, "el usuario devuelto es el persistido")
}

// TestRegistroE02EmailDuplicadoDevuelveErrUsuarioDuplicado cubre E-04, CA-4 y
// RB-10: el indice unico de email se traduce en el error sentinela que el
// handler responde como 422 USUARIO_DUPLICADO.
func TestRegistroE02EmailDuplicadoDevuelveErrUsuarioDuplicado(t *testing.T) {
	t.Parallel()

	repo := nuevoMockUsuarioRepository(t)
	repo.On("Crear", mock.Anything, mock.AnythingOfType("*domain.Usuario")).
		Return(domain.ErrUsuarioDuplicado)

	_, _, err := nuevoAuthServiceParaPruebas(repo).
		Registrar(context.Background(), "Ana Pérez", emailDePrueba, passwordDePrueba)

	assert.ErrorIs(t, err, domain.ErrUsuarioDuplicado, "E-04 / CA-4 / RB-10")
}

// TestRegistroE03ValidacionesPreviasNoPersistenNada cubre E-03, E-05, E-10 y
// CA-3: un dato invalido se rechaza antes de tocar el repositorio.
func TestRegistroE03ValidacionesPreviasNoPersistenNada(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre        string
		nombreValor   string
		email         string
		password      string
		errorEsperado error
		escenario     string
	}{
		{"nombre vacio", "   ", emailDePrueba, passwordDePrueba, domain.ErrUsuarioNombreVacio, "E-03 / CA-3"},
		{"nombre demasiado largo", strings.Repeat("a", 101), emailDePrueba, passwordDePrueba, domain.ErrUsuarioNombreLargo, "E-10"},
		{"email invalido", "Ana Pérez", "no-es-email", passwordDePrueba, domain.ErrUsuarioEmailInvalido, "E-05 / CA-5"},
		{"password vacio", "Ana Pérez", emailDePrueba, "", domain.ErrPasswordVacio, "E-03 / CA-3"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()

			repo := nuevoMockUsuarioRepository(t)

			_, _, err := nuevoAuthServiceParaPruebas(repo).
				Registrar(context.Background(), c.nombreValor, c.email, c.password)

			assert.ErrorIs(t, err, c.errorEsperado, c.escenario)
			repo.AssertNotCalled(t, "Crear", mock.Anything, mock.Anything,
				"sin persistir ante un dato invalido")
		})
	}
}

// TestRegistroE06ElUsuarioCreadoPuedeIniciarSesion cubre E-06 y CA-6: el hash
// generado por el registro es el que despues valida el login.
func TestRegistroE06ElUsuarioCreadoPuedeIniciarSesion(t *testing.T) {
	t.Parallel()

	var persistido *domain.Usuario
	repo := nuevoMockUsuarioRepository(t)
	repo.On("Crear", mock.Anything, mock.AnythingOfType("*domain.Usuario")).
		Run(func(args mock.Arguments) {
			usuario := args.Get(1).(*domain.Usuario)
			usuario.ID = 3
			usuario.CreadoEn = time.Now().UTC()
			persistido = usuario
		}).
		Return(nil)

	servicio := nuevoAuthServiceParaPruebas(repo)
	_, _, err := servicio.Registrar(context.Background(), "Ana Pérez", emailDePrueba, passwordDePrueba)
	require.NoError(t, err, "E-06 / CA-6")
	require.NotNil(t, persistido, "E-06: el usuario quedo persistido")

	repo.On("ObtenerPorEmail", mock.Anything, emailDePrueba).Return(persistido, nil)

	token, usuario, errLogin := servicio.Login(context.Background(), emailDePrueba, passwordDePrueba)

	require.NoError(t, errLogin, "E-06 / CA-6: el login valida el hash creado en el registro")
	require.NotEmpty(t, token)
	assert.Equal(t, emailDePrueba, usuario.Email)
}
