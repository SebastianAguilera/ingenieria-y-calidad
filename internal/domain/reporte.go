package domain

import "time"

type ReporteProyecto struct {
	Base        string
	ProyectoID  uint
	Nombre      string
	Descripcion string
	FechaInicio time.Time
	FechaFin    *time.Time
	Estado      EstadoProyecto
	Historias   []Historia
	Metricas    MetricasProyecto
	Defectos    []Defecto
}

type ReporteSprint struct {
	Base          string
	ProyectoID    uint
	SprintID      uint
	Identificador string
	SprintGoal    string
	FechaInicio   time.Time
	FechaFin      time.Time
	Estado        EstadoSprint
	Historias     []Historia
	Metricas      MetricasSprint
	Defectos      []Defecto
}
