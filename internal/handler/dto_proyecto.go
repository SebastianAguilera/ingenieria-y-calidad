package handler

import (
	"time"

	"ingenieria-y-calidad/internal/domain"
)

const formatoFecha = time.DateOnly

type integracionDTO struct {
	Nombre string `json:"nombre"`
	Email  string `json:"email"`
}

type altaProyectoDTO struct {
	Nombre      string           `json:"nombre"`
	FechaInicio string           `json:"fecha_inicio"`
	FechaFin    *string          `json:"fecha_fin"`
	Integrantes []integracionDTO `json:"integrantes"`
}

type actualizacionProyectoDTO struct {
	Nombre      string  `json:"nombre"`
	FechaInicio string  `json:"fecha_inicio"`
	FechaFin    *string `json:"fecha_fin"`
}

type cambioEstadoDTO struct {
	Estado domain.EstadoProyecto `json:"estado"`
}

type respuestaIntegrante struct {
	ID uint `json:"id"`
	// ProyectoID solo se informa cuando la respuesta corresponde al alta de un
	// vinculo. Anidado dentro de un proyecto seria redundante y cero.
	Nombre     string `json:"nombre"`
	Email      string `json:"email"`
	ProyectoID uint   `json:"proyecto_id,omitempty"`
	CreadoEn   string `json:"creado_en,omitempty"`
}

type respuestaProyecto struct {
	ID                  uint                  `json:"id"`
	Nombre              string                `json:"nombre"`
	FechaInicio         string                `json:"fecha_inicio"`
	FechaFin            *string               `json:"fecha_fin"`
	Estado              domain.EstadoProyecto `json:"estado"`
	CantidadIntegrantes int                   `json:"cantidad_integrantes"`
	Integrantes         []respuestaIntegrante `json:"integrantes"`
	CreadoEn            string                `json:"creado_en"`
	ActualizadoEn       string                `json:"actualizado_en"`
}

type respuestaListadoProyectos struct {
	Total     int                 `json:"total"`
	Proyectos []respuestaProyecto `json:"proyectos"`
}

type altaIntegranteDTO struct {
	Nombre string `json:"nombre"`
	Email  string `json:"email"`
}

func formatearFecha(t time.Time) string {
	return t.Format(formatoFecha)
}

func formatearFechaOpcional(t *time.Time) *string {
	if t == nil {
		return nil
	}
	formateada := formatearFecha(*t)
	return &formateada
}

func formatearTimestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func construirRespuestaProyecto(p *domain.Proyecto, conIntegrantes bool) respuestaProyecto {
	respuesta := respuestaProyecto{
		ID:                  p.ID,
		Nombre:              p.Nombre,
		FechaInicio:         formatearFecha(p.FechaInicio),
		FechaFin:            formatearFechaOpcional(p.FechaFin),
		Estado:              p.Estado,
		CantidadIntegrantes: len(p.Integrantes),
		CreadoEn:            formatearTimestamp(p.CreadoEn),
		ActualizadoEn:       formatearTimestamp(p.ActualizadoEn),
		Integrantes:         []respuestaIntegrante{},
	}

	if conIntegrantes {
		respuesta.Integrantes = construirRespuestaIntegrantes(p.Integrantes)
	}

	return respuesta
}

func construirRespuestaIntegrantes(integrantes []domain.Integrante) []respuestaIntegrante {
	respuestas := make([]respuestaIntegrante, 0, len(integrantes))
	for _, integrante := range integrantes {
		respuestas = append(respuestas, respuestaIntegrante{
			ID:       integrante.ID,
			Nombre:   integrante.Nombre,
			Email:    integrante.Email,
			CreadoEn: formatearTimestamp(integrante.CreadoEn),
		})
	}
	return respuestas
}

func construirRespuestaVinculo(i *domain.Integrante, proyectoID uint) respuestaIntegrante {
	return respuestaIntegrante{
		ID:         i.ID,
		Nombre:     i.Nombre,
		Email:      i.Email,
		ProyectoID: proyectoID,
		CreadoEn:   formatearTimestamp(i.CreadoEn),
	}
}
