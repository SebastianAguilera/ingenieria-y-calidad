package handler

import "ingenieria-y-calidad/internal/domain"

// altaHistoriaDTO es el cuerpo de POST /api/proyectos/:id/historias. La
// prioridad y el estado son opcionales: si llegan vacios el servicio aplica los
// valores por defecto.
type altaHistoriaDTO struct {
	Titulo      string                   `json:"titulo"`
	Descripcion string                   `json:"descripcion"`
	Prioridad   domain.PrioridadHistoria `json:"prioridad"`
	Estado      domain.EstadoHistoria    `json:"estado"`
}

// actualizacionHistoriaDTO es el cuerpo de PATCH
// /api/proyectos/:id/historias/:historiaId. Cada campo es un puntero para
// distinguir "no informado" de "informado con valor".
type actualizacionHistoriaDTO struct {
	Titulo      *string                   `json:"titulo"`
	Descripcion *string                   `json:"descripcion"`
	Prioridad   *domain.PrioridadHistoria `json:"prioridad"`
	Estado      *domain.EstadoHistoria    `json:"estado"`
}

// respuestaHistoria es la representacion de una historia. Por la decision D-03
// no se emite proyecto_id: va implicito en la ruta del proyecto.
type respuestaHistoria struct {
	ID            uint                     `json:"id"`
	Codigo        string                   `json:"codigo"`
	Titulo        string                   `json:"titulo"`
	Descripcion   string                   `json:"descripcion"`
	Prioridad     domain.PrioridadHistoria `json:"prioridad"`
	Estado        domain.EstadoHistoria    `json:"estado"`
	CreadoEn      string                   `json:"creado_en"`
	ActualizadoEn string                   `json:"actualizado_en"`
}

type respuestaListadoHistorias struct {
	Total     int                 `json:"total"`
	Historias []respuestaHistoria `json:"historias"`
}

func construirRespuestaHistoria(h *domain.Historia) respuestaHistoria {
	return respuestaHistoria{
		ID:            h.ID,
		Codigo:        h.Codigo,
		Titulo:        h.Titulo,
		Descripcion:   h.Descripcion,
		Prioridad:     h.Prioridad,
		Estado:        h.Estado,
		CreadoEn:      formatearTimestamp(h.CreadoEn),
		ActualizadoEn: formatearTimestamp(h.ActualizadoEn),
	}
}
