package service

import (
	"context"
	"fmt"

	"ingenieria-y-calidad/internal/domain"
)

// MetricasService calcula las métricas de Sprint y de proyecto a partir del
// estado actual de historias, Sprints, asignaciones al Sprint Backlog y defectos.
type MetricasService struct {
	proyectos     domain.ProyectoRepository
	sprints       domain.SprintRepository
	sprintBacklog domain.SprintBacklogRepository
	historias     domain.HistoriaRepository
	defectos      domain.DefectoRepository
}

// NewMetricasService construye el servicio de métricas con inyección de
// dependencias.
func NewMetricasService(
	proyectos domain.ProyectoRepository,
	sprints domain.SprintRepository,
	sprintBacklog domain.SprintBacklogRepository,
	historias domain.HistoriaRepository,
	defectos domain.DefectoRepository,
) *MetricasService {
	return &MetricasService{
		proyectos:     proyectos,
		sprints:       sprints,
		sprintBacklog: sprintBacklog,
		historias:     historias,
		defectos:      defectos,
	}
}

// CalcularPorSprint calcula las métricas actuales de un Sprint perteneciente al
// proyecto indicado.
func (s *MetricasService) CalcularPorSprint(ctx context.Context, proyectoID, sprintID uint) (domain.MetricasSprint, error) {
	if _, err := s.proyectos.ObtenerPorID(ctx, proyectoID); err != nil {
		return domain.MetricasSprint{}, fmt.Errorf("obteniendo proyecto: %w", err)
	}

	if _, err := s.sprints.ObtenerPorID(ctx, proyectoID, sprintID); err != nil {
		return domain.MetricasSprint{}, fmt.Errorf("obteniendo sprint: %w", err)
	}

	asignaciones, err := s.sprintBacklog.ObtenerPorSprint(ctx, proyectoID, sprintID)
	if err != nil {
		return domain.MetricasSprint{}, fmt.Errorf("obteniendo sprint backlog: %w", err)
	}

	historias := make([]domain.Historia, 0, len(asignaciones))
	if len(asignaciones) > 0 {
		historiaIDs := make([]uint, 0, len(asignaciones))
		for _, a := range asignaciones {
			historiaIDs = append(historiaIDs, a.HistoriaID)
		}

		historias, err = s.historias.ObtenerPorIDs(ctx, proyectoID, historiaIDs)
		if err != nil {
			return domain.MetricasSprint{}, fmt.Errorf("obteniendo historias del sprint: %w", err)
		}
	}

	defectos, err := s.defectos.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return domain.MetricasSprint{}, fmt.Errorf("listando defectos del proyecto: %w", err)
	}

	return domain.CalcularMetricasSprint(historias, defectos, sprintID), nil
}

// CalcularPorProyecto calcula las métricas actuales agregadas de un proyecto.
func (s *MetricasService) CalcularPorProyecto(ctx context.Context, proyectoID uint) (domain.MetricasProyecto, error) {
	if _, err := s.proyectos.ObtenerPorID(ctx, proyectoID); err != nil {
		return domain.MetricasProyecto{}, fmt.Errorf("obteniendo proyecto: %w", err)
	}

	sprints, err := s.sprints.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return domain.MetricasProyecto{}, fmt.Errorf("listando sprints del proyecto: %w", err)
	}

	asignaciones, err := s.sprintBacklog.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return domain.MetricasProyecto{}, fmt.Errorf("listando sprint backlog del proyecto: %w", err)
	}

	historias, err := s.historias.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return domain.MetricasProyecto{}, fmt.Errorf("listando historias del proyecto: %w", err)
	}

	defectos, err := s.defectos.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return domain.MetricasProyecto{}, fmt.Errorf("listando defectos del proyecto: %w", err)
	}

	return domain.CalcularMetricasProyecto(sprints, asignaciones, historias, defectos), nil
}
