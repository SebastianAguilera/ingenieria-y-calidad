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
	proyectos     domain.ProyectoRepository
	sprints       domain.SprintRepository
	historias     domain.HistoriaRepository
	sprintBacklog domain.SprintBacklogRepository
	// ahora devuelve la hora actual; los tests la reemplazan para fijar el reloj.
	ahora func() time.Time
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
		ahora:         time.Now,
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

	ahora := s.ahora()
	if !dentroDeLaVentana(ahora, sprint.FechaInicio, sprint.FechaFin) {
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
		FechaAsignacion: ahora.UTC(),
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

// ListarSprintBacklog devuelve el Sprint solicitado junto a sus historias
// asignadas, ordenadas por el repositorio (orden ASC, id ASC). Las historias se
// resuelven en una sola consulta por lote para evitar el patron N+1.
func (s *SprintBacklogService) ListarSprintBacklog(ctx context.Context, proyectoID, sprintID uint) (*domain.SprintBacklogDetalle, error) {
	if err := s.exigirProyecto(ctx, proyectoID); err != nil {
		return nil, err
	}

	sprint, err := s.sprints.ObtenerPorID(ctx, proyectoID, sprintID)
	if err != nil {
		return nil, fmt.Errorf("obteniendo sprint: %w", err)
	}

	items, err := s.sprintBacklog.ObtenerPorSprint(ctx, proyectoID, sprintID)
	if err != nil {
		return nil, fmt.Errorf("listando sprint backlog: %w", err)
	}

	detalle := &domain.SprintBacklogDetalle{
		Sprint: sprint,
		Items:  make([]domain.SprintBacklogItemDetalle, 0, len(items)),
	}
	if len(items) == 0 {
		return detalle, nil
	}

	ids := make([]uint, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.HistoriaID)
	}

	historias, err := s.historias.ObtenerPorIDs(ctx, proyectoID, ids)
	if err != nil {
		return nil, fmt.Errorf("obteniendo historias del sprint backlog: %w", err)
	}

	porID := make(map[uint]domain.Historia, len(historias))
	for _, historia := range historias {
		porID[historia.ID] = historia
	}

	for _, item := range items {
		historia, existe := porID[item.HistoriaID]
		if !existe {
			return nil, fmt.Errorf("historia %d del sprint backlog no encontrada: %w",
				item.HistoriaID, domain.ErrHistoriaNoEncontrada)
		}
		detalle.Items = append(detalle.Items, domain.SprintBacklogItemDetalle{
			Item:     item,
			Historia: historia,
		})
	}

	return detalle, nil
}

func (s *SprintBacklogService) exigirProyecto(ctx context.Context, proyectoID uint) error {
	if _, err := s.proyectos.ObtenerPorID(ctx, proyectoID); err != nil {
		return fmt.Errorf("obteniendo proyecto: %w", err)
	}
	return nil
}

// zonaDelEquipo es la hora de Argentina (UTC−3, sin horario de verano). Se fija
// en el código porque el servidor de producción (Render) corre en UTC y su
// imagen Alpine no trae la base de zonas horarias.
var zonaDelEquipo = time.FixedZone("ART", -3*60*60)

// dentroDeLaVentana aplica RN-04 por día calendario: el sprint abarca completos
// su primer y su último día. Las fechas del sprint son días ("AAAA-MM-DD")
// guardados a las 00:00 UTC, así que se leen en UTC; el día de hoy se toma en
// la hora del equipo, sin importar la zona horaria del servidor.
func dentroDeLaVentana(ahora, inicio, fin time.Time) bool {
	hoy := diaCalendario(ahora.In(zonaDelEquipo))
	return !hoy.Before(diaCalendario(inicio.UTC())) && !hoy.After(diaCalendario(fin.UTC()))
}

// diaCalendario descarta la hora y la zona: conserva solo año, mes y día.
func diaCalendario(t time.Time) time.Time {
	anio, mes, dia := t.Date()
	return time.Date(anio, mes, dia, 0, 0, 0, 0, time.UTC)
}
