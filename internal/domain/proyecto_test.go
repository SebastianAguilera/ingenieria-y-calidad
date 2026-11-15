package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEstadoProyectoValido(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		estado    EstadoProyecto
		esperado  bool
		escenario string
	}{
		{"estado Activo", EstadoActivo, true, "E-36"},
		{"estado Cerrado", EstadoCerrado, true, "E-36"},
		{"estado activo en minuscula", EstadoProyecto("activo"), false, "E-36"},
		{"estado CERRADO en mayuscula", EstadoProyecto("CERRADO"), false, "E-36"},
		{"estado vacio", EstadoProyecto(""), false, "E-36"},
		{"estado desconocido", EstadoProyecto("Archivado"), false, "E-36"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.esperado, c.estado.Valido(), "escenario BDD %s", c.escenario)
		})
	}
}

func TestNormalizarEmail(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		entrada   string
		esperado  string
		escenario string
	}{
		{"deja intacto un email en minuscula", "ada@utn.edu.ar", "ada@utn.edu.ar", "E-41"},
		{"convierte a minusculas", "Ada@UTN.edu.ar", "ada@utn.edu.ar", "E-17"},
		{"recorta espacios de los extremos", "  ada@utn.edu.ar\t", "ada@utn.edu.ar", "E-17"},
		{"combina recorte y minusculas", "\n Ada@UTN.EDU.AR \n", "ada@utn.edu.ar", "E-17"},
		{"devuelve vacio para cadena vacia", "", "", "E-15"},
		{"devuelve vacio para solo espacios", "   \t  ", "", "E-15"},
		{"no toca los espacios embebidos", "ada lovelace@utn.edu.ar", "ada lovelace@utn.edu.ar", "E-42"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.esperado, NormalizarEmail(c.entrada), "escenario BDD %s", c.escenario)
		})
	}
}

func TestEmailValido(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		email     string
		esperado  bool
		escenario string
	}{
		{"acepta un email simple", "ada@utn.edu.ar", true, "E-16"},
		{"acepta un email con punto y guion", "jazmin.perez-castro@utn.edu.ar", true, "E-16"},
		{"rechaza cadena vacia", "", false, "E-15"},
		{"rechaza solo espacios", "   ", false, "E-15"},
		{"rechaza un texto sin arroba", "ada.utn.edu.ar", false, "E-15"},
		{"rechaza dos arrobas", "ada@@utn.edu.ar", false, "E-15"},
		{"rechaza un arroba al inicio", "@utn.edu.ar", false, "E-15"},
		{"rechaza un arroba al final", "ada@", false, "E-15"},
		{"rechaza espacios embebidos", "ada lovelace@utn.edu.ar", false, "E-42"},
		{"rechaza la forma Nombre <correo>", "Ada Lovelace <ada@utn.edu.ar>", false, "E-14"},
		{"rechaza longitud 151", strings.Repeat("a", 140) + "@utn.edu.ar", false, "E-16"},
		{"acepta longitud 150 exacta", strings.Repeat("a", 139) + "@utn.edu.ar", true, "E-16"},
		{"mide la longitud sobre el email normalizado", strings.Repeat("a", 134) + "@utn.edu.ar", true, "E-41"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.esperado, EmailValido(c.email), "escenario BDD %s", c.escenario)
		})
	}
}
