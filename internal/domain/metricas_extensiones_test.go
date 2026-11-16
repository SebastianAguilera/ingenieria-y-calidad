package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalcularSPPendientes(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		historias []Historia
		esperado  float64
	}{
		{"sin historias", nil, 0},
		{
			"todas Hecho",
			[]Historia{
				{ID: 1, Estado: EstadoHistoriaHecho, StoryPoints: punteroFloat64(5)},
				{ID: 2, Estado: EstadoHistoriaHecho, StoryPoints: punteroFloat64(3)},
			},
			0,
		},
		{
			"estados mixtos",
			[]Historia{
				{ID: 1, Estado: EstadoHistoriaHecho, StoryPoints: punteroFloat64(5)},
				{ID: 2, Estado: EstadoHistoriaEnProgreso, StoryPoints: punteroFloat64(3)},
				{ID: 3, Estado: EstadoHistoriaBacklog, StoryPoints: punteroFloat64(2)},
			},
			5,
		},
		{
			"story points nil aportan cero",
			[]Historia{
				{ID: 1, Estado: EstadoHistoriaBacklog},
				{ID: 2, Estado: EstadoHistoriaEnProgreso, StoryPoints: punteroFloat64(2)},
			},
			2,
		},
	}

	for _, caso := range casos {
		caso := caso
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()
			assert.InDelta(t, caso.esperado, CalcularSPPendientes(caso.historias), 1e-9)
		})
	}
}

func TestCalcularSprintsRestantes(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		historias []Historia
		velocidad *float64
		esperado  *float64
	}{
		{
			nombre:    "E-01 cuarenta pendientes y velocidad seis",
			historias: historiasPendientes(40),
			velocidad: punteroFloat64(6),
			esperado:  punteroFloat64(7),
		},
		{
			nombre:    "E-02 sin cerrados velocidad nil",
			historias: historiasPendientes(40),
			velocidad: nil,
			esperado:  nil,
		},
		{
			nombre:    "E-02 velocidad cero",
			historias: historiasPendientes(40),
			velocidad: punteroFloat64(0),
			esperado:  nil,
		},
		{
			nombre:    "E-02 sin pendientes",
			historias: nil,
			velocidad: punteroFloat64(6),
			esperado:  punteroFloat64(0),
		},
		{
			nombre:    "E-02 cociente exacto",
			historias: historiasPendientes(12),
			velocidad: punteroFloat64(6),
			esperado:  punteroFloat64(2),
		},
		{
			nombre:    "E-02 con fraccion redondea hacia arriba",
			historias: historiasPendientes(7),
			velocidad: punteroFloat64(3),
			esperado:  punteroFloat64(3),
		},
	}

	for _, caso := range casos {
		caso := caso
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()
			obtenido := CalcularSprintsRestantes(caso.historias, caso.velocidad)
			if caso.esperado == nil {
				assert.Nil(t, obtenido)
				return
			}
			require.NotNil(t, obtenido)
			assert.InDelta(t, *caso.esperado, *obtenido, 1e-9)
		})
	}
}

func TestContarHistoriasEnProgreso(t *testing.T) {
	t.Parallel()

	historias := []Historia{
		{ID: 1, Estado: EstadoHistoriaEnProgreso},
		{ID: 2, Estado: EstadoHistoriaEnProgreso},
		{ID: 3, Estado: EstadoHistoriaBacklog},
		{ID: 4, Estado: EstadoHistoriaHecho},
	}

	assert.Equal(t, 2, ContarHistoriasEnProgreso(historias))
	assert.Equal(t, 0, ContarHistoriasEnProgreso(nil))
}

func TestCalcularTasaEntrega(t *testing.T) {
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
			historias:    []Historia{{ID: 1, Estado: EstadoHistoriaHecho}},
			esperado:     nil,
		},
		{
			nombre:  "E-04 dos cerrados con tres y uno",
			sprints: []Sprint{{ID: 1, Estado: EstadoSprintCerrado}, {ID: 2, Estado: EstadoSprintCerrado}},
			asignaciones: []SprintBacklogItem{
				{SprintID: 1, HistoriaID: 1},
				{SprintID: 1, HistoriaID: 2},
				{SprintID: 1, HistoriaID: 3},
				{SprintID: 2, HistoriaID: 4},
			},
			historias: []Historia{
				{ID: 1, Estado: EstadoHistoriaHecho},
				{ID: 2, Estado: EstadoHistoriaHecho},
				{ID: 3, Estado: EstadoHistoriaHecho},
				{ID: 4, Estado: EstadoHistoriaHecho},
			},
			esperado: punteroFloat64(2),
		},
		{
			nombre:   "E-05 un cerrado vacio aporta cero",
			sprints:  []Sprint{{ID: 1, Estado: EstadoSprintCerrado}},
			esperado: punteroFloat64(0),
		},
		{
			nombre:  "E-05 activo excluido",
			sprints: []Sprint{{ID: 1, Estado: EstadoSprintCerrado}, {ID: 2, Estado: EstadoSprintActivo}},
			asignaciones: []SprintBacklogItem{
				{SprintID: 1, HistoriaID: 1},
				{SprintID: 1, HistoriaID: 2},
				{SprintID: 2, HistoriaID: 3},
			},
			historias: []Historia{
				{ID: 1, Estado: EstadoHistoriaHecho},
				{ID: 2, Estado: EstadoHistoriaHecho},
				{ID: 3, Estado: EstadoHistoriaHecho},
			},
			esperado: punteroFloat64(2),
		},
		{
			nombre:  "E-05 cerrado vacio en el divisor",
			sprints: []Sprint{{ID: 1, Estado: EstadoSprintCerrado}, {ID: 2, Estado: EstadoSprintCerrado}},
			asignaciones: []SprintBacklogItem{
				{SprintID: 1, HistoriaID: 1},
				{SprintID: 1, HistoriaID: 2},
			},
			historias: []Historia{
				{ID: 1, Estado: EstadoHistoriaHecho},
				{ID: 2, Estado: EstadoHistoriaHecho},
			},
			esperado: punteroFloat64(1),
		},
		{
			nombre:  "E-05 historia no Hecho no cuenta",
			sprints: []Sprint{{ID: 1, Estado: EstadoSprintCerrado}},
			asignaciones: []SprintBacklogItem{
				{SprintID: 1, HistoriaID: 1},
				{SprintID: 1, HistoriaID: 2},
			},
			historias: []Historia{
				{ID: 1, Estado: EstadoHistoriaHecho},
				{ID: 2, Estado: EstadoHistoriaEnProgreso},
			},
			esperado: punteroFloat64(1),
		},
		{
			nombre:       "E-05 historia Hecho no asignada no cuenta",
			sprints:      []Sprint{{ID: 1, Estado: EstadoSprintCerrado}},
			asignaciones: []SprintBacklogItem{{SprintID: 1, HistoriaID: 1}},
			historias: []Historia{
				{ID: 1, Estado: EstadoHistoriaHecho},
				{ID: 2, Estado: EstadoHistoriaHecho},
			},
			esperado: punteroFloat64(1),
		},
	}

	for _, caso := range casos {
		caso := caso
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()
			obtenido := CalcularTasaEntrega(caso.sprints, caso.asignaciones, caso.historias)
			if caso.esperado == nil {
				assert.Nil(t, obtenido)
				return
			}
			require.NotNil(t, obtenido)
			assert.InDelta(t, *caso.esperado, *obtenido, 1e-9)
		})
	}
}

func TestCalcularLeadTimeEstimado(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		wip         int
		tasaEntrega *float64
		esperado    *float64
	}{
		{"E-06 wip seis y tasa dos", 6, punteroFloat64(2), punteroFloat64(3)},
		{"E-07 tasa nil", 6, nil, nil},
		{"E-07 tasa cero", 6, punteroFloat64(0), nil},
		{"E-08 wip cero con tasa", 0, punteroFloat64(2), punteroFloat64(0)},
	}

	for _, caso := range casos {
		caso := caso
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()
			obtenido := CalcularLeadTimeEstimado(caso.wip, caso.tasaEntrega)
			if caso.esperado == nil {
				assert.Nil(t, obtenido)
				return
			}
			require.NotNil(t, obtenido)
			assert.InDelta(t, *caso.esperado, *obtenido, 1e-9)
		})
	}
}

func TestContarDefectosAbiertos(t *testing.T) {
	t.Parallel()

	defectos := []Defecto{
		{ID: 1, Estado: EstadoDefectoAbierto},
		{ID: 2, Estado: EstadoDefectoEnProgreso},
		{ID: 3, Estado: EstadoDefectoResuelto},
		{ID: 4, Estado: EstadoDefectoCerrado},
		{ID: 5, Estado: EstadoDefectoRechazado},
	}

	assert.Equal(t, 1, ContarDefectosAbiertos(defectos))
	assert.Equal(t, 0, ContarDefectosAbiertos(nil))
}

func TestCalcularDefectosPorSeveridad(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre   string
		defectos []Defecto
		esperado DefectosPorSeveridad
	}{
		{
			nombre: "E-09 desglose con blocker en cero",
			defectos: []Defecto{
				{ID: 1, Severidad: SeveridadMinor, Estado: EstadoDefectoAbierto},
				{ID: 2, Severidad: SeveridadMajor, Estado: EstadoDefectoResuelto},
				{ID: 3, Severidad: SeveridadCritical, Estado: EstadoDefectoAbierto},
				{ID: 4, Severidad: SeveridadCritical, Estado: EstadoDefectoRechazado},
			},
			esperado: DefectosPorSeveridad{Minor: 1, Major: 1, Critical: 2, Blocker: 0},
		},
		{
			nombre:   "E-09 sin defectos",
			esperado: DefectosPorSeveridad{},
		},
		{
			nombre: "todas las severidades",
			defectos: []Defecto{
				{ID: 1, Severidad: SeveridadMinor},
				{ID: 2, Severidad: SeveridadMajor},
				{ID: 3, Severidad: SeveridadCritical},
				{ID: 4, Severidad: SeveridadBlocker},
			},
			esperado: DefectosPorSeveridad{Minor: 1, Major: 1, Critical: 1, Blocker: 1},
		},
		{
			nombre: "severidad invalida se ignora",
			defectos: []Defecto{
				{ID: 1, Severidad: SeveridadDefecto("desconocida")},
			},
			esperado: DefectosPorSeveridad{},
		},
	}

	for _, caso := range casos {
		caso := caso
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, caso.esperado, CalcularDefectosPorSeveridad(caso.defectos))
		})
	}
}

func TestCalcularMetricasProyectoExtensiones(t *testing.T) {
	t.Parallel()

	t.Run("E-01 caso completo con flujo y defectos", func(t *testing.T) {
		t.Parallel()

		sprints := []Sprint{
			{ID: 1, Estado: EstadoSprintCerrado},
			{ID: 2, Estado: EstadoSprintCerrado},
			{ID: 3, Estado: EstadoSprintActivo},
		}
		asignaciones := []SprintBacklogItem{
			{SprintID: 1, HistoriaID: 1},
			{SprintID: 1, HistoriaID: 2},
			{SprintID: 1, HistoriaID: 3},
			{SprintID: 2, HistoriaID: 4},
			{SprintID: 3, HistoriaID: 5},
			{SprintID: 3, HistoriaID: 6},
		}
		historias := []Historia{
			{ID: 1, Estado: EstadoHistoriaHecho, StoryPoints: punteroFloat64(5)},
			{ID: 2, Estado: EstadoHistoriaHecho, StoryPoints: punteroFloat64(3)},
			{ID: 3, Estado: EstadoHistoriaHecho, StoryPoints: punteroFloat64(2)},
			{ID: 4, Estado: EstadoHistoriaHecho, StoryPoints: punteroFloat64(4)},
			{ID: 5, Estado: EstadoHistoriaEnProgreso, StoryPoints: punteroFloat64(8)},
			{ID: 6, Estado: EstadoHistoriaBacklog, StoryPoints: punteroFloat64(13)},
			{ID: 7, Estado: EstadoHistoriaBacklog, StoryPoints: punteroFloat64(2)},
			{ID: 8, Estado: EstadoHistoriaHecho, StoryPoints: punteroFloat64(5)},
			{ID: 9, Estado: EstadoHistoriaEnProgreso, StoryPoints: punteroFloat64(1)},
		}
		defectos := []Defecto{
			{ID: 1, Severidad: SeveridadMinor, Estado: EstadoDefectoAbierto},
			{ID: 2, Severidad: SeveridadMajor, Estado: EstadoDefectoResuelto},
			{ID: 3, Severidad: SeveridadCritical, Estado: EstadoDefectoAbierto},
			{ID: 4, Severidad: SeveridadCritical, Estado: EstadoDefectoRechazado},
			{ID: 5, Severidad: SeveridadBlocker, Estado: EstadoDefectoEnProgreso},
		}

		metricas := CalcularMetricasProyecto(sprints, asignaciones, historias, defectos)

		require.NotNil(t, metricas.Velocidad)
		assert.InDelta(t, 7, *metricas.Velocidad, 1e-9, "SP completados: 10 del Sprint 1 y 4 del Sprint 2")
		require.NotNil(t, metricas.SprintsRestantes)
		assert.InDelta(t, 4, *metricas.SprintsRestantes, 1e-9, "ceil(24/7)")
		assert.Equal(t, 2, metricas.WIP)
		require.NotNil(t, metricas.TasaEntrega)
		assert.InDelta(t, 2, *metricas.TasaEntrega, 1e-9)
		require.NotNil(t, metricas.LeadTimeEstimado)
		assert.InDelta(t, 1, *metricas.LeadTimeEstimado, 1e-9)
		assert.True(t, metricas.LeadTimeAproximado)
		assert.Equal(t, 2, metricas.DefectosAbiertos)
		assert.Equal(t, DefectosPorSeveridad{Minor: 1, Major: 1, Critical: 2, Blocker: 1}, metricas.DefectosPorSeveridad)
	})

	t.Run("E-09 proyecto sin datos", func(t *testing.T) {
		t.Parallel()

		metricas := CalcularMetricasProyecto(nil, nil, nil, nil)

		assert.Nil(t, metricas.SprintsRestantes)
		assert.Equal(t, 0, metricas.WIP)
		assert.Nil(t, metricas.TasaEntrega)
		assert.Nil(t, metricas.LeadTimeEstimado)
		assert.False(t, metricas.LeadTimeAproximado)
		assert.Equal(t, 0, metricas.DefectosAbiertos)
		assert.Equal(t, DefectosPorSeveridad{}, metricas.DefectosPorSeveridad)
	})

	t.Run("E-07 tasa cero deja lead time sin base", func(t *testing.T) {
		t.Parallel()

		sprints := []Sprint{{ID: 1, Estado: EstadoSprintCerrado}}
		historias := []Historia{
			{ID: 1, Estado: EstadoHistoriaBacklog, StoryPoints: punteroFloat64(6)},
			{ID: 2, Estado: EstadoHistoriaEnProgreso, StoryPoints: punteroFloat64(3)},
		}

		metricas := CalcularMetricasProyecto(sprints, nil, historias, nil)

		require.NotNil(t, metricas.TasaEntrega)
		assert.InDelta(t, 0, *metricas.TasaEntrega, 1e-9)
		assert.Nil(t, metricas.LeadTimeEstimado)
		assert.False(t, metricas.LeadTimeAproximado)
		assert.Equal(t, 1, metricas.WIP)
	})
}

func historiasPendientes(storyPoints float64) []Historia {
	return []Historia{
		{ID: 1, Estado: EstadoHistoriaBacklog, StoryPoints: punteroFloat64(storyPoints)},
	}
}
