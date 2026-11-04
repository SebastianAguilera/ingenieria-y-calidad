package domain

import "time"

// SeveridadDefecto representa la gravedad de un defecto. Solo admite los
// valores declarados en las constantes siguientes.
type SeveridadDefecto string

const (
	SeveridadMinor    SeveridadDefecto = "minor"
	SeveridadMajor    SeveridadDefecto = "major"
	SeveridadCritical SeveridadDefecto = "critical"
	SeveridadBlocker  SeveridadDefecto = "blocker"
)

// Valida indica si la severidad corresponde a un valor aceptado por el dominio.
func (s SeveridadDefecto) Valida() bool {
	switch s {
	case SeveridadMinor, SeveridadMajor, SeveridadCritical, SeveridadBlocker:
		return true
	default:
		return false
	}
}

// EstadoDefecto representa el estado de un defecto dentro de su ciclo de vida.
// Solo admite los valores declarados en las constantes siguientes.
type EstadoDefecto string

const (
	EstadoDefectoAbierto    EstadoDefecto = "open"
	EstadoDefectoEnProgreso EstadoDefecto = "in_progress"
	EstadoDefectoResuelto   EstadoDefecto = "resolved"
	EstadoDefectoCerrado    EstadoDefecto = "closed"
	EstadoDefectoRechazado  EstadoDefecto = "rejected"
)

// Valido indica si el estado corresponde a un valor aceptado por el dominio.
func (e EstadoDefecto) Valido() bool {
	switch e {
	case EstadoDefectoAbierto, EstadoDefectoEnProgreso, EstadoDefectoResuelto,
		EstadoDefectoCerrado, EstadoDefectoRechazado:
		return true
	default:
		return false
	}
}

// EsDeResolucion indica si el estado exige un Sprint de resolución.
func (e EstadoDefecto) EsDeResolucion() bool {
	return e == EstadoDefectoResuelto || e == EstadoDefectoCerrado
}

// Defecto es un bug detectado en una historia de usuario durante un Sprint.
// SprintResolucionID y ResueltoEn solo se informan cuando el defecto se
// resuelve o se cierra.
type Defecto struct {
	ID                 uint
	ProyectoID         uint
	HistoriaID         uint
	SprintDeteccionID  uint
	SprintResolucionID *uint
	ReportadoPor       *uint
	Descripcion        string
	Severidad          SeveridadDefecto
	Estado             EstadoDefecto
	CreadoEn           time.Time
	ResueltoEn         *time.Time
}

// NuevoDefecto describe el alta de un defecto. El estado no se acepta del
// cliente: todo defecto nace en estado abierto.
type NuevoDefecto struct {
	Descripcion       string
	Severidad         SeveridadDefecto
	HistoriaID        uint
	SprintDeteccionID uint
	ReportadoPor      *uint
}

// ActualizacionDefecto describe la modificación parcial de un defecto. Cada
// campo nil significa "no informado".
type ActualizacionDefecto struct {
	Estado             *EstadoDefecto
	SprintResolucionID *uint
}
