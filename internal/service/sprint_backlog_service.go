package service

import (
	"context"
	"fmt"
	"time"

	"ingenieria-y-calidad/internal/domain"
)

// SprintBacklogService implementa las reglas de negocio para asignar y
// desasignar historias al Sprint Backlog.
type SprintBacklogService struct {
	proyectos      domain.ProyectoRepository
	sprints        domain.SprintRepository
	historias      domain.HistoriaRepository
	sprintBacklog  domain.SprintBacklogRepository
}

// NewSprintBacklogService construye el servicio con inyección de dependencias.
func NewSprintBacklogService(
	proyectos domain.ProyectoRepository,
	sprints domain.SprintRepository,
	historias domain.HistoriaRepository,
	sprintBacklog domain.SprintBacklogRepository,
) *SprintBacklogService {
	return &SprintBacklogService{
		proyectos:     proyectos,
		sprints:       sprints,
		historias:     historias,
		sprintBacklog: sprintBacklog,
	}
}

// AsignarASprint mueve una historia en estado Backlog al Sprint Backlog de un Sprint Activo.
func (s *SprintBacklogService) AsignarASprint(ctx context.Context, proyectoID, sprintID, historiaID uint) error {
	if err := s.exigirProyecto(ctx, proyectoID); err != nil {
		return err
	}

	sprint, err := s.sprints.ObtenerPorID(ctx, proyectoID, sprintID)
	if err != nil {
		return fmt.Errorf("obteniendo sprint: %w", err)
	}

	if sprint.Estado != domain.EstadoSprintActivo {
		return domain.ErrSprintNoActivo
	}

	historia, err := s.historias.ObtenerPorID(ctx, proyectoID, historiaID)
	if err != nil {
		return fmt.Errorf("obteniendo historia: %w", err)
	}

	if historia.Estado != domain.EstadoHistoriaBacklog {
		return domain.ErrHistoriaNoPendiente
	}

	asignacionExistente, err := s.sprintBacklog.ObtenerPorHistoria(ctx, historiaID)
	if err != nil {
		return fmt.Errorf("verificando asignación de historia: %w", err)
	}
	if asignacionExistente != nil {
		return domain.ErrHistoriaYaAsignada
	}

	ahora := time.Now()
	if ahora.Before(sprint.FechaInicio) || ahora.After(sprint.FechaFin) {
		return domain.ErrFueraDeVentanaTemporal
	}

	maxOrden, err := s.sprintBacklog.ObtenerMaxOrden(ctx, proyectoID, sprintID)
	if err != nil {
		return fmt.Errorf("obteniendo orden del sprint backlog: %w", err)
	}

	item := &domain.SprintBacklogItem{
		ProyectoID:      proyectoID,
		SprintID:        sprintID,
		HistoriaID:      historiaID,
		Orden:           maxOrden + 1,
		FechaAsignacion: time.Now().UTC(),
	}

	if err := s.sprintBacklog.Asignar(ctx, item); err != nil {
		return fmt.Errorf("asignando historia al sprint: %w", err)
	}

	return nil
}

// DesasignarDeSprint saca una historia del Sprint Backlog y la devuelve al Product Backlog.
func (s *SprintBacklogService) DesasignarDeSprint(ctx context.Context, proyectoID, sprintID, historiaID uint) error {
	if err := s.exigirProyecto(ctx, proyectoID); err != nil {
		return err
	}

	sprint, err := s.sprints.ObtenerPorID(ctx, proyectoID, sprintID)
	if err != nil {
		return fmt.Errorf("obteniendo sprint: %w", err)
	}

	if sprint.Estado == domain.EstadoSprintCerrado {
		return domain.ErrSprintCerrado
	}

	asignacion, err := s.sprintBacklog.ObtenerPorHistoria(ctx, historiaID)
	if err != nil {
		return fmt.Errorf("verificando asignación de historia: %w", err)
	}
	if asignacion == nil || asignacion.ProyectoID != proyectoID || asignacion.SprintID != sprintID {
		return domain.ErrHistoriaNoPerteneceASprint
	}

	if err := s.sprintBacklog.Desasignar(ctx, proyectoID, sprintID, historiaID); err != nil {
		return fmt.Errorf("desasignando historia del sprint: %w", err)
	}

	return nil
}

func (s *SprintBacklogService) exigirProyecto(ctx context.Context, proyectoID uint) error {
	if _, err := s.proyectos.ObtenerPorID(ctx, proyectoID); err != nil {
		return fmt.Errorf("obteniendo proyecto: %w", err)
	}
	return nil
}
