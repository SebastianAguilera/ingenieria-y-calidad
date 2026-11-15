package domain

import "strings"

type AcceptanceCriterion struct {
	ID          uint
	HistoriaID  uint
	Descripcion string
	Position    int
	IsMet       bool
}

type NuevoAcceptanceCriterion struct {
	Descripcion string
}

type ActualizacionAcceptanceCriterion struct {
	Descripcion *string
	IsMet       *bool
}

func ValidarDescripcionCriterio(desc string) error {
	recortado := strings.TrimSpace(desc)
	if len(recortado) == 0 {
		return ErrDescripcionCriterioObligatoria
	}
	if len([]rune(recortado)) > 500 {
		return ErrDescripcionCriterioDemasiadoLarga
	}
	return nil
}
