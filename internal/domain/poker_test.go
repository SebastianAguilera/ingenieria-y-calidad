package domain

import (
	"testing"
	"time"
)

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

func TestSesionPokerExponeCancelacionYEstimacion(t *testing.T) {
	canceladoEn := time.Date(2026, 11, 13, 18, 0, 0, 0, time.UTC)
	estimacion := &StoryEstimate{
		ID:          12,
		HistoriaID:  5,
		SesionID:    9,
		RondaID:     3,
		Puntos:      5,
		Fuente:      "planning_poker",
		EstimadoPor: punteroUint(2),
		EstimadoEn:  canceladoEn,
	}
	sesion := SesionPoker{
		ID:            9,
		Estado:        EstadoSesionCancelada,
		CanceladoEn:   &canceladoEn,
		StoryEstimate: estimacion,
	}

	if sesion.CanceladoEn == nil || !sesion.CanceladoEn.Equal(canceladoEn) {
		t.Fatalf("CanceladoEn = %v, se esperaba %v", sesion.CanceladoEn, canceladoEn)
	}
	if sesion.StoryEstimate == nil {
		t.Fatal("StoryEstimate no debe ser nil")
	}
	if sesion.StoryEstimate.SesionID != sesion.ID {
		t.Fatalf("SesionID = %d, se esperaba %d", sesion.StoryEstimate.SesionID, sesion.ID)
	}
	if sesion.StoryEstimate.RondaID == 0 {
		t.Fatal("RondaID debe identificar la ronda acordada")
	}
}

func punteroUint(valor uint) *uint {
	return &valor
}
