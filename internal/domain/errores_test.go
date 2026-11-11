package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestErroresSentinelasDeAutenticacion cubre las condiciones de error de la
// SDD seccion 7: los cuatro sentinelas nuevos existen, no estan vacios y son
// errores distintos entre si y de ErrUsuarioNoEncontrado, que en login no debe
// usarse (R-04, anti user-enumeration).
func TestErroresSentinelasDeAutenticacion(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		err       error
		mensaje   string
		estado    int
		escenario string
	}{
		{
			nombre:    "credenciales invalidas",
			err:       ErrCredencialesInvalidas,
			mensaje:   "email o contraseña incorrectos",
			estado:    401,
			escenario: "E-03 E-04 E-05 E-06",
		},
		{
			nombre:    "token invalido",
			err:       ErrTokenInvalido,
			mensaje:   "el token de autenticación no es válido o ha expirado",
			estado:    401,
			escenario: "E-13",
		},
		{
			nombre:    "token requerido",
			err:       ErrTokenRequerido,
			mensaje:   "la solicitud no incluye un token de autenticación",
			estado:    401,
			escenario: "E-11 E-12",
		},
		{
			nombre:    "password vacio",
			err:       ErrPasswordVacio,
			mensaje:   "la contraseña es obligatoria",
			estado:    422,
			escenario: "E-07",
		},
	}

	vistos := map[string]bool{}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Error(t, c.err, "escenario %s", c.escenario)
			assert.Equal(t, c.mensaje, c.err.Error(),
				"el mensaje es contrato de la API, escenario %s", c.escenario)
			assert.NotZero(t, c.estado)
		})

		if c.err != nil {
			vistos[c.err.Error()] = true
		}
	}

	assert.Len(t, vistos, len(casos), "cada condicion de error necesita su propio sentinel")
}

// TestErroresDeAuthNoConfundenConUsuarioNoEncontrado protege R-04: en login no
// se debe usar ErrUsuarioNoEncontrado, porque devolveria 404 USUARIO_NO_ENCONTRADO
// y revelaria si el email existe.
func TestErroresDeAuthNoConfundenConUsuarioNoEncontrado(t *testing.T) {
	t.Parallel()

	sentinelaDeLogin := []error{ErrCredencialesInvalidas, ErrPasswordVacio}
	for _, err := range sentinelaDeLogin {
		assert.False(t, errors.Is(err, ErrUsuarioNoEncontrado),
			"%v no debe envolver a ErrUsuarioNoEncontrado", err)
		assert.NotEqual(t, ErrUsuarioNoEncontrado, err)
	}

	assert.False(t, errors.Is(ErrTokenRequerido, ErrTokenInvalido),
		"TOKEN_REQUERIDO y TOKEN_INVALIDO son condiciones distintas (R-06)")
	assert.False(t, errors.Is(ErrTokenInvalido, ErrTokenRequerido))
}
