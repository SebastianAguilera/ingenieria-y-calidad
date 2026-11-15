package handler

import (
	"fmt"

	"ingenieria-y-calidad/internal/domain"
)

const baseReporte = "estado_actual"

type respuestaReporteHistoriaItem struct {
	ID             uint                  `json:"id"`
	Codigo         string                `json:"codigo"`
	Titulo         string                `json:"titulo"`
	Estado         domain.EstadoHistoria `json:"estado"`
	StoryPoints    *float64              `json:"story_points"`
	HorasEstimadas *float64              `json:"horas_estimadas"`
	HorasReales    *float64              `json:"horas_reales"`
}

type respuestaReporteDefectoItem struct {
	ID                 uint                    `json:"id"`
	Codigo             string                  `json:"codigo"`
	Descripcion        string                  `json:"descripcion"`
	Severidad          domain.SeveridadDefecto `json:"severidad"`
	Estado             domain.EstadoDefecto    `json:"estado"`
	SprintDeteccionID  uint                    `json:"sprint_deteccion_id"`
	SprintResolucionID *uint                   `json:"sprint_resolucion_id"`
}

type respuestaReporteProyecto struct {
	ProyectoID  uint                           `json:"proyecto_id"`
	Nombre      string                         `json:"nombre"`
	Descripcion string                         `json:"descripcion"`
	FechaInicio string                         `json:"fecha_inicio"`
	FechaFin    *string                        `json:"fecha_fin"`
	Estado      domain.EstadoProyecto          `json:"estado"`
	Base        string                         `json:"base"`
	Historias   []respuestaReporteHistoriaItem `json:"historias"`
	Metricas    respuestaMetricasProyecto      `json:"metricas"`
	Defectos    []respuestaReporteDefectoItem  `json:"defectos"`
}

type respuestaReporteSprint struct {
	ProyectoID    uint                           `json:"proyecto_id"`
	SprintID      uint                           `json:"sprint_id"`
	Identificador string                         `json:"identificador"`
	SprintGoal    string                         `json:"sprint_goal"`
	FechaInicio   string                         `json:"fecha_inicio"`
	FechaFin      string                         `json:"fecha_fin"`
	Estado        domain.EstadoSprint            `json:"estado"`
	Base          string                         `json:"base"`
	Historias     []respuestaReporteHistoriaItem `json:"historias"`
	Metricas      respuestaMetricasSprint        `json:"metricas"`
	Defectos      []respuestaReporteDefectoItem  `json:"defectos"`
}

func construirRespuestaReporteProyecto(rp domain.ReporteProyecto) respuestaReporteProyecto {
	historias := make([]respuestaReporteHistoriaItem, len(rp.Historias))
	for i, h := range rp.Historias {
		historias[i] = respuestaReporteHistoriaItem{
			ID:             h.ID,
			Codigo:         h.Codigo,
			Titulo:         h.Titulo,
			Estado:         h.Estado,
			StoryPoints:    h.StoryPoints,
			HorasEstimadas: h.HorasEstimadas,
			HorasReales:    h.EsfuerzoReal,
		}
	}
	if historias == nil {
		historias = []respuestaReporteHistoriaItem{}
	}

	defectos := make([]respuestaReporteDefectoItem, len(rp.Defectos))
	for i, d := range rp.Defectos {
		defectos[i] = respuestaReporteDefectoItem{
			ID:                 d.ID,
			Codigo:             fmt.Sprintf("DEF-%d", d.ID),
			Descripcion:        d.Descripcion,
			Severidad:          d.Severidad,
			Estado:             d.Estado,
			SprintDeteccionID:  d.SprintDeteccionID,
			SprintResolucionID: d.SprintResolucionID,
		}
	}
	if defectos == nil {
		defectos = []respuestaReporteDefectoItem{}
	}

	return respuestaReporteProyecto{
		ProyectoID:  rp.ProyectoID,
		Nombre:      rp.Nombre,
		Descripcion: rp.Descripcion,
		FechaInicio: formatearFecha(rp.FechaInicio),
		FechaFin:    formatearFechaOpcional(rp.FechaFin),
		Estado:      rp.Estado,
		Base:        rp.Base,
		Historias:   historias,
		Metricas:    construirRespuestaMetricasProyecto(rp.ProyectoID, rp.Metricas),
		Defectos:    defectos,
	}
}

func construirRespuestaReporteSprint(rs domain.ReporteSprint) respuestaReporteSprint {
	historias := make([]respuestaReporteHistoriaItem, len(rs.Historias))
	for i, h := range rs.Historias {
		historias[i] = respuestaReporteHistoriaItem{
			ID:             h.ID,
			Codigo:         h.Codigo,
			Titulo:         h.Titulo,
			Estado:         h.Estado,
			StoryPoints:    h.StoryPoints,
			HorasEstimadas: h.HorasEstimadas,
			HorasReales:    h.EsfuerzoReal,
		}
	}
	if historias == nil {
		historias = []respuestaReporteHistoriaItem{}
	}

	defectos := make([]respuestaReporteDefectoItem, len(rs.Defectos))
	for i, d := range rs.Defectos {
		defectos[i] = respuestaReporteDefectoItem{
			ID:                 d.ID,
			Codigo:             fmt.Sprintf("DEF-%d", d.ID),
			Descripcion:        d.Descripcion,
			Severidad:          d.Severidad,
			Estado:             d.Estado,
			SprintDeteccionID:  d.SprintDeteccionID,
			SprintResolucionID: d.SprintResolucionID,
		}
	}
	if defectos == nil {
		defectos = []respuestaReporteDefectoItem{}
	}

	return respuestaReporteSprint{
		ProyectoID:    rs.ProyectoID,
		SprintID:      rs.SprintID,
		Identificador: rs.Identificador,
		SprintGoal:    rs.SprintGoal,
		FechaInicio:   formatearFecha(rs.FechaInicio),
		FechaFin:      formatearFecha(rs.FechaFin),
		Estado:        rs.Estado,
		Base:          rs.Base,
		Historias:     historias,
		Metricas:      construirRespuestaMetricasSprint(rs.SprintID, rs.Metricas),
		Defectos:      defectos,
	}
}
