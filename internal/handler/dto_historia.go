package handler

import (
	"bytes"
	"encoding/json"

	"ingenieria-y-calidad/internal/domain"
)

// altaHistoriaDTO es el cuerpo de POST /api/proyectos/:id/historias. La
// prioridad y el estado son opcionales: si llegan vacios el servicio aplica los
// valores por defecto. story_points se recibe crudo para poder responder 422 (y
// no 400) cuando el valor no es numerico.
type altaHistoriaDTO struct {
	Titulo      string                   `json:"titulo"`
	Descripcion string                   `json:"descripcion"`
	Prioridad   domain.PrioridadHistoria `json:"prioridad"`
	Estado      domain.EstadoHistoria    `json:"estado"`
	StoryPoints json.RawMessage          `json:"story_points"`
}

// actualizacionHistoriaDTO es el cuerpo de PATCH
// /api/proyectos/:id/historias/:historiaId. Cada campo es un puntero para
// distinguir "no informado" de "informado con valor".
type actualizacionHistoriaDTO struct {
	Titulo      *string                   `json:"titulo"`
	Descripcion *string                   `json:"descripcion"`
	Prioridad   *domain.PrioridadHistoria `json:"prioridad"`
	Estado      *domain.EstadoHistoria    `json:"estado"`
	StoryPoints json.RawMessage           `json:"story_points"`
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
	StoryPoints   *float64                 `json:"story_points"`
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
		StoryPoints:   h.StoryPoints,
		CreadoEn:      formatearTimestamp(h.CreadoEn),
		ActualizadoEn: formatearTimestamp(h.ActualizadoEn),
	}
}

// parsearStoryPoints convierte el valor crudo de story_points. Un campo ausente
// o null devuelve nil ("no informado"); cualquier valor que no sea un numero
// JSON devuelve domain.ErrStoryPointsInvalido. La pertenencia a la escala la
// valida el service.
func parsearStoryPoints(crudo json.RawMessage) (*float64, error) {
	crudo = bytes.TrimSpace(crudo)
	if len(crudo) == 0 || bytes.Equal(crudo, []byte("null")) {
		return nil, nil
	}

	var valor float64
	if err := json.Unmarshal(crudo, &valor); err != nil {
		return nil, domain.ErrStoryPointsInvalido
	}
	return &valor, nil
}
