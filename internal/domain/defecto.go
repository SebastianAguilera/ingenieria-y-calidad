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
//
// SprintDeteccionPresente indica que el body incluyó la clave
// sprint_deteccion_id, aunque su valor sea igual al actual o null. El service
// la rechaza siempre (US-18, RN-09).
type ActualizacionDefecto struct {
	Estado                  *EstadoDefecto
	SprintResolucionID      *uint
	Descripcion             *string
	Severidad               *SeveridadDefecto
	SprintDeteccionPresente bool
}

// TipoListadoDefecto indica por qué Sprint filtrar el listado de defectos de
// un Sprint (US-18). Es obligatorio y sensible a mayúsculas.
type TipoListadoDefecto string

const (
	TipoListadoDeteccion  TipoListadoDefecto = "deteccion"
	TipoListadoResolucion TipoListadoDefecto = "resolucion"
)

// Valido indica si el tipo de listado corresponde a un valor aceptado.
func (t TipoListadoDefecto) Valido() bool {
	return t == TipoListadoDeteccion || t == TipoListadoResolucion
}
