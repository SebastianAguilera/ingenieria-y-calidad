package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeveridadDefectoValida(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		severidad SeveridadDefecto
		esperado  bool
		escenario string
	}{
		{"minor", SeveridadMinor, true, "BDD-01"},
		{"major", SeveridadMajor, true, "BDD-01"},
		{"critical", SeveridadCritical, true, "BDD-01"},
		{"blocker", SeveridadBlocker, true, "BDD-01"},
		{"mayusculas", SeveridadDefecto("Major"), false, "BDD-02"},
		{"valor desconocido", SeveridadDefecto("urgente"), false, "BDD-02"},
		{"vacia", SeveridadDefecto(""), false, "BDD-02"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.esperado, c.severidad.Valida(), "escenario BDD %s", c.escenario)
		})
	}
}

func TestEstadoDefectoValido(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		estado    EstadoDefecto
		esperado  bool
		escenario string
	}{
		{"open", EstadoDefectoAbierto, true, "BDD-01"},
		{"in_progress", EstadoDefectoEnProgreso, true, "BDD-07"},
		{"resolved", EstadoDefectoResuelto, true, "BDD-08"},
		{"closed", EstadoDefectoCerrado, true, "BDD-09"},
		{"rejected", EstadoDefectoRechazado, true, "BDD-12"},
		{"valor desconocido", EstadoDefecto("fixed"), false, "BDD-12"},
		{"vacio", EstadoDefecto(""), false, "BDD-12"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.esperado, c.estado.Valido(), "escenario BDD %s", c.escenario)
		})
	}
}

func TestEstadoDefectoEsDeResolucion(t *testing.T) {
	t.Parallel()

	casos := []struct {
		estado    EstadoDefecto
		esperado  bool
		escenario string
	}{
		{EstadoDefectoAbierto, false, "BDD-10"},
		{EstadoDefectoEnProgreso, false, "BDD-10"},
		{EstadoDefectoResuelto, true, "BDD-08"},
		{EstadoDefectoCerrado, true, "BDD-09"},
		{EstadoDefectoRechazado, false, "BDD-11"},
	}

	for _, c := range casos {
		t.Run(string(c.estado), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.esperado, c.estado.EsDeResolucion(), "escenario BDD %s", c.escenario)
		})
	}
}

// TestTipoListadoDefectoValido cubre la regla de US-18 para el listado por
// Sprint: tipo obligatorio y sensible a mayúsculas (E-04, E-05).
func TestTipoListadoDefectoValido(t *testing.T) {
	casos := []struct {
		nombre    string
		tipo      TipoListadoDefecto
		esperado  bool
		escenario string
	}{
		{"deteccion", TipoListadoDeteccion, true, "E-01"},
		{"resolucion", TipoListadoResolucion, true, "E-02"},
		{"vacio", TipoListadoDefecto(""), false, "E-04"},
		{"mayusculas", TipoListadoDefecto("Deteccion"), false, "E-05"},
		{"desconocido", TipoListadoDefecto("todos"), false, "E-05"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			assert.Equal(t, c.esperado, c.tipo.Valido(), "escenario BDD %s", c.escenario)
		})
	}
}
