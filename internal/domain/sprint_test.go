package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEstadoSprintValido(t *testing.T) {
	t.Parallel()

	casos := []struct {
		estado    EstadoSprint
		esValido  bool
		escenario string
	}{
		{EstadoSprintPlanificado, true, "E-01"},
		{EstadoSprintActivo, true, "E-29"},
		{EstadoSprintCerrado, true, "E-30"},
		{EstadoSprint("Terminado"), false, "E-32"},
		{EstadoSprint("planificado"), false, "E-32"},
		{EstadoSprint("activo"), false, "E-32"},
		{EstadoSprint("Cerrada"), false, "E-32"},
		{EstadoSprint(""), false, "E-32"},
	}

	for _, c := range casos {
		t.Run(string(c.estado), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.esValido, c.estado.Valido(), "escenario BDD %s", c.escenario)
		})
	}
}

func TestEstadoSprintTransicionValidaA(t *testing.T) {
	t.Parallel()

	// Matriz completa de la maquina de estados RB-05: solo Planificado ->
	// Activo y Activo -> Cerrado son validas. Cerrado es terminal y no se
	// admite la transicion al mismo estado.
	casos := []struct {
		nombre    string
		origen    EstadoSprint
		destino   EstadoSprint
		esValida  bool
		escenario string
	}{
		{"Planificado a Activo es valido", EstadoSprintPlanificado, EstadoSprintActivo, true, "E-29"},
		{"Activo a Cerrado es valido", EstadoSprintActivo, EstadoSprintCerrado, true, "E-30"},
		{"Activo a Planificado es invalido", EstadoSprintActivo, EstadoSprintPlanificado, false, "E-31"},
		{"Planificado a Cerrado es invalido", EstadoSprintPlanificado, EstadoSprintCerrado, false, "E-31"},
		{"Cerrado a Activo es invalido", EstadoSprintCerrado, EstadoSprintActivo, false, "E-31"},
		{"Cerrado a Planificado es invalido", EstadoSprintCerrado, EstadoSprintPlanificado, false, "E-31"},
		{"Activo a Activo es invalido", EstadoSprintActivo, EstadoSprintActivo, false, "E-31"},
		{"Planificado a Planificado es invalido", EstadoSprintPlanificado, EstadoSprintPlanificado, false, "E-31"},
		{"Cerrado a Cerrado es invalido", EstadoSprintCerrado, EstadoSprintCerrado, false, "E-31"},
		{"origen invalido nunca valida", EstadoSprint("otro"), EstadoSprintActivo, false, "E-32"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.esValida, c.origen.TransicionValidaA(c.destino),
				"transicion %s -> %s (escenario BDD %s)", c.origen, c.destino, c.escenario)
		})
	}
}
