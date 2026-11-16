package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func TestMetricasServiceCalcularPorSprint(t *testing.T) {
	t.Parallel()
	proyectos := new(MockProyectoRepository)
	sprints := new(MockSprintRepository)
	backlog := new(MockSprintBacklogRepository)
	historias := new(MockHistoriaRepository)
	defectos := new(MockDefectoRepository)

	svc := NewMetricasService(proyectos, sprints, backlog, historias, defectos)
	ctx := context.Background()

	proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
	sprints.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(&domain.Sprint{ID: 10, ProyectoID: 1, Estado: domain.EstadoSprintActivo}, nil)
	backlog.On("ObtenerPorSprint", mock.Anything, uint(1), uint(10)).Return([]domain.SprintBacklogItem{
		{ID: 1, ProyectoID: 1, SprintID: 10, HistoriaID: 100},
	}, nil)

	sp5 := 5.0
	he10 := 10.0
	hr12 := 12.0
	historias.On("ObtenerPorIDs", mock.Anything, uint(1), []uint{100}).Return([]domain.Historia{
		{
			ID:             100,
			ProyectoID:     1,
			Estado:         domain.EstadoHistoriaHecho,
			StoryPoints:    &sp5,
			HorasEstimadas: &he10,
			EsfuerzoReal:   &hr12,
		},
	}, nil)

	resID := uint(10)
	defectos.On("ListarPorProyecto", mock.Anything, uint(1)).Return([]domain.Defecto{
		{
			ID:                 1,
			ProyectoID:         1,
			SprintDeteccionID:  10,
			SprintResolucionID: &resID,
			Estado:             domain.EstadoDefectoResuelto,
		},
	}, nil)

	m, err := svc.CalcularPorSprint(ctx, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 5.0, m.StoryPointsPlanificados)
	assert.Equal(t, 5.0, m.StoryPointsCompletados)
	require.NotNil(t, m.PorcentajeCompletado)
	assert.Equal(t, 1.0, *m.PorcentajeCompletado)
	assert.Equal(t, 1, m.HistoriasCompletadas)
	assert.Equal(t, 10.0, m.HorasEstimadas)
	assert.Equal(t, 12.0, m.HorasReales)
	require.NotNil(t, m.DesviacionEsfuerzo)
	assert.InDelta(t, 0.2, *m.DesviacionEsfuerzo, 1e-9)
	assert.Equal(t, 1, m.DefectosDetectados)
	assert.Equal(t, 1, m.DefectosResueltos)

	proyectos.AssertExpectations(t)
	sprints.AssertExpectations(t)
	backlog.AssertExpectations(t)
	historias.AssertExpectations(t)
	defectos.AssertExpectations(t)
}

func TestMetricasServiceCalcularPorSprintErrores(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("proyecto inexistente", func(t *testing.T) {
		t.Parallel()
		proyectos := new(MockProyectoRepository)
		svc := NewMetricasService(proyectos, new(MockSprintRepository), new(MockSprintBacklogRepository), new(MockHistoriaRepository), new(MockDefectoRepository))

		proyectos.On("ObtenerPorID", mock.Anything, uint(999)).Return(nil, domain.ErrProyectoNoEncontrado)
		_, err := svc.CalcularPorSprint(ctx, 999, 1)
		assert.ErrorIs(t, err, domain.ErrProyectoNoEncontrado)
	})

	t.Run("sprint inexistente", func(t *testing.T) {
		t.Parallel()
		proyectos := new(MockProyectoRepository)
		sprints := new(MockSprintRepository)
		svc := NewMetricasService(proyectos, sprints, new(MockSprintBacklogRepository), new(MockHistoriaRepository), new(MockDefectoRepository))

		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
		sprints.On("ObtenerPorID", mock.Anything, uint(1), uint(999)).Return(nil, domain.ErrSprintNoEncontrado)
		_, err := svc.CalcularPorSprint(ctx, 1, 999)
		assert.ErrorIs(t, err, domain.ErrSprintNoEncontrado)
	})
}

func TestMetricasServiceCalcularPorProyecto(t *testing.T) {
	t.Parallel()
	proyectos := new(MockProyectoRepository)
	sprints := new(MockSprintRepository)
	backlog := new(MockSprintBacklogRepository)
	historias := new(MockHistoriaRepository)
	defectos := new(MockDefectoRepository)

	svc := NewMetricasService(proyectos, sprints, backlog, historias, defectos)
	ctx := context.Background()

	proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil).Once()
	sprints.On("ListarPorProyecto", mock.Anything, uint(1)).Return([]domain.Sprint{}, nil).Once()
	backlog.On("ListarPorProyecto", mock.Anything, uint(1)).Return([]domain.SprintBacklogItem{}, nil).Once()
	historias.On("ListarPorProyecto", mock.Anything, uint(1)).Return([]domain.Historia{}, nil).Once()
	defectos.On("ListarPorProyecto", mock.Anything, uint(1)).Return([]domain.Defecto{}, nil).Once()

	m, err := svc.CalcularPorProyecto(ctx, 1)
	require.NoError(t, err)
	assert.Nil(t, m.Velocidad)
	assert.Nil(t, m.PorcentajeHistoriasCompletadas)
	assert.Equal(t, 0.0, m.HorasEstimadas)
	assert.Equal(t, 0.0, m.HorasReales)
	assert.Nil(t, m.DesviacionEsfuerzo)
	assert.Equal(t, 0, m.DefectosDetectados)
	assert.Equal(t, 0, m.DefectosResueltos)
	assert.Nil(t, m.RatioDefectos)

	proyectos.AssertExpectations(t)
	sprints.AssertExpectations(t)
	backlog.AssertExpectations(t)
	historias.AssertExpectations(t)
	defectos.AssertExpectations(t)
}

func TestMetricasServiceCalcularPorSprintSinHistoriasNoConsultaHistorias(t *testing.T) {
	t.Parallel()
	proyectos := new(MockProyectoRepository)
	sprints := new(MockSprintRepository)
	backlog := new(MockSprintBacklogRepository)
	historias := new(MockHistoriaRepository)
	defectos := new(MockDefectoRepository)

	svc := NewMetricasService(proyectos, sprints, backlog, historias, defectos)
	ctx := context.Background()

	proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil).Once()
	sprints.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(&domain.Sprint{ID: 10, ProyectoID: 1}, nil).Once()
	backlog.On("ObtenerPorSprint", mock.Anything, uint(1), uint(10)).Return([]domain.SprintBacklogItem{}, nil).Once()
	defectos.On("ListarPorProyecto", mock.Anything, uint(1)).Return([]domain.Defecto{}, nil).Once()

	m, err := svc.CalcularPorSprint(ctx, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 0.0, m.StoryPointsPlanificados)
	assert.Equal(t, 0.0, m.StoryPointsCompletados)
	assert.Nil(t, m.PorcentajeCompletado)
	assert.Equal(t, 0, m.HistoriasCompletadas)
	assert.Nil(t, m.DesviacionEsfuerzo)

	historias.AssertNotCalled(t, "ObtenerPorIDs", mock.Anything, mock.Anything, mock.Anything)
	proyectos.AssertExpectations(t)
	sprints.AssertExpectations(t)
	backlog.AssertExpectations(t)
	defectos.AssertExpectations(t)
}

func TestMetricasServiceCalcularPorProyectoExtensiones(t *testing.T) {
	t.Parallel()
	proyectos := new(MockProyectoRepository)
	sprints := new(MockSprintRepository)
	backlog := new(MockSprintBacklogRepository)
	historias := new(MockHistoriaRepository)
	defectos := new(MockDefectoRepository)

	svc := NewMetricasService(proyectos, sprints, backlog, historias, defectos)
	ctx := context.Background()

	proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil).Once()
	sprints.On("ListarPorProyecto", mock.Anything, uint(1)).Return([]domain.Sprint{
		{ID: 1, ProyectoID: 1, Estado: domain.EstadoSprintCerrado},
		{ID: 2, ProyectoID: 1, Estado: domain.EstadoSprintCerrado},
		{ID: 3, ProyectoID: 1, Estado: domain.EstadoSprintActivo},
	}, nil).Once()
	backlog.On("ListarPorProyecto", mock.Anything, uint(1)).Return([]domain.SprintBacklogItem{
		{SprintID: 1, HistoriaID: 100},
		{SprintID: 1, HistoriaID: 101},
		{SprintID: 1, HistoriaID: 102},
		{SprintID: 2, HistoriaID: 103},
		{SprintID: 3, HistoriaID: 104},
		{SprintID: 3, HistoriaID: 105},
		{SprintID: 3, HistoriaID: 106},
		{SprintID: 3, HistoriaID: 107},
		{SprintID: 3, HistoriaID: 108},
		{SprintID: 3, HistoriaID: 109},
	}, nil).Once()
	historias.On("ListarPorProyecto", mock.Anything, uint(1)).Return([]domain.Historia{
		{ID: 100, Estado: domain.EstadoHistoriaHecho, StoryPoints: punteroFloat64(5)},
		{ID: 101, Estado: domain.EstadoHistoriaHecho, StoryPoints: punteroFloat64(3)},
		{ID: 102, Estado: domain.EstadoHistoriaHecho, StoryPoints: punteroFloat64(2)},
		{ID: 103, Estado: domain.EstadoHistoriaHecho, StoryPoints: punteroFloat64(2)},
		{ID: 104, Estado: domain.EstadoHistoriaEnProgreso},
		{ID: 105, Estado: domain.EstadoHistoriaEnProgreso},
		{ID: 106, Estado: domain.EstadoHistoriaEnProgreso},
		{ID: 107, Estado: domain.EstadoHistoriaEnProgreso},
		{ID: 108, Estado: domain.EstadoHistoriaEnProgreso},
		{ID: 109, Estado: domain.EstadoHistoriaEnProgreso},
		{ID: 110, Estado: domain.EstadoHistoriaBacklog, StoryPoints: punteroFloat64(40)},
	}, nil).Once()
	defectos.On("ListarPorProyecto", mock.Anything, uint(1)).Return([]domain.Defecto{
		{ID: 1, Severidad: domain.SeveridadMinor, Estado: domain.EstadoDefectoAbierto},
		{ID: 2, Severidad: domain.SeveridadMajor, Estado: domain.EstadoDefectoResuelto},
		{ID: 3, Severidad: domain.SeveridadCritical, Estado: domain.EstadoDefectoAbierto},
		{ID: 4, Severidad: domain.SeveridadCritical, Estado: domain.EstadoDefectoRechazado},
	}, nil).Once()

	m, err := svc.CalcularPorProyecto(ctx, 1)
	require.NoError(t, err)

	require.NotNil(t, m.Velocidad)
	assert.InDelta(t, 6, *m.Velocidad, 1e-9)
	require.NotNil(t, m.SprintsRestantes)
	assert.InDelta(t, 7, *m.SprintsRestantes, 1e-9)
	assert.Equal(t, 6, m.WIP)
	require.NotNil(t, m.TasaEntrega)
	assert.InDelta(t, 2, *m.TasaEntrega, 1e-9)
	require.NotNil(t, m.LeadTimeEstimado)
	assert.InDelta(t, 3, *m.LeadTimeEstimado, 1e-9)
	assert.True(t, m.LeadTimeAproximado)
	assert.Equal(t, 2, m.DefectosAbiertos)
	assert.Equal(t, domain.DefectosPorSeveridad{Minor: 1, Major: 1, Critical: 2, Blocker: 0}, m.DefectosPorSeveridad)

	proyectos.AssertExpectations(t)
	sprints.AssertExpectations(t)
	backlog.AssertExpectations(t)
	historias.AssertExpectations(t)
	defectos.AssertExpectations(t)
}
