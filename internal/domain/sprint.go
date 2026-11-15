package domain

import "time"

// EstadoSprint representa el estado de un Sprint dentro de su ciclo de vida.
// Solo admite los valores declarados en las constantes siguientes.
type EstadoSprint string

const (
	EstadoSprintPlanificado EstadoSprint = "Planificado"
	EstadoSprintActivo      EstadoSprint = "Activo"
	EstadoSprintCerrado     EstadoSprint = "Cerrado"
)

// Valido indica si el estado corresponde a un valor aceptado por el dominio.
func (e EstadoSprint) Valido() bool {
	return e == EstadoSprintPlanificado || e == EstadoSprintActivo || e == EstadoSprintCerrado
}

// TransicionValidaA implementa la maquina de estados Planificado -> Activo ->
// Cerrado. Cerrado es un estado terminal y no se admite la transicion al
// mismo estado.
func (e EstadoSprint) TransicionValidaA(destino EstadoSprint) bool {
	switch e {
	case EstadoSprintPlanificado:
		return destino == EstadoSprintActivo
	case EstadoSprintActivo:
		return destino == EstadoSprintCerrado
	default:
		return false
	}
}

// Sprint es un ciclo de trabajo planificado dentro de un proyecto, con un
// Sprint Goal y un rango de fechas. Su identificador es unico dentro del
// proyecto y nace en estado Planificado.
type Sprint struct {
	ID            uint
	ProyectoID    uint
	Identificador string
	SprintGoal    string
	Estado        EstadoSprint
	FechaInicio   time.Time
	FechaFin      time.Time
	CreadoEn      time.Time
	ActualizadoEn time.Time
}

// NuevoSprint describe el alta de un Sprint dentro de un proyecto. El estado
// no se acepta del cliente: todo Sprint nace en Planificado.
type NuevoSprint struct {
	Identificador string
	SprintGoal    string
	FechaInicio   time.Time
	FechaFin      time.Time
}
