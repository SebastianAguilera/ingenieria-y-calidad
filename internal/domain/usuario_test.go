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
