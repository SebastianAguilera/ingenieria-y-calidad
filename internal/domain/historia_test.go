package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrioridadHistoriaValida(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		prioridad PrioridadHistoria
		esperado  bool
		escenario string
	}{
		{"prioridad Alta", PrioridadAlta, true, "E-10"},
		{"prioridad Media", PrioridadMedia, true, "E-10"},
		{"prioridad Baja", PrioridadBaja, true, "E-10"},
		{"prioridad alta en minuscula", PrioridadHistoria("alta"), false, "E-10"},
		{"prioridad vacia", PrioridadHistoria(""), false, "E-10"},
		{"prioridad desconocida", PrioridadHistoria("Urgente"), false, "E-10"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.esperado, c.prioridad.Valido(), "escenario BDD %s", c.escenario)
		})
	}
}

func TestEstadoHistoriaValido(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		estado    EstadoHistoria
		esperado  bool
		escenario string
	}{
		{"estado Backlog", EstadoHistoriaBacklog, true, "E-11"},
		{"estado En Progreso", EstadoHistoriaEnProgreso, true, "E-11"},
		{"estado Hecho", EstadoHistoriaHecho, true, "E-11"},
		{"estado backlog en minuscula", EstadoHistoria("backlog"), false, "E-11"},
		{"estado vacio", EstadoHistoria(""), false, "E-11"},
		{"estado desconocido", EstadoHistoria("En revisión"), false, "E-11"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.esperado, c.estado.Valido(), "escenario BDD %s", c.escenario)
		})
	}
}

func TestStoryPointsValidos(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		valor     float64
		esperado  bool
		escenario string
	}{
		{"cero", 0, true, "BDD-03"},
		{"medio punto", 0.5, true, "BDD-04"},
		{"uno", 1, true, "BDD-01"},
		{"dos", 2, true, "BDD-01"},
		{"tres", 3, true, "BDD-01"},
		{"cinco", 5, true, "BDD-01"},
		{"ocho", 8, true, "BDD-02"},
		{"trece", 13, true, "BDD-06"},
		{"veintiuno", 21, true, "BDD-01"},
		{"cuatro fuera de escala", 4, false, "BDD-07"},
		{"veintidos fuera de escala", 22, false, "BDD-10"},
		{"un decimal fuera de escala", 1.5, false, "BDD-07"},
		{"negativo", -1, false, "BDD-08"},
		{"medio negativo", -0.5, false, "BDD-08"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.esperado, StoryPointsValidos(c.valor), "escenario BDD %s", c.escenario)
		})
	}
}

func TestFormatearCodigoHistoria(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		numero    int
		esperado  string
		escenario string
	}{
		{"primera historia", 1, "US-01", "E-01"},
		{"segunda historia", 2, "US-02", "E-12"},
		{"decima historia", 10, "US-10", "E-13"},
		{"centesima historia", 100, "US-100", "E-13"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.esperado, FormatearCodigoHistoria(c.numero), "escenario BDD %s", c.escenario)
		})
	}
}
