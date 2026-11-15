package service

import (
	"context"
	"fmt"
	"strings"

	"ingenieria-y-calidad/internal/domain"
)

// SprintService implementa las reglas de negocio de los Sprints: alta con
// validacion de identificador, goal y fechas, listado por proyecto y la
// maquina de estados Planificado -> Activo -> Cerrado con un unico Sprint
// activo por proyecto. Depende unicamente de las interfaces de internal/domain.
type SprintService struct {
	proyectos domain.ProyectoRepository
	sprints   domain.SprintRepository
}

// NewSprintService construye el servicio de Sprints con inyeccion de
// dependencias.
func NewSprintService(proyectos domain.ProyectoRepository, sprints domain.SprintRepository) *SprintService {
	return &SprintService{
		proyectos: proyectos,
		sprints:   sprints,
	}
}

// Crear valida el alta, verifica que las fechas no se solapen con otros
// Sprints del proyecto y persiste el Sprint en estado Planificado. El estado
// no se acepta del cliente (RB-08).
func (s *SprintService) Crear(ctx context.Context, proyectoID uint, nuevo domain.NuevoSprint) (*domain.Sprint, error) {
	identificador, sprintGoal, err := validarAltaSprint(nuevo)
	if err != nil {
		return nil, err
	}

	if err := s.validarFechas(nuevo); err != nil {
		return nil, err
	}

	if err := s.exigirProyecto(ctx, proyectoID); err != nil {
		return nil, err
	}

	solapa, err := s.sprints.HaySolapamiento(ctx, proyectoID, nuevo.FechaInicio, nuevo.FechaFin)
	if err != nil {
		return nil, fmt.Errorf("verificando el solapamiento de fechas: %w", err)
	}
	if solapa {
		return nil, domain.ErrSprintFechasSolapadas
	}

	sprint := &domain.Sprint{
		ProyectoID:    proyectoID,
		Identificador: identificador,
		SprintGoal:    sprintGoal,
		Estado:        domain.EstadoSprintPlanificado,
		FechaInicio:   nuevo.FechaInicio,
		FechaFin:      nuevo.FechaFin,
	}

	if err := s.sprints.Crear(ctx, sprint); err != nil {
		return nil, fmt.Errorf("creando sprint: %w", err)
	}

	return sprint, nil
}

// Listar devuelve los Sprints de un proyecto ordenados por fecha de inicio.
func (s *SprintService) Listar(ctx context.Context, proyectoID uint) ([]domain.Sprint, error) {
	if err := s.exigirProyecto(ctx, proyectoID); err != nil {
		return nil, err
	}

	sprints, err := s.sprints.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("listando sprints: %w", err)
	}
	if sprints == nil {
		return []domain.Sprint{}, nil
	}
	return sprints, nil
}

// CambiarEstado aplica una transicion de la maquina de estados
// Planificado -> Activo -> Cerrado. Al activar un Sprint se exige que el
// proyecto no tenga otro Sprint Activo (RB-06).
func (s *SprintService) CambiarEstado(ctx context.Context, proyectoID, sprintID uint, destino domain.EstadoSprint) (*domain.Sprint, error) {
	if !destino.Valido() {
		return nil, domain.ErrEstadoSprintInvalido
	}

	if err := s.exigirProyecto(ctx, proyectoID); err != nil {
		return nil, err
	}

	sprint, err := s.sprints.ObtenerPorID(ctx, proyectoID, sprintID)
	if err != nil {
		return nil, fmt.Errorf("obteniendo sprint: %w", err)
	}

	if !sprint.Estado.TransicionValidaA(destino) {
		return nil, domain.ErrTransicionSprintInvalida
	}

	if destino == domain.EstadoSprintActivo {
		activo, err := s.sprints.ObtenerActivo(ctx, proyectoID)
		if err != nil {
			return nil, fmt.Errorf("verificando el sprint activo: %w", err)
		}
		if activo != nil {
			return nil, domain.ErrSprintActivoExistente
		}
	}

	sprint.Estado = destino
	if err := s.sprints.ActualizarEstado(ctx, sprint); err != nil {
		return nil, fmt.Errorf("actualizando el estado del sprint: %w", err)
	}

	return sprint, nil
}

// exigirProyecto confirma que el proyecto existe antes de operar sobre sus
// Sprints.
func (s *SprintService) exigirProyecto(ctx context.Context, proyectoID uint) error {
	if _, err := s.proyectos.ObtenerPorID(ctx, proyectoID); err != nil {
		return fmt.Errorf("obteniendo proyecto: %w", err)
	}
	return nil
}

// validarFechas ejecuta RB-03: la fecha de inicio es obligatoria y la de fin
// no puede ser anterior.
func (s *SprintService) validarFechas(nuevo domain.NuevoSprint) error {
	if nuevo.FechaInicio.IsZero() {
		return domain.ErrSprintFechaInicioObligatoria
	}
	if nuevo.FechaFin.Before(nuevo.FechaInicio) {
		return domain.ErrSprintFechaFinAnteriorAlInicio
	}
	return nil
}

// validarAltaSprint ejecuta RB-01 y RB-02 y normaliza los textos con
// TrimSpace antes de tocar el repositorio.
func validarAltaSprint(nuevo domain.NuevoSprint) (string, string, error) {
	identificador := strings.TrimSpace(nuevo.Identificador)
	if len(identificador) == 0 {
		return "", "", domain.ErrIdentificadorSprintObligatorio
	}
	if len([]rune(identificador)) > 50 {
		return "", "", domain.ErrIdentificadorSprintDemasiadoLargo
	}

	sprintGoal := strings.TrimSpace(nuevo.SprintGoal)
	if len(sprintGoal) == 0 {
		return "", "", domain.ErrSprintGoalObligatorio
	}
	if len([]rune(sprintGoal)) > 300 {
		return "", "", domain.ErrSprintGoalDemasiadoLargo
	}

	return identificador, sprintGoal, nil
}
