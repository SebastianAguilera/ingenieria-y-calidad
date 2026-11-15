package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func TestReporteServiceObtenerReporteProyecto(t *testing.T) {
	t.Parallel()
	proyectos := new(MockProyectoRepository)
	sprints := new(MockSprintRepository)
	backlog := new(MockSprintBacklogRepository)
	historias := new(MockHistoriaRepository)
	defectos := new(MockDefectoRepository)

	metricasSvc := NewMetricasService(proyectos, sprints, backlog, historias, defectos)
	svc := NewReporteService(proyectos, sprints, backlog, historias, defectos, metricasSvc)
	ctx := context.Background()

	now := time.Now()
	proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{
		ID:          1,
		Nombre:      "Proyecto Test",
		Descripcion: "Desc",
		FechaInicio: now,
		Estado:      domain.EstadoActivo,
	}, nil).Times(3)

	sprintsList := []domain.Sprint{
		{ID: 1, ProyectoID: 1, Estado: domain.EstadoSprintCerrado},
	}
	backlogList := []domain.SprintBacklogItem{
		{SprintID: 1, HistoriaID: 10},
	}
	sp5 := 5.0
	historiasList := []domain.Historia{
		{ID: 10, ProyectoID: 1, Codigo: "US-01", Titulo: "Historia 1", Estado: domain.EstadoHistoriaHecho, StoryPoints: &sp5},
	}
	defectosList := []domain.Defecto{
		{ID: 100, ProyectoID: 1, Descripcion: "Bug 1", Severidad: domain.SeveridadMajor, Estado: domain.EstadoDefectoAbierto, SprintDeteccionID: 1},
	}

	sprints.On("ListarPorProyecto", mock.Anything, uint(1)).Return(sprintsList, nil).Times(2)
	backlog.On("ListarPorProyecto", mock.Anything, uint(1)).Return(backlogList, nil).Times(2)
	historias.On("ListarPorProyecto", mock.Anything, uint(1)).Return(historiasList, nil).Times(3)
	defectos.On("ListarPorProyecto", mock.Anything, uint(1)).Return(defectosList, nil).Times(3)

	rep, err := svc.ObtenerReporteProyecto(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "estado_actual", rep.Base)
	assert.Equal(t, uint(1), rep.ProyectoID)
	assert.Equal(t, "Proyecto Test", rep.Nombre)
	assert.Len(t, rep.Historias, 1)
	assert.Nil(t, rep.Historias[0].HorasEstimadas)
	assert.Len(t, rep.Defectos, 1)

	// Comparación directa con US-20/US-21
	metricasEsperadas, err := metricasSvc.CalcularPorProyecto(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, metricasEsperadas, rep.Metricas)

	proyectos.AssertExpectations(t)
	sprints.AssertExpectations(t)
	backlog.AssertExpectations(t)
	historias.AssertExpectations(t)
	defectos.AssertExpectations(t)
}

func TestReporteServiceObtenerReporteProyectoSinHistorias(t *testing.T) {
	t.Parallel()
	proyectos := new(MockProyectoRepository)
	sprints := new(MockSprintRepository)
	backlog := new(MockSprintBacklogRepository)
	historias := new(MockHistoriaRepository)
	defectos := new(MockDefectoRepository)

	metricasSvc := NewMetricasService(proyectos, sprints, backlog, historias, defectos)
	svc := NewReporteService(proyectos, sprints, backlog, historias, defectos, metricasSvc)
	ctx := context.Background()

	now := time.Now()
	proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{
		ID:          1,
		Nombre:      "Proyecto Vacio",
		FechaInicio: now,
		Estado:      domain.EstadoActivo,
	}, nil).Times(2)

	sprints.On("ListarPorProyecto", mock.Anything, uint(1)).Return([]domain.Sprint{}, nil).Once()
	backlog.On("ListarPorProyecto", mock.Anything, uint(1)).Return([]domain.SprintBacklogItem{}, nil).Once()
	historias.On("ListarPorProyecto", mock.Anything, uint(1)).Return([]domain.Historia(nil), nil).Times(2)
	defectos.On("ListarPorProyecto", mock.Anything, uint(1)).Return([]domain.Defecto(nil), nil).Times(2)

	rep, err := svc.ObtenerReporteProyecto(ctx, 1)
	require.NoError(t, err)
	assert.NotNil(t, rep.Historias)
	assert.Empty(t, rep.Historias)
	assert.NotNil(t, rep.Defectos)
	assert.Empty(t, rep.Defectos)
	assert.Nil(t, rep.Metricas.Velocidad)

	proyectos.AssertExpectations(t)
	sprints.AssertExpectations(t)
	backlog.AssertExpectations(t)
	historias.AssertExpectations(t)
	defectos.AssertExpectations(t)
}

func TestReporteServiceObtenerReporteSprint(t *testing.T) {
	t.Parallel()
	proyectos := new(MockProyectoRepository)
	sprints := new(MockSprintRepository)
	backlog := new(MockSprintBacklogRepository)
	historias := new(MockHistoriaRepository)
	defectos := new(MockDefectoRepository)

	metricasSvc := NewMetricasService(proyectos, sprints, backlog, historias, defectos)
	svc := NewReporteService(proyectos, sprints, backlog, historias, defectos, metricasSvc)
	ctx := context.Background()

	now := time.Now()
	proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{
		ID:          1,
		Nombre:      "Proyecto Test",
		FechaInicio: now,
		Estado:      domain.EstadoActivo,
	}, nil).Times(3)

	sprints.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(&domain.Sprint{
		ID:            10,
		ProyectoID:    1,
		Identificador: "Sprint 10",
		Estado:        domain.EstadoSprintActivo,
		FechaInicio:   now,
		FechaFin:      now,
	}, nil).Times(3)

	backlog.On("ObtenerPorSprint", mock.Anything, uint(1), uint(10)).Return([]domain.SprintBacklogItem{
		{SprintID: 10, HistoriaID: 100},
	}, nil).Times(3)

	historias.On("ObtenerPorIDs", mock.Anything, uint(1), []uint{100}).Return([]domain.Historia{
		{ID: 100, ProyectoID: 1, Codigo: "US-100", Titulo: "Historia Sprint", Estado: domain.EstadoHistoriaHecho},
	}, nil).Times(3)

	res10 := uint(10)
	res20 := uint(20)
	defectos.On("ListarPorProyecto", mock.Anything, uint(1)).Return([]domain.Defecto{
		{ID: 1, ProyectoID: 1, SprintDeteccionID: 10, SprintResolucionID: &res10, Estado: domain.EstadoDefectoResuelto, Severidad: domain.SeveridadCritical, Descripcion: "Detectado y resuelto en 10"},
		{ID: 2, ProyectoID: 1, SprintDeteccionID: 5, SprintResolucionID: &res10, Estado: domain.EstadoDefectoCerrado, Severidad: domain.SeveridadMajor, Descripcion: "Detectado en 5, resuelto en 10"},
		{ID: 3, ProyectoID: 1, SprintDeteccionID: 20, SprintResolucionID: &res20, Estado: domain.EstadoDefectoResuelto, Severidad: domain.SeveridadMinor, Descripcion: "De otro sprint 20"},
	}, nil).Times(3)

	rep, err := svc.ObtenerReporteSprint(ctx, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, "estado_actual", rep.Base)
	assert.Equal(t, uint(10), rep.SprintID)
	assert.Len(t, rep.Historias, 1)
	assert.Len(t, rep.Defectos, 2)
	assert.Equal(t, 1, rep.Metricas.DefectosDetectados)
	assert.Equal(t, 2, rep.Metricas.DefectosResueltos)

	// Comparación directa con US-20
	metricasEsperadas, err := metricasSvc.CalcularPorSprint(ctx, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, metricasEsperadas, rep.Metricas)

	proyectos.AssertExpectations(t)
	sprints.AssertExpectations(t)
	backlog.AssertExpectations(t)
	historias.AssertExpectations(t)
	defectos.AssertExpectations(t)
}

func TestReporteServiceErrores(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("proyecto inexistente", func(t *testing.T) {
		t.Parallel()
		proyectos := new(MockProyectoRepository)
		sprints := new(MockSprintRepository)
		backlog := new(MockSprintBacklogRepository)
		historias := new(MockHistoriaRepository)
		defectos := new(MockDefectoRepository)
		metricasSvc := NewMetricasService(proyectos, sprints, backlog, historias, defectos)
		svc := NewReporteService(proyectos, sprints, backlog, historias, defectos, metricasSvc)

		proyectos.On("ObtenerPorID", mock.Anything, uint(999)).Return(nil, domain.ErrProyectoNoEncontrado).Once()
		_, err := svc.ObtenerReporteProyecto(ctx, 999)
		assert.ErrorIs(t, err, domain.ErrProyectoNoEncontrado)
	})

	t.Run("sprint inexistente", func(t *testing.T) {
		t.Parallel()
		proyectos := new(MockProyectoRepository)
		sprints := new(MockSprintRepository)
		backlog := new(MockSprintBacklogRepository)
		historias := new(MockHistoriaRepository)
		defectos := new(MockDefectoRepository)
		metricasSvc := NewMetricasService(proyectos, sprints, backlog, historias, defectos)
		svc := NewReporteService(proyectos, sprints, backlog, historias, defectos, metricasSvc)

		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil).Once()
		sprints.On("ObtenerPorID", mock.Anything, uint(1), uint(999)).Return(nil, domain.ErrSprintNoEncontrado).Once()
		_, err := svc.ObtenerReporteSprint(ctx, 1, 999)
		assert.ErrorIs(t, err, domain.ErrSprintNoEncontrado)
	})
}
