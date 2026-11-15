package handler

import "ingenieria-y-calidad/internal/domain"

// baseMetricas indica que las métricas se calculan con el estado actual de los
// datos: no hay snapshots históricos (US-20, RN-19).
const baseMetricas = "estado_actual"

// respuestaMetricasSprint es el contrato JSON de
// GET /api/proyectos/:id/sprints/:sprintId/metricas.
type respuestaMetricasSprint struct {
	SprintID                uint     `json:"sprint_id"`
	Base                    string   `json:"base"`
	StoryPointsPlanificados float64  `json:"story_points_planificados"`
	StoryPointsCompletados  float64  `json:"story_points_completados"`
	PorcentajeCompletado    *float64 `json:"porcentaje_completado"`
	HistoriasCompletadas    int      `json:"historias_completadas"`
	HorasEstimadas          float64  `json:"horas_estimadas"`
	HorasReales             float64  `json:"horas_reales"`
	DesviacionEsfuerzo      *float64 `json:"desviacion_esfuerzo"`
	DefectosDetectados      int      `json:"defectos_detectados"`
	DefectosResueltos       int      `json:"defectos_resueltos"`
}

// respuestaMetricasProyecto es el contrato JSON de
// GET /api/proyectos/:id/metricas.
type respuestaMetricasProyecto struct {
	ProyectoID                     uint     `json:"proyecto_id"`
	Base                           string   `json:"base"`
	Velocidad                      *float64 `json:"velocidad"`
	PorcentajeHistoriasCompletadas *float64 `json:"porcentaje_historias_completadas"`
	HorasEstimadas                 float64  `json:"horas_estimadas"`
	HorasReales                    float64  `json:"horas_reales"`
	DesviacionEsfuerzo             *float64 `json:"desviacion_esfuerzo"`
	DefectosDetectados             int      `json:"defectos_detectados"`
	DefectosResueltos              int      `json:"defectos_resueltos"`
	RatioDefectos                  *float64 `json:"ratio_defectos"`
}

func construirRespuestaMetricasSprint(sprintID uint, m domain.MetricasSprint) respuestaMetricasSprint {
	return respuestaMetricasSprint{
		SprintID:                sprintID,
		Base:                    baseMetricas,
		StoryPointsPlanificados: m.StoryPointsPlanificados,
		StoryPointsCompletados:  m.StoryPointsCompletados,
		PorcentajeCompletado:    m.PorcentajeCompletado,
		HistoriasCompletadas:    m.HistoriasCompletadas,
		HorasEstimadas:          m.HorasEstimadas,
		HorasReales:             m.HorasReales,
		DesviacionEsfuerzo:      m.DesviacionEsfuerzo,
		DefectosDetectados:      m.DefectosDetectados,
		DefectosResueltos:       m.DefectosResueltos,
	}
}

func construirRespuestaMetricasProyecto(proyectoID uint, m domain.MetricasProyecto) respuestaMetricasProyecto {
	return respuestaMetricasProyecto{
		ProyectoID:                     proyectoID,
		Base:                           baseMetricas,
		Velocidad:                      m.Velocidad,
		PorcentajeHistoriasCompletadas: m.PorcentajeHistoriasCompletadas,
		HorasEstimadas:                 m.HorasEstimadas,
		HorasReales:                    m.HorasReales,
		DesviacionEsfuerzo:             m.DesviacionEsfuerzo,
		DefectosDetectados:             m.DefectosDetectados,
		DefectosResueltos:              m.DefectosResueltos,
		RatioDefectos:                  m.RatioDefectos,
	}
}
