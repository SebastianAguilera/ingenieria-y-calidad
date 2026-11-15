package service

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func nuevoPDFServiceConMocks(t *testing.T) (
	*PDFService,
	*MockProyectoRepository,
	*MockSprintRepository,
	*MockSprintBacklogRepository,
	*MockHistoriaRepository,
	*MockDefectoRepository,
) {
	t.Helper()
	proyectos := new(MockProyectoRepository)
	sprints := new(MockSprintRepository)
	backlog := new(MockSprintBacklogRepository)
	historias := new(MockHistoriaRepository)
	defectos := new(MockDefectoRepository)
	metricas := NewMetricasService(proyectos, sprints, backlog, historias, defectos)
	reportes := NewReporteService(proyectos, sprints, backlog, historias, defectos, metricas)
	svc := NewPDFService(reportes)
	return svc, proyectos, sprints, backlog, historias, defectos
}

// ---------------------------------------------------------------------------
// Tests del constructor de PDF puro (sin mocks, solo domain structs)
// ---------------------------------------------------------------------------

func TestGenerarPDFProyectoPuro(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	vel := 5.0
	sp5 := 5.0
	reporte := domain.ReporteProyecto{
		Base:        "estado_actual",
		ProyectoID:  1,
		Nombre:      "Proyecto Alpha",
		Descripcion: "Descripcion de prueba",
		FechaInicio: now,
		Estado:      domain.EstadoActivo,
		Metricas: domain.MetricasProyecto{
			Velocidad: &vel,
		},
		Historias: []domain.Historia{
			{ID: 10, Codigo: "US-01", Titulo: "Login", Estado: domain.EstadoHistoriaHecho, StoryPoints: &sp5},
		},
		Defectos: []domain.Defecto{
			{ID: 100, Descripcion: "Bug login", Severidad: domain.SeveridadCritical, Estado: domain.EstadoDefectoAbierto, SprintDeteccionID: 1},
		},
	}

	pdfBytes, err := generarPDFProyecto(reporte)
	require.NoError(t, err)
	assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF")), "debe ser un PDF valido")
	assert.Greater(t, len(pdfBytes), 100, "el PDF debe tener contenido")
	contenido := string(pdfBytes)
	assert.Contains(t, contenido, "Encabezado")
	assert.Contains(t, contenido, "Metricas")
	assert.Contains(t, contenido, "Historias")
	assert.Contains(t, contenido, "Defectos")
}

func TestGenerarPDFProyectoMetricasNulas(t *testing.T) {
	t.Parallel()

	reporte := domain.ReporteProyecto{
		Base:        "estado_actual",
		ProyectoID:  2,
		Nombre:      "Proyecto Vacio",
		FechaInicio: time.Now(),
		Estado:      domain.EstadoActivo,
		Metricas:    domain.MetricasProyecto{},
		Historias:   []domain.Historia{},
		Defectos:    []domain.Defecto{},
	}

	pdfBytes, err := generarPDFProyecto(reporte)
	require.NoError(t, err)
	assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF")))
	contenido := string(pdfBytes)
	assert.Contains(t, contenido, "Sin datos", "las metricas null deben mostrar 'Sin datos'")
}

func TestGenerarPDFSprintPuro(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	det := 2
	res := 1
	reporte := domain.ReporteSprint{
		Base:          "estado_actual",
		ProyectoID:    1,
		SprintID:      10,
		Identificador: "Sprint-10",
		SprintGoal:    "Terminar login",
		FechaInicio:   now,
		FechaFin:      now.AddDate(0, 0, 14),
		Estado:        domain.EstadoSprintActivo,
		Metricas: domain.MetricasSprint{
			StoryPointsPlanificados: 10,
			StoryPointsCompletados:  5,
			DefectosDetectados:      det,
			DefectosResueltos:       res,
		},
		Historias: []domain.Historia{
			{ID: 100, Codigo: "US-100", Titulo: "Historia Sprint", Estado: domain.EstadoHistoriaHecho},
		},
		Defectos: []domain.Defecto{},
	}

	pdfBytes, err := generarPDFSprint(reporte)
	require.NoError(t, err)
	assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF")))
	contenido := string(pdfBytes)
	assert.Contains(t, contenido, "Encabezado")
	assert.Contains(t, contenido, "Metricas")
	assert.Contains(t, contenido, "Historias")
	assert.Contains(t, contenido, "Defectos")
}

// ---------------------------------------------------------------------------
// Tests de slugify y nombre de archivo
// ---------------------------------------------------------------------------

func TestSlugify(t *testing.T) {
	t.Parallel()
	casos := []struct {
		entrada  string
		esperado string
	}{
		{"Proyecto Alpha", "proyecto-alpha"},
		{"Sprint 10", "sprint-10"},
		{"Mi Proyecto!", "mi-proyecto"},
		{"", "reporte"},
		{"   ", "reporte"},
		{"ABC 123", "abc-123"},
	}
	for _, c := range casos {
		t.Run(c.entrada, func(t *testing.T) {
			assert.Equal(t, c.esperado, slugify(c.entrada))
		})
	}
}

func TestNombreArchivoProyecto(t *testing.T) {
	t.Parallel()
	fecha := "2026-11-15"
	nombre := nombreArchivoProyecto("Proyecto Alpha", fecha)
	assert.True(t, strings.HasPrefix(nombre, "reporte-"), "debe empezar con reporte-")
	assert.True(t, strings.HasSuffix(nombre, ".pdf"), "debe terminar en .pdf")
	assert.Contains(t, nombre, "proyecto")
	assert.Contains(t, nombre, fecha)
}

func TestNombreArchivoSprint(t *testing.T) {
	t.Parallel()
	fecha := "2026-11-15"
	nombre := nombreArchivoSprint("Proyecto Alpha", "Sprint-10", fecha)
	assert.True(t, strings.HasPrefix(nombre, "reporte-"), "debe empezar con reporte-")
	assert.True(t, strings.HasSuffix(nombre, ".pdf"), "debe terminar en .pdf")
	assert.Contains(t, nombre, "sprint")
	assert.Contains(t, nombre, fecha)
}

// ---------------------------------------------------------------------------
// Tests del PDFService completo (con mocks de repositorios)
// ---------------------------------------------------------------------------

func TestPDFServiceGenerarPDFProyecto(t *testing.T) {
	t.Parallel()
	svc, proyectos, sprints, backlog, historias, defectos := nuevoPDFServiceConMocks(t)
	ctx := context.Background()

	now := time.Now()
	proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{
		ID: 1, Nombre: "Proyecto Test", FechaInicio: now, Estado: domain.EstadoActivo,
	}, nil).Times(2)
	sprints.On("ListarPorProyecto", mock.Anything, uint(1)).Return([]domain.Sprint{}, nil).Once()
	backlog.On("ListarPorProyecto", mock.Anything, uint(1)).Return([]domain.SprintBacklogItem{}, nil).Once()
	historias.On("ListarPorProyecto", mock.Anything, uint(1)).Return([]domain.Historia{}, nil).Times(2)
	defectos.On("ListarPorProyecto", mock.Anything, uint(1)).Return([]domain.Defecto{}, nil).Times(2)

	pdfBytes, filename, err := svc.GenerarPDFProyecto(ctx, 1)
	require.NoError(t, err)
	assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF")))
	assert.True(t, strings.HasSuffix(filename, ".pdf"))
	assert.Contains(t, filename, "proyecto")

	proyectos.AssertExpectations(t)
}

func TestPDFServiceGenerarPDFSprint(t *testing.T) {
	t.Parallel()
	svc, proyectos, sprints, backlog, historias, defectos := nuevoPDFServiceConMocks(t)
	ctx := context.Background()

	now := time.Now()
	proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{
		ID: 1, Nombre: "Proyecto Test", FechaInicio: now, Estado: domain.EstadoActivo,
	}, nil).Times(2)
	sprints.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(&domain.Sprint{
		ID: 10, ProyectoID: 1, Identificador: "Sprint-10",
		Estado: domain.EstadoSprintActivo, FechaInicio: now, FechaFin: now,
	}, nil).Times(2)
	backlog.On("ObtenerPorSprint", mock.Anything, uint(1), uint(10)).Return([]domain.SprintBacklogItem{}, nil).Times(2)
	historias.On("ObtenerPorIDs", mock.Anything, mock.Anything, mock.Anything).Return([]domain.Historia{}, nil).Maybe()
	defectos.On("ListarPorProyecto", mock.Anything, uint(1)).Return([]domain.Defecto{}, nil).Times(2)

	pdfBytes, filename, err := svc.GenerarPDFSprint(ctx, 1, 10)
	require.NoError(t, err)
	assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF")))
	assert.True(t, strings.HasSuffix(filename, ".pdf"))
	assert.Contains(t, filename, "sprint")

	proyectos.AssertExpectations(t)
}

func TestPDFServiceErrores(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("proyecto inexistente", func(t *testing.T) {
		t.Parallel()
		svc, proyectos, _, _, _, _ := nuevoPDFServiceConMocks(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(999)).Return(nil, domain.ErrProyectoNoEncontrado).Once()
		_, _, err := svc.GenerarPDFProyecto(ctx, 999)
		assert.ErrorIs(t, err, domain.ErrProyectoNoEncontrado)
	})

	t.Run("sprint inexistente", func(t *testing.T) {
		t.Parallel()
		svc, proyectos, sprints, _, _, _ := nuevoPDFServiceConMocks(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil).Once()
		sprints.On("ObtenerPorID", mock.Anything, uint(1), uint(999)).Return(nil, domain.ErrSprintNoEncontrado).Once()
		_, _, err := svc.GenerarPDFSprint(ctx, 1, 999)
		assert.ErrorIs(t, err, domain.ErrSprintNoEncontrado)
	})
}
