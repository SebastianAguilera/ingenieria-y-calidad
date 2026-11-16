package service

import (
	"context"
	"fmt"

	"ingenieria-y-calidad/internal/domain"
)

type ReporteService struct {
	proyectos     domain.ProyectoRepository
	sprints       domain.SprintRepository
	sprintBacklog domain.SprintBacklogRepository
	historias     domain.HistoriaRepository
	defectos      domain.DefectoRepository
	metricas      *MetricasService
}

func NewReporteService(
	proyectos domain.ProyectoRepository,
	sprints domain.SprintRepository,
	sprintBacklog domain.SprintBacklogRepository,
	historias domain.HistoriaRepository,
	defectos domain.DefectoRepository,
	metricas *MetricasService,
) *ReporteService {
	return &ReporteService{
		proyectos:     proyectos,
		sprints:       sprints,
		sprintBacklog: sprintBacklog,
		historias:     historias,
		defectos:      defectos,
		metricas:      metricas,
	}
}

func (s *ReporteService) ObtenerReporteProyecto(ctx context.Context, proyectoID uint) (domain.ReporteProyecto, error) {
	proyecto, err := s.proyectos.ObtenerPorID(ctx, proyectoID)
	if err != nil {
		return domain.ReporteProyecto{}, fmt.Errorf("obteniendo proyecto: %w", err)
	}

	metricas, err := s.metricas.CalcularPorProyecto(ctx, proyectoID)
	if err != nil {
		return domain.ReporteProyecto{}, fmt.Errorf("calculando métricas: %w", err)
	}

	historias, err := s.historias.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return domain.ReporteProyecto{}, fmt.Errorf("listando historias: %w", err)
	}
	if historias == nil {
		historias = []domain.Historia{}
	}

	defectos, err := s.defectos.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return domain.ReporteProyecto{}, fmt.Errorf("listando defectos: %w", err)
	}
	if defectos == nil {
		defectos = []domain.Defecto{}
	}

	return domain.ReporteProyecto{
		Base:        "estado_actual",
		ProyectoID:  proyecto.ID,
		Nombre:      proyecto.Nombre,
		Descripcion: proyecto.Descripcion,
		FechaInicio: proyecto.FechaInicio,
		FechaFin:    proyecto.FechaFin,
		Estado:      proyecto.Estado,
		Historias:   historias,
		Metricas:    metricas,
		Defectos:    defectos,
	}, nil
}

func (s *ReporteService) ObtenerReporteSprint(ctx context.Context, proyectoID, sprintID uint) (domain.ReporteSprint, error) {
	proyecto, err := s.proyectos.ObtenerPorID(ctx, proyectoID)
	if err != nil {
		return domain.ReporteSprint{}, fmt.Errorf("obteniendo proyecto: %w", err)
	}

	sprint, err := s.sprints.ObtenerPorID(ctx, proyectoID, sprintID)
	if err != nil {
		return domain.ReporteSprint{}, fmt.Errorf("obteniendo sprint: %w", err)
	}

	metricas, err := s.metricas.CalcularPorSprint(ctx, proyectoID, sprintID)
	if err != nil {
		return domain.ReporteSprint{}, fmt.Errorf("calculando métricas de sprint: %w", err)
	}

	asignaciones, err := s.sprintBacklog.ObtenerPorSprint(ctx, proyectoID, sprintID)
	if err != nil {
		return domain.ReporteSprint{}, fmt.Errorf("obteniendo sprint backlog: %w", err)
	}

	historias := []domain.Historia{}
	if len(asignaciones) > 0 {
		historiaIDs := make([]uint, 0, len(asignaciones))
		for _, a := range asignaciones {
			historiaIDs = append(historiaIDs, a.HistoriaID)
		}
		historias, err = s.historias.ObtenerPorIDs(ctx, proyectoID, historiaIDs)
		if err != nil {
			return domain.ReporteSprint{}, fmt.Errorf("obteniendo historias del sprint: %w", err)
		}
		if historias == nil {
			historias = []domain.Historia{}
		}
	}

	todosDefectos, err := s.defectos.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return domain.ReporteSprint{}, fmt.Errorf("listando defectos: %w", err)
	}

	defectosSprint := filtrarDefectosSprint(todosDefectos, sprintID)
	if defectosSprint == nil {
		defectosSprint = []domain.Defecto{}
	}

	_ = proyecto
	return domain.ReporteSprint{
		Base:          "estado_actual",
		ProyectoID:    proyecto.ID,
		SprintID:      sprint.ID,
		Identificador: sprint.Identificador,
		SprintGoal:    sprint.SprintGoal,
		FechaInicio:   sprint.FechaInicio,
		FechaFin:      sprint.FechaFin,
		Estado:        sprint.Estado,
		Historias:     historias,
		Metricas:      metricas,
		Defectos:      defectosSprint,
	}, nil
}

func filtrarDefectosSprint(defectos []domain.Defecto, sprintID uint) []domain.Defecto {
	filtrados := make([]domain.Defecto, 0)
	for _, d := range defectos {
		detectado := d.SprintDeteccionID == sprintID
		resueltoEnSprint := d.SprintResolucionID != nil && *d.SprintResolucionID == sprintID && d.Estado.EsDeResolucion()
		if detectado || resueltoEnSprint {
			filtrados = append(filtrados, d)
		}
	}
	return filtrados
}
