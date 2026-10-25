package domain

import "time"

// SprintBacklogItem representa la asociación entre una historia de usuario
// y un Sprint, con su orden dentro del Sprint Backlog y la fecha de asignación.
type SprintBacklogItem struct {
	ID              uint
	ProyectoID      uint
	SprintID        uint
	HistoriaID      uint
	Orden           int
	FechaAsignacion time.Time
	CreadoEn        time.Time
	ActualizadoEn   time.Time
}
