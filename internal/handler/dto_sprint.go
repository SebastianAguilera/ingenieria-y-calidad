package handler

import "ingenieria-y-calidad/internal/domain"

// altaSprintDTO es el cuerpo de POST /api/proyectos/:id/sprints. Los cuatro
// campos son obligatorios y las fechas viajan como fecha calendario
// "YYYY-MM-DD". El estado no se acepta del cliente: el service siempre da de
// alta el Sprint en "Planificado" (decision D-03).
type altaSprintDTO struct {
	Identificador string `json:"identificador"`
	SprintGoal    string `json:"sprint_goal"`
	FechaInicio   string `json:"fecha_inicio"`
	FechaFin      string `json:"fecha_fin"`
}

// cambioEstadoSprintDTO es el cuerpo de PATCH
// /api/proyectos/:id/sprints/:sprintId/estado.
type cambioEstadoSprintDTO struct {
	Estado domain.EstadoSprint `json:"estado"`
}

// respuestaSprint es la representacion de un Sprint. Por la decision D-04 no
// se emite proyecto_id: va implicito en la ruta del proyecto.
type respuestaSprint struct {
	ID            uint                `json:"id"`
	Identificador string              `json:"identificador"`
	SprintGoal    string              `json:"sprint_goal"`
	Estado        domain.EstadoSprint `json:"estado"`
	FechaInicio   string              `json:"fecha_inicio"`
	FechaFin      string              `json:"fecha_fin"`
	CreadoEn      string              `json:"creado_en"`
	ActualizadoEn string              `json:"actualizado_en"`
}

type respuestaListadoSprints struct {
	Total   int               `json:"total"`
	Sprints []respuestaSprint `json:"sprints"`
}

func construirRespuestaSprint(s *domain.Sprint) respuestaSprint {
	return respuestaSprint{
		ID:            s.ID,
		Identificador: s.Identificador,
		SprintGoal:    s.SprintGoal,
		Estado:        s.Estado,
		FechaInicio:   formatearFecha(s.FechaInicio),
		FechaFin:      formatearFecha(s.FechaFin),
		CreadoEn:      formatearTimestamp(s.CreadoEn),
		ActualizadoEn: formatearTimestamp(s.ActualizadoEn),
	}
}
