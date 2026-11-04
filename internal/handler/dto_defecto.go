package handler

import (
	"time"

	"ingenieria-y-calidad/internal/domain"
)

// altaDefectoDTO es el cuerpo de POST /api/proyectos/:id/defectos. El estado no
// se acepta del cliente: el service siempre da de alta el defecto en "open".
type altaDefectoDTO struct {
	Descripcion       string                  `json:"descripcion"`
	Severidad         domain.SeveridadDefecto `json:"severidad"`
	HistoriaID        uint                    `json:"historia_id"`
	SprintDeteccionID uint                    `json:"sprint_deteccion_id"`
	ReportadoPor      *uint                   `json:"reportado_por"`
}

// actualizacionDefectoDTO es el cuerpo de PATCH
// /api/proyectos/:id/defectos/:defectoId. Cada campo es un puntero para
// distinguir "no informado" de "informado con valor".
type actualizacionDefectoDTO struct {
	Estado             *domain.EstadoDefecto `json:"estado"`
	SprintResolucionID *uint                 `json:"sprint_resolucion_id"`
}

// respuestaDefecto es la representación de un defecto. Por la decisión D-05 no
// se emite proyecto_id: va implícito en la ruta del proyecto.
type respuestaDefecto struct {
	ID                 uint                    `json:"id"`
	HistoriaID         uint                    `json:"historia_id"`
	SprintDeteccionID  uint                    `json:"sprint_deteccion_id"`
	SprintResolucionID *uint                   `json:"sprint_resolucion_id"`
	ReportadoPor       *uint                   `json:"reportado_por"`
	Descripcion        string                  `json:"descripcion"`
	Severidad          domain.SeveridadDefecto `json:"severidad"`
	Estado             domain.EstadoDefecto    `json:"estado"`
	CreadoEn           string                  `json:"creado_en"`
	ResueltoEn         *string                 `json:"resuelto_en"`
}

type respuestaListadoDefectos struct {
	Total    int                `json:"total"`
	Defectos []respuestaDefecto `json:"defectos"`
}

func construirRespuestaDefecto(d *domain.Defecto) respuestaDefecto {
	return respuestaDefecto{
		ID:                 d.ID,
		HistoriaID:         d.HistoriaID,
		SprintDeteccionID:  d.SprintDeteccionID,
		SprintResolucionID: d.SprintResolucionID,
		ReportadoPor:       d.ReportadoPor,
		Descripcion:        d.Descripcion,
		Severidad:          d.Severidad,
		Estado:             d.Estado,
		CreadoEn:           formatearTimestamp(d.CreadoEn),
		ResueltoEn:         formatearTimestampOpcional(d.ResueltoEn),
	}
}

func formatearTimestampOpcional(t *time.Time) *string {
	if t == nil {
		return nil
	}
	formateado := formatearTimestamp(*t)
	return &formateado
}
