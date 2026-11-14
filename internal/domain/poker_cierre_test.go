package domain

import (
	"errors"
	"testing"
	"time"
)

// Tests de US-17: reglas de transición para acordar y cancelar una sesión de
// Planning Poker y construcción del story_estimate.

func TestEstadoSesionPokerFinalizada(t *testing.T) {
	casos := []struct {
		nombre     string
		estado     EstadoSesionPoker
		finalizada bool
	}{
		{"votando", EstadoSesionVotando, false},
		{"revelada", EstadoSesionRevelada, false},
		{"cerrada", EstadoSesionCerrada, true},
		{"cancelada", EstadoSesionCancelada, true},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if obtenido := caso.estado.Finalizada(); obtenido != caso.finalizada {
				t.Fatalf("Finalizada() = %v, se esperaba %v", obtenido, caso.finalizada)
			}
		})
	}
}

// TestValidarAcuerdoSesionPoker cubre RB-01: solo se acuerda desde Revelada.
func TestValidarAcuerdoSesionPoker(t *testing.T) {
	casos := []struct {
		nombre   string
		estado   EstadoSesionPoker
		esperado error
	}{
		{"E-01 revelada permite acordar", EstadoSesionRevelada, nil},
		{"E-04 votando no está revelada", EstadoSesionVotando, ErrSesionPokerNoRevelada},
		{"E-05 cerrada ya finalizó", EstadoSesionCerrada, ErrSesionPokerCerrada},
		{"E-05 cancelada ya finalizó", EstadoSesionCancelada, ErrSesionPokerCerrada},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			err := ValidarAcuerdoSesionPoker(caso.estado)
			if !errors.Is(err, caso.esperado) || (caso.esperado == nil && err != nil) {
				t.Fatalf("ValidarAcuerdoSesionPoker(%q) = %v, se esperaba %v", caso.estado, err, caso.esperado)
			}
		})
	}
}

// TestValidarCancelacionSesionPoker cubre RB-08: solo se cancela una sesión
// activa (Votando o Revelada).
func TestValidarCancelacionSesionPoker(t *testing.T) {
	casos := []struct {
		nombre   string
		estado   EstadoSesionPoker
		esperado error
	}{
		{"E-11 votando permite cancelar", EstadoSesionVotando, nil},
		{"E-11 revelada permite cancelar", EstadoSesionRevelada, nil},
		{"E-12 cerrada ya finalizó", EstadoSesionCerrada, ErrSesionPokerCerrada},
		{"E-12 cancelada ya finalizó", EstadoSesionCancelada, ErrSesionPokerCerrada},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			err := ValidarCancelacionSesionPoker(caso.estado)
			if !errors.Is(err, caso.esperado) || (caso.esperado == nil && err != nil) {
				t.Fatalf("ValidarCancelacionSesionPoker(%q) = %v, se esperaba %v", caso.estado, err, caso.esperado)
			}
		})
	}
}

// TestNuevaEstimacionPlanningPoker cubre RB-05: el story_estimate lleva la
// historia, la sesión, la ronda, los puntos, la fuente y quién acordó.
func TestNuevaEstimacionPlanningPoker(t *testing.T) {
	momento := time.Date(2026, 11, 14, 18, 30, 0, 0, time.UTC)
	sesion := SesionPoker{ID: 7, HistoriaID: 5}

	estimacion := NuevaEstimacionPlanningPoker(sesion, 14, 2, 0.5, momento)

	if estimacion.HistoriaID != 5 {
		t.Errorf("HistoriaID = %d, se esperaba 5", estimacion.HistoriaID)
	}
	if estimacion.SesionID == nil || *estimacion.SesionID != 7 {
		t.Errorf("SesionID = %v, se esperaba 7", estimacion.SesionID)
	}
	if estimacion.RondaID == nil || *estimacion.RondaID != 14 {
		t.Errorf("RondaID = %v, se esperaba 14", estimacion.RondaID)
	}
	if estimacion.EstimadoPor == nil || *estimacion.EstimadoPor != 2 {
		t.Errorf("EstimadoPor = %v, se esperaba 2", estimacion.EstimadoPor)
	}
	if estimacion.Puntos != 0.5 {
		t.Errorf("Puntos = %v, se esperaba 0.5", estimacion.Puntos)
	}
	if estimacion.Fuente != FuenteEstimacionPlanningPoker || FuenteEstimacionPlanningPoker != "planning_poker" {
		t.Errorf("Fuente = %q, se esperaba planning_poker", estimacion.Fuente)
	}
	if !estimacion.EstimadoEn.Equal(momento) {
		t.Errorf("EstimadoEn = %v, se esperaba %v", estimacion.EstimadoEn, momento)
	}
}
