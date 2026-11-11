package domain

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRolesDeProyecto(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		rol    RolProyecto
		valor  string
	}{
		{"constructor de producto", RolConstructorDeProducto, "Constructor de Producto"},
		{"arquitecto de producto", RolArquitectoDeProducto, "Arquitecto de Producto"},
		{"facilitador agil", RolFacilitadorAgil, "Facilitador Ágil"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.valor, string(c.rol))
		})
	}
}

func TestErroresDeUsuarioYRoles(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		err    error
	}{
		{"usuario duplicado", ErrUsuarioDuplicado},
		{"usuario no encontrado", ErrUsuarioNoEncontrado},
		{"rol invalido", ErrRolInvalido},
		{"usuario ya asociado", ErrUsuarioYaAsociado},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			assert.Error(t, c.err)
			assert.NotEmpty(t, c.err.Error())
		})
	}
}

func TestRolProyectoValido(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre   string
		rol      RolProyecto
		esperado bool
	}{
		{"constructor de producto", RolConstructorDeProducto, true},
		{"arquitecto de producto", RolArquitectoDeProducto, true},
		{"facilitador agil", RolFacilitadorAgil, true},
		{"rol vacio", RolProyecto(""), false},
		{"rol en minusculas", RolProyecto("constructor de producto"), false},
		{"rol desconocido", RolProyecto("Scrum Master"), false},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.esperado, c.rol.Valido())
		})
	}
}

func TestParseRolProyecto(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre   string
		input    string
		esperado RolProyecto
		err      error
	}{
		{"constructor de producto", "Constructor de Producto", RolConstructorDeProducto, nil},
		{"arquitecto de producto", "Arquitecto de Producto", RolArquitectoDeProducto, nil},
		{"facilitador agil", "Facilitador Ágil", RolFacilitadorAgil, nil},
		{"rol invalido", "Scrum Master", "", ErrRolInvalido},
		{"rol vacio", "", "", ErrRolInvalido},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			rol, err := ParseRolProyecto(c.input)
			if c.err != nil {
				assert.ErrorIs(t, err, c.err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, c.esperado, rol)
		})
	}
}

// TestPasswordHashNoSeSerializa cubre E-18, CA-7 y R-02: el campo PasswordHash
// lleva tag json:"-" y por lo tanto ninguna respuesta JSON puede contener ni el
// nombre del campo ni el hash bcrypt almacenado.
func TestPasswordHashNoSeSerializa(t *testing.T) {
	t.Parallel()

	hash := "$2a$10$abcdefghijklmnopqrstuvABCDEFGHIJKLMNOPQRSTUVWXYZ01234"
	usuario := Usuario{
		ID:           1,
		Nombre:       "Ana Pérez",
		Email:        "ana@example.com",
		PasswordHash: hash,
	}

	campo, encontrado := reflect.TypeOf(usuario).FieldByName("PasswordHash")
	require.True(t, encontrado, "el usuario debe exponer PasswordHash")
	assert.Equal(t, "-", campo.Tag.Get("json"),
		"PasswordHash debe declararse con json:\"-\" (R-02)")

	serializado, err := json.Marshal(usuario)
	require.NoError(t, err)

	assert.NotContains(t, string(serializado), "password_hash",
		"ningun JSON del sistema puede exponer la columna, CA-7")
	assert.NotContains(t, string(serializado), hash,
		"ningun JSON del sistema puede exponer el hash, CA-7")
	assert.Contains(t, string(serializado), "ana@example.com")
}
