package domain

import "testing"

func TestEstadoSesionPokerValido(t *testing.T) {
	casos := []struct {
		nombre string
		estado EstadoSesionPoker
		valido bool
	}{
		{"votando", EstadoSesionVotando, true},
		{"revelada", EstadoSesionRevelada, true},
		{"cerrada", EstadoSesionCerrada, true},
		{"cancelada", EstadoSesionCancelada, true},
		{"vacio", "", false},
		{"desconocido", "Pausada", false},
		{"minusculas", "votando", false},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if obtenido := caso.estado.Valido(); obtenido != caso.valido {
				t.Fatalf("Valido() = %v, se esperaba %v", obtenido, caso.valido)
			}
		})
	}
}

func TestEstadoSesionPokerActiva(t *testing.T) {
	casos := []struct {
		estado EstadoSesionPoker
		activa bool
		nombre string
	}{
		{EstadoSesionVotando, true, "votando"},
		{EstadoSesionRevelada, true, "revelada"},
		{EstadoSesionCerrada, false, "cerrada"},
		{EstadoSesionCancelada, false, "cancelada"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if obtenido := caso.estado.Activa(); obtenido != caso.activa {
				t.Fatalf("Activa() = %v, se esperaba %v", obtenido, caso.activa)
			}
		})
	}
}
