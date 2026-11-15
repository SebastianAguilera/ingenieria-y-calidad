package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalcularMetricasSprint(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		escenario string
		historias []Historia
		defectos  []Defecto
		sprintID  uint
		esperado  MetricasSprint
	}{
		{
			nombre:    "E-01 planificados y completados segun estado",
			escenario: "E-01",
			historias: []Historia{
				{ID: 1, Estado: EstadoHistoriaHecho, StoryPoints: punteroFloat64(5)},
				{ID: 2, Estado: EstadoHistoriaEnProgreso, StoryPoints: punteroFloat64(3)},
				{ID: 3, Estado: EstadoHistoriaBacklog},
			},
			esperado: MetricasSprint{
				StoryPointsPlanificados: 8,
				StoryPointsCompletados:  5,
				PorcentajeCompletado:    punteroFloat64(0.625),
				HistoriasCompletadas:    1,
			},
		},
		{
			nombre:    "E-02 sprint sin historias",
			escenario: "E-02",
			esperado:  MetricasSprint{},
		},
		{
			nombre:    "E-03 horas y desviacion",
			escenario: "E-03",
			historias: []Historia{
				{ID: 1, Estado: EstadoHistoriaHecho, StoryPoints: punteroFloat64(5), HorasEstimadas: punteroFloat64(5), EsfuerzoReal: punteroFloat64(8)},
				{ID: 2, Estado: EstadoHistoriaHecho, StoryPoints: punteroFloat64(3), HorasEstimadas: punteroFloat64(5), EsfuerzoReal: punteroFloat64(4)},
				{ID: 3, Estado: EstadoHistoriaEnProgreso, StoryPoints: punteroFloat64(2), HorasEstimadas: punteroFloat64(3)},
			},
			esperado: MetricasSprint{
				StoryPointsPlanificados: 10,
				StoryPointsCompletados:  8,
				PorcentajeCompletado:    punteroFloat64(0.8),
				HistoriasCompletadas:    2,
				HorasEstimadas:          13,
				HorasReales:             12,
				DesviacionEsfuerzo:      punteroFloat64(0.2),
			},
		},
		{
			nombre:    "E-07 story points nil no rompen el calculo",
			escenario: "E-07",
			historias: []Historia{
				{ID: 1, Estado: EstadoHistoriaHecho},
				{ID: 2, Estado: EstadoHistoriaHecho, StoryPoints: punteroFloat64(5)},
			},
			esperado: MetricasSprint{
				StoryPointsPlanificados: 5,
				StoryPointsCompletados:  5,
				PorcentajeCompletado:    punteroFloat64(1),
				HistoriasCompletadas:    2,
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			obtenido := CalcularMetricasSprint(caso.historias, caso.defectos, caso.sprintID)
			assertMetricasSprint(t, caso.esperado, obtenido, caso.escenario)
		})
	}
}

func TestCalcularMetricasSprintDefectoDetectadoYResuelto(t *testing.T) {
	t.Parallel()

	defectos := []Defecto{
		{
			ID:                 1,
			SprintDeteccionID:  2,
			SprintResolucionID: punteroUint(3),
			Estado:             EstadoDefectoResuelto,
		},
	}

	detectado := CalcularMetricasSprint(nil, defectos, 2)
	assert.Equal(t, 1, detectado.DefectosDetectados, "escenario E-10: Sprint de deteccion")
	assert.Equal(t, 0, detectado.DefectosResueltos, "escenario E-10: Sprint de deteccion")

	resuelto := CalcularMetricasSprint(nil, defectos, 3)
	assert.Equal(t, 0, resuelto.DefectosDetectados, "escenario E-10: Sprint de resolucion")
	assert.Equal(t, 1, resuelto.DefectosResueltos, "escenario E-10: Sprint de resolucion")
}

func TestCalcularMetricasSprintCuentaRechazadosComoDetectados(t *testing.T) {
	t.Parallel()

	defectos := []Defecto{
		{ID: 1, SprintDeteccionID: 1, Estado: EstadoDefectoRechazado},
	}

	metricas := CalcularMetricasSprint(nil, defectos, 1)
	assert.Equal(t, 1, metricas.DefectosDetectados, "escenario E-08: rejected cuenta como detectado")
	assert.Equal(t, 0, metricas.DefectosResueltos)
}

func TestCalcularSPPlanificados(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		historias []Historia
		esperado  float64
	}{
		{"sin historias", nil, 0},
		{"story points nil", []Historia{{ID: 1}}, 0},
		{
			"suma todos los estados",
			[]Historia{
				{ID: 1, Estado: EstadoHistoriaBacklog, StoryPoints: punteroFloat64(2)},
				{ID: 2, Estado: EstadoHistoriaHecho, StoryPoints: punteroFloat64(3)},
			},
			5,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()
			assert.InDelta(t, caso.esperado, CalcularSPPlanificados(caso.historias), 1e-9)
		})
	}
}

func TestCalcularSPCompletados(t *testing.T) {
	t.Parallel()

	historias := []Historia{
		{ID: 1, Estado: EstadoHistoriaHecho, StoryPoints: punteroFloat64(5)},
		{ID: 2, Estado: EstadoHistoriaHecho},
		{ID: 3, Estado: EstadoHistoriaEnProgreso, StoryPoints: punteroFloat64(3)},
		{ID: 4, Estado: EstadoHistoriaBacklog, StoryPoints: punteroFloat64(2)},
	}

	assert.InDelta(t, 5, CalcularSPCompletados(historias), 1e-9)
	assert.InDelta(t, 0, CalcularSPCompletados(nil), 1e-9)
}

func TestCalcularPorcentajeCompletado(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre       string
		completados  float64
		planificados float64
		esperado     *float64
	}{
		{"sin planificados", 0, 0, nil},
		{"cero completados", 0, 8, punteroFloat64(0)},
		{"cinco de ocho", 5, 8, punteroFloat64(0.625)},
		{"un tercio", 1, 3, punteroFloat64(1.0 / 3.0)},
		{"todo completado", 5, 5, punteroFloat64(1)},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()
			assertFloatPtr(t, caso.esperado, CalcularPorcentajeCompletado(caso.completados, caso.planificados), "porcentaje completado")
		})
	}
}

func TestContarHistoriasCompletadas(t *testing.T) {
	t.Parallel()

	historias := []Historia{
		{ID: 1, Estado: EstadoHistoriaHecho},
		{ID: 2, Estado: EstadoHistoriaHecho},
		{ID: 3, Estado: EstadoHistoriaEnProgreso},
		{ID: 4, Estado: EstadoHistoriaBacklog},
	}

	assert.Equal(t, 2, ContarHistoriasCompletadas(historias))
	assert.Equal(t, 0, ContarHistoriasCompletadas(nil))
}

func TestSumarHorasEstimadas(t *testing.T) {
	t.Parallel()

	historias := []Historia{
		{ID: 1, HorasEstimadas: punteroFloat64(5)},
		{ID: 2},
		{ID: 3, HorasEstimadas: punteroFloat64(3)},
	}

	assert.InDelta(t, 8, SumarHorasEstimadas(historias), 1e-9)
	assert.InDelta(t, 0, SumarHorasEstimadas(nil), 1e-9)
}

func TestSumarHorasReales(t *testing.T) {
	t.Parallel()

	historias := []Historia{
		{ID: 1, EsfuerzoReal: punteroFloat64(8)},
		{ID: 2},
		{ID: 3, EsfuerzoReal: punteroFloat64(4)},
	}

	assert.InDelta(t, 12, SumarHorasReales(historias), 1e-9)
	assert.InDelta(t, 0, SumarHorasReales(nil), 1e-9)
}

func TestCalcularDesviacionEsfuerzo(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		historias []Historia
		esperado  *float64
	}{
		{"sin historias", nil, nil},
		{
			"sin pares completos",
			[]Historia{
				{ID: 1, HorasEstimadas: punteroFloat64(5)},
				{ID: 2, EsfuerzoReal: punteroFloat64(4)},
			},
			nil,
		},
		{
			"estimadas en cero",
			[]Historia{
				{ID: 1, HorasEstimadas: punteroFloat64(0), EsfuerzoReal: punteroFloat64(4)},
			},
			nil,
		},
		{
			"sin desviacion",
			[]Historia{
				{ID: 1, HorasEstimadas: punteroFloat64(5), EsfuerzoReal: punteroFloat64(5)},
			},
			punteroFloat64(0),
		},
		{
			"desviacion positiva",
			[]Historia{
				{ID: 1, HorasEstimadas: punteroFloat64(5), EsfuerzoReal: punteroFloat64(8)},
				{ID: 2, HorasEstimadas: punteroFloat64(5), EsfuerzoReal: punteroFloat64(4)},
			},
			punteroFloat64(0.2),
		},
		{
			"desviacion negativa",
			[]Historia{
				{ID: 1, HorasEstimadas: punteroFloat64(5), EsfuerzoReal: punteroFloat64(4)},
			},
			punteroFloat64(-0.2),
		},
		{
			"ignora pares incompletos",
			[]Historia{
				{ID: 1, HorasEstimadas: punteroFloat64(5), EsfuerzoReal: punteroFloat64(8)},
				{ID: 2, HorasEstimadas: punteroFloat64(5)},
				{ID: 3, EsfuerzoReal: punteroFloat64(4)},
			},
			punteroFloat64(0.6),
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()
			assertFloatPtr(t, caso.esperado, CalcularDesviacionEsfuerzo(caso.historias), "desviacion de esfuerzo")
		})
	}
}

func TestCalcularVelocidad(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre       string
		sprints      []Sprint
		asignaciones []SprintBacklogItem
		historias    []Historia
		esperado     *float64
	}{
		{
			nombre:       "E-05 sin cerrados",
			sprints:      []Sprint{{ID: 1, Estado: EstadoSprintActivo}},
			asignaciones: []SprintBacklogItem{{SprintID: 1, HistoriaID: 1}},
			historias:    []Historia{{ID: 1, Estado: EstadoHistoriaHecho, StoryPoints: punteroFloat64(20)}},
			esperado:     nil,
		},
		{
			nombre:   "E-05 un cerrado sin historias",
			sprints:  []Sprint{{ID: 1, Estado: EstadoSprintCerrado}},
			esperado: punteroFloat64(0),
		},
		{
			nombre:  "E-05 cerrados con SP 0",
			sprints: []Sprint{{ID: 1, Estado: EstadoSprintCerrado}, {ID: 2, Estado: EstadoSprintCerrado}},
			asignaciones: []SprintBacklogItem{
				{SprintID: 1, HistoriaID: 1},
				{SprintID: 2, HistoriaID: 2},
			},
			historias: []Historia{
				{ID: 1, Estado: EstadoHistoriaHecho, StoryPoints: punteroFloat64(0)},
				{ID: 2, Estado: EstadoHistoriaHecho},
			},
			esperado: punteroFloat64(0),
		},
		{
			nombre:  "E-05 activo excluido",
			sprints: []Sprint{{ID: 1, Estado: EstadoSprintCerrado}, {ID: 2, Estado: EstadoSprintActivo}},
			asignaciones: []SprintBacklogItem{
				{SprintID: 1, HistoriaID: 1},
				{SprintID: 2, HistoriaID: 2},
			},
			historias: []Historia{
				{ID: 1, Estado: EstadoHistoriaHecho, StoryPoints: punteroFloat64(8)},
				{ID: 2, Estado: EstadoHistoriaHecho, StoryPoints: punteroFloat64(20)},
			},
			esperado: punteroFloat64(8),
		},
		{
			nombre:  "promedio de cerrados",
			sprints: []Sprint{{ID: 1, Estado: EstadoSprintCerrado}, {ID: 2, Estado: EstadoSprintCerrado}},
			asignaciones: []SprintBacklogItem{
				{SprintID: 1, HistoriaID: 1},
				{SprintID: 2, HistoriaID: 2},
			},
			historias: []Historia{
				{ID: 1, Estado: EstadoHistoriaHecho, StoryPoints: punteroFloat64(8)},
				{ID: 2, Estado: EstadoHistoriaHecho, StoryPoints: punteroFloat64(4)},
			},
			esperado: punteroFloat64(6),
		},
		{
			nombre:       "historia no Hecho no suma",
			sprints:      []Sprint{{ID: 1, Estado: EstadoSprintCerrado}},
			asignaciones: []SprintBacklogItem{{SprintID: 1, HistoriaID: 1}},
			historias:    []Historia{{ID: 1, Estado: EstadoHistoriaEnProgreso, StoryPoints: punteroFloat64(8)}},
			esperado:     punteroFloat64(0),
		},
		{
			nombre:       "historia inexistente se ignora",
			sprints:      []Sprint{{ID: 1, Estado: EstadoSprintCerrado}},
			asignaciones: []SprintBacklogItem{{SprintID: 1, HistoriaID: 99}},
			esperado:     punteroFloat64(0),
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()
			assertFloatPtr(t, caso.esperado, CalcularVelocidad(caso.sprints, caso.asignaciones, caso.historias), "velocidad")
		})
	}
}

func TestCalcularRatioDefectos(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre               string
		defectos             []Defecto
		historiasCompletadas int
		esperado             *float64
	}{
		{
			nombre: "E-08 un rejected entre tres",
			defectos: []Defecto{
				{ID: 1, Estado: EstadoDefectoAbierto},
				{ID: 2, Estado: EstadoDefectoRechazado},
				{ID: 3, Estado: EstadoDefectoResuelto},
			},
			historiasCompletadas: 6,
			esperado:             punteroFloat64(1.0 / 3.0),
		},
		{
			nombre: "E-08 sin historias completadas",
			defectos: []Defecto{
				{ID: 1, Estado: EstadoDefectoAbierto},
				{ID: 2, Estado: EstadoDefectoRechazado},
				{ID: 3, Estado: EstadoDefectoResuelto},
			},
			historiasCompletadas: 0,
			esperado:             nil,
		},
		{
			nombre:               "E-08 solo rejected",
			defectos:             []Defecto{{ID: 1, Estado: EstadoDefectoRechazado}},
			historiasCompletadas: 6,
			esperado:             punteroFloat64(0),
		},
		{
			nombre:               "sin defectos",
			historiasCompletadas: 6,
			esperado:             punteroFloat64(0),
		},
		{
			nombre: "puede superar uno",
			defectos: []Defecto{
				{ID: 1, Estado: EstadoDefectoAbierto},
				{ID: 2, Estado: EstadoDefectoEnProgreso},
				{ID: 3, Estado: EstadoDefectoCerrado},
			},
			historiasCompletadas: 2,
			esperado:             punteroFloat64(1.5),
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()
			assertFloatPtr(t, caso.esperado, CalcularRatioDefectos(caso.defectos, caso.historiasCompletadas), "ratio de defectos")
		})
	}
}

func TestCalcularMetricasProyecto(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre       string
		escenario    string
		sprints      []Sprint
		asignaciones []SprintBacklogItem
		historias    []Historia
		defectos     []Defecto
		esperado     MetricasProyecto
	}{
		{
			nombre:    "E-04 velocidad y porcentaje de historias completadas",
			escenario: "E-04",
			sprints: []Sprint{
				{ID: 1, Estado: EstadoSprintCerrado},
				{ID: 2, Estado: EstadoSprintCerrado},
				{ID: 3, Estado: EstadoSprintActivo},
			},
			asignaciones: []SprintBacklogItem{
				{SprintID: 1, HistoriaID: 1},
				{SprintID: 2, HistoriaID: 2},
				{SprintID: 3, HistoriaID: 3},
			},
			historias: []Historia{
				{ID: 1, Estado: EstadoHistoriaHecho, StoryPoints: punteroFloat64(8)},
				{ID: 2, Estado: EstadoHistoriaHecho, StoryPoints: punteroFloat64(4)},
				{ID: 3, Estado: EstadoHistoriaBacklog, StoryPoints: punteroFloat64(3)},
				{ID: 4, Estado: EstadoHistoriaEnProgreso, StoryPoints: punteroFloat64(5)},
			},
			esperado: MetricasProyecto{
				Velocidad:                      punteroFloat64(6),
				PorcentajeHistoriasCompletadas: punteroFloat64(0.5),
				RatioDefectos:                  punteroFloat64(0),
			},
		},
		{
			nombre:    "E-06 dos pares",
			escenario: "E-06",
			historias: []Historia{
				{ID: 1, HorasEstimadas: punteroFloat64(5), EsfuerzoReal: punteroFloat64(8)},
				{ID: 2, HorasEstimadas: punteroFloat64(5), EsfuerzoReal: punteroFloat64(4)},
			},
			esperado: MetricasProyecto{
				PorcentajeHistoriasCompletadas: punteroFloat64(0),
				HorasEstimadas:                 10,
				HorasReales:                    12,
				DesviacionEsfuerzo:             punteroFloat64(0.2),
			},
		},
		{
			nombre:    "E-06 una sin horas reales queda excluida",
			escenario: "E-06",
			historias: []Historia{
				{ID: 1, HorasEstimadas: punteroFloat64(5), EsfuerzoReal: punteroFloat64(8)},
				{ID: 2, HorasEstimadas: punteroFloat64(5)},
			},
			esperado: MetricasProyecto{
				PorcentajeHistoriasCompletadas: punteroFloat64(0),
				HorasEstimadas:                 10,
				HorasReales:                    8,
				DesviacionEsfuerzo:             punteroFloat64(0.6),
			},
		},
		{
			nombre:    "E-06 sin pares",
			escenario: "E-06",
			historias: []Historia{
				{ID: 1, HorasEstimadas: punteroFloat64(5)},
				{ID: 2, EsfuerzoReal: punteroFloat64(4)},
			},
			esperado: MetricasProyecto{
				PorcentajeHistoriasCompletadas: punteroFloat64(0),
				HorasEstimadas:                 5,
				HorasReales:                    4,
			},
		},
		{
			nombre:    "E-08 rejected no suma al numerador del ratio",
			escenario: "E-08",
			historias: []Historia{
				{ID: 1, Estado: EstadoHistoriaHecho},
				{ID: 2, Estado: EstadoHistoriaHecho},
				{ID: 3, Estado: EstadoHistoriaHecho},
				{ID: 4, Estado: EstadoHistoriaHecho},
				{ID: 5, Estado: EstadoHistoriaHecho},
				{ID: 6, Estado: EstadoHistoriaHecho},
			},
			defectos: []Defecto{
				{ID: 1, Estado: EstadoDefectoAbierto},
				{ID: 2, Estado: EstadoDefectoRechazado},
				{ID: 3, Estado: EstadoDefectoResuelto},
			},
			esperado: MetricasProyecto{
				PorcentajeHistoriasCompletadas: punteroFloat64(1),
				DefectosDetectados:             3,
				DefectosResueltos:              1,
				RatioDefectos:                  punteroFloat64(1.0 / 3.0),
			},
		},
		{
			nombre:    "E-09 sin historias",
			escenario: "E-09",
			esperado:  MetricasProyecto{},
		},
		{
			nombre:    "E-09 sin historias Hecho",
			escenario: "E-09",
			historias: []Historia{
				{ID: 1, Estado: EstadoHistoriaBacklog},
				{ID: 2, Estado: EstadoHistoriaBacklog},
				{ID: 3, Estado: EstadoHistoriaBacklog},
				{ID: 4, Estado: EstadoHistoriaBacklog},
			},
			esperado: MetricasProyecto{
				PorcentajeHistoriasCompletadas: punteroFloat64(0),
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			obtenido := CalcularMetricasProyecto(caso.sprints, caso.asignaciones, caso.historias, caso.defectos)
			assertMetricasProyecto(t, caso.esperado, obtenido, caso.escenario)
		})
	}
}

func assertMetricasSprint(t *testing.T, esperado, obtenido MetricasSprint, escenario string) {
	t.Helper()

	assert.InDelta(t, esperado.StoryPointsPlanificados, obtenido.StoryPointsPlanificados, 1e-9, "escenario %s: story points planificados", escenario)
	assert.InDelta(t, esperado.StoryPointsCompletados, obtenido.StoryPointsCompletados, 1e-9, "escenario %s: story points completados", escenario)
	assertFloatPtr(t, esperado.PorcentajeCompletado, obtenido.PorcentajeCompletado, "escenario %s: porcentaje completado", escenario)
	assert.Equal(t, esperado.HistoriasCompletadas, obtenido.HistoriasCompletadas, "escenario %s: historias completadas", escenario)
	assert.InDelta(t, esperado.HorasEstimadas, obtenido.HorasEstimadas, 1e-9, "escenario %s: horas estimadas", escenario)
	assert.InDelta(t, esperado.HorasReales, obtenido.HorasReales, 1e-9, "escenario %s: horas reales", escenario)
	assertFloatPtr(t, esperado.DesviacionEsfuerzo, obtenido.DesviacionEsfuerzo, "escenario %s: desviacion de esfuerzo", escenario)
	assert.Equal(t, esperado.DefectosDetectados, obtenido.DefectosDetectados, "escenario %s: defectos detectados", escenario)
	assert.Equal(t, esperado.DefectosResueltos, obtenido.DefectosResueltos, "escenario %s: defectos resueltos", escenario)
}

func assertMetricasProyecto(t *testing.T, esperado, obtenido MetricasProyecto, escenario string) {
	t.Helper()

	assertFloatPtr(t, esperado.Velocidad, obtenido.Velocidad, "escenario %s: velocidad", escenario)
	assertFloatPtr(t, esperado.PorcentajeHistoriasCompletadas, obtenido.PorcentajeHistoriasCompletadas, "escenario %s: porcentaje de historias completadas", escenario)
	assert.InDelta(t, esperado.HorasEstimadas, obtenido.HorasEstimadas, 1e-9, "escenario %s: horas estimadas", escenario)
	assert.InDelta(t, esperado.HorasReales, obtenido.HorasReales, 1e-9, "escenario %s: horas reales", escenario)
	assertFloatPtr(t, esperado.DesviacionEsfuerzo, obtenido.DesviacionEsfuerzo, "escenario %s: desviacion de esfuerzo", escenario)
	assert.Equal(t, esperado.DefectosDetectados, obtenido.DefectosDetectados, "escenario %s: defectos detectados", escenario)
	assert.Equal(t, esperado.DefectosResueltos, obtenido.DefectosResueltos, "escenario %s: defectos resueltos", escenario)
	assertFloatPtr(t, esperado.RatioDefectos, obtenido.RatioDefectos, "escenario %s: ratio de defectos", escenario)
}

func assertFloatPtr(t *testing.T, esperado, obtenido *float64, msg string, args ...interface{}) {
	t.Helper()

	msgAndArgs := append([]interface{}{msg}, args...)
	if esperado == nil {
		assert.Nil(t, obtenido, msgAndArgs...)
		return
	}

	require.NotNil(t, obtenido, msgAndArgs...)
	assert.InDelta(t, *esperado, *obtenido, 1e-9, msgAndArgs...)
}

func punteroFloat64(valor float64) *float64 {
	return &valor
}
