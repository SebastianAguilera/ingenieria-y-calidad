package handler

import (
	"bytes"
	"encoding/json"

	"ingenieria-y-calidad/internal/domain"
)

type altaAcceptanceCriterionDTO struct {
	Descripcion string `json:"descripcion"`
}

type actualizacionAcceptanceCriterionDTO struct {
	Descripcion *string         `json:"descripcion"`
	IsMet       json.RawMessage `json:"is_met"`
}

type respuestaAcceptanceCriterion struct {
	ID          uint   `json:"id"`
	HistoriaID  uint   `json:"historia_id"`
	Descripcion string `json:"descripcion"`
	Position    int    `json:"position"`
	IsMet       bool   `json:"is_met"`
}

type respuestaListadoAcceptanceCriteria struct {
	Total     int                            `json:"total"`
	Criterios []respuestaAcceptanceCriterion `json:"criterios"`
}

func construirRespuestaAcceptanceCriterion(c *domain.AcceptanceCriterion) respuestaAcceptanceCriterion {
	return respuestaAcceptanceCriterion{
		ID:          c.ID,
		HistoriaID:  c.HistoriaID,
		Descripcion: c.Descripcion,
		Position:    c.Position,
		IsMet:       c.IsMet,
	}
}

func parsearIsMet(crudo json.RawMessage) (*bool, error) {
	crudo = bytes.TrimSpace(crudo)
	if len(crudo) == 0 || bytes.Equal(crudo, []byte("null")) {
		return nil, nil
	}
	var valor bool
	if err := json.Unmarshal(crudo, &valor); err != nil {
		return nil, domain.ErrActualizacionCriterioVacia // o validación
	}
	return &valor, nil
}
