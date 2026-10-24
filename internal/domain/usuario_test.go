package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
