package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHorasEstimadasValidas(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		valor     float64
		esperado  bool
		escenario string
	}{
		{"entero positivo", 5, true, "BDD-01"},
		{"decimal", 4.5, true, "BDD-01"},
		{"decimal pequeno", 0.25, true, "CL-01"},
		{"cero", 0, false, "BDD-09"},
		{"negativo", -2, false, "BDD-10"},
		{"medio negativo", -0.5, false, "BDD-10"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.esperado, HorasEstimadasValidas(c.valor), "escenario BDD %s", c.escenario)
		})
	}
}
