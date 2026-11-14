package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"ingenieria-y-calidad/internal/domain"
)

// DefectoService implementa las reglas de negocio del registro y la gestión de
// defectos. Depende únicamente de las interfaces definidas en internal/domain.
type DefectoService struct {
	proyectos domain.ProyectoRepository
	historias domain.HistoriaRepository
	sprints   domain.SprintRepository
	usuarios  domain.UsuarioRepository
	defectos  domain.DefectoRepository
}

// NewDefectoService construye el servicio de defectos con inyección de
// dependencias.
func NewDefectoService(
	proyectos domain.ProyectoRepository,
	historias domain.HistoriaRepository,
	sprints domain.SprintRepository,
	usuarios domain.UsuarioRepository,
	defectos domain.DefectoRepository,
) *DefectoService {
	return &DefectoService{
		proyectos: proyectos,
		historias: historias,
		sprints:   sprints,
		usuarios:  usuarios,
		defectos:  defectos,
	}
}

// Crear valida el alta, verifica las referencias y persiste el defecto en
// estado abierto.
func (s *DefectoService) Crear(ctx context.Context, proyectoID uint, nuevo domain.NuevoDefecto) (*domain.Defecto, error) {
	descripcion := strings.TrimSpace(nuevo.Descripcion)
	if len(descripcion) == 0 {
		return nil, domain.ErrDescripcionDefectoObligatoria
	}
	if !nuevo.Severidad.Valida() {
		return nil, domain.ErrSeveridadDefectoInvalida
	}

	if err := s.exigirProyecto(ctx, proyectoID); err != nil {
		return nil, err
	}

	if _, err := s.historias.ObtenerPorID(ctx, proyectoID, nuevo.HistoriaID); err != nil {
		return nil, fmt.Errorf("obteniendo historia del defecto: %w", err)
	}

	if _, err := s.sprints.ObtenerPorID(ctx, proyectoID, nuevo.SprintDeteccionID); err != nil {
		return nil, fmt.Errorf("obteniendo sprint de detección: %w", err)
	}

	if nuevo.ReportadoPor != nil {
		if _, err := s.usuarios.ObtenerPorID(ctx, *nuevo.ReportadoPor); err != nil {
			return nil, fmt.Errorf("obteniendo usuario reportante: %w", err)
		}
	}

	defecto := &domain.Defecto{
		ProyectoID:        proyectoID,
		HistoriaID:        nuevo.HistoriaID,
		SprintDeteccionID: nuevo.SprintDeteccionID,
		ReportadoPor:      nuevo.ReportadoPor,
		Descripcion:       descripcion,
		Severidad:         nuevo.Severidad,
		Estado:            domain.EstadoDefectoAbierto,
	}

	if err := s.defectos.Crear(ctx, defecto); err != nil {
		return nil, fmt.Errorf("creando defecto: %w", err)
	}

	return defecto, nil
}

// Actualizar aplica una modificación parcial sobre un defecto existente. Si el
// estado resultante es de resolución exige el Sprint de resolución; si el
// defecto vuelve a un estado no resuelto, limpia la resolución.
func (s *DefectoService) Actualizar(ctx context.Context, proyectoID, defectoID uint, actualizacion domain.ActualizacionDefecto) (*domain.Defecto, error) {
	if actualizacion.Estado == nil && actualizacion.SprintResolucionID == nil &&
		actualizacion.Descripcion == nil && actualizacion.Severidad == nil &&
		!actualizacion.SprintDeteccionPresente {
		return nil, domain.ErrActualizacionDefectoVacia
	}
	if actualizacion.Estado != nil && !actualizacion.Estado.Valido() {
		return nil, domain.ErrEstadoDefectoInvalido
	}

	if err := s.exigirProyecto(ctx, proyectoID); err != nil {
		return nil, err
	}

	if actualizacion.SprintDeteccionPresente {
		return nil, domain.ErrSprintDeteccionNoModificable
	}

	descripcion, err := validarDescripcionActualizada(actualizacion.Descripcion)
	if err != nil {
		return nil, err
	}
	if actualizacion.Severidad != nil && !actualizacion.Severidad.Valida() {
		return nil, domain.ErrSeveridadDefectoInvalida
	}

	defecto, err := s.defectos.ObtenerPorID(ctx, proyectoID, defectoID)
	if err != nil {
		return nil, fmt.Errorf("obteniendo defecto: %w", err)
	}

	if descripcion != nil {
		defecto.Descripcion = *descripcion
	}
	if actualizacion.Severidad != nil {
		defecto.Severidad = *actualizacion.Severidad
	}

	estadoDestino := defecto.Estado
	if actualizacion.Estado != nil {
		estadoDestino = *actualizacion.Estado
	}

	if actualizacion.SprintResolucionID != nil {
		if !estadoDestino.EsDeResolucion() {
			return nil, domain.ErrSprintResolucionNoPermitido
		}
		if _, err := s.sprints.ObtenerPorID(ctx, proyectoID, *actualizacion.SprintResolucionID); err != nil {
			return nil, fmt.Errorf("obteniendo sprint de resolución: %w", err)
		}
		defecto.SprintResolucionID = actualizacion.SprintResolucionID
	} else if estadoDestino.EsDeResolucion() && defecto.SprintResolucionID == nil {
		return nil, domain.ErrSprintResolucionRequerido
	}

	if estadoDestino.EsDeResolucion() {
		if defecto.ResueltoEn == nil {
			ahora := time.Now().UTC()
			defecto.ResueltoEn = &ahora
		}
	} else {
		defecto.SprintResolucionID = nil
		defecto.ResueltoEn = nil
	}

	defecto.Estado = estadoDestino
	if err := s.defectos.Actualizar(ctx, defecto); err != nil {
		return nil, fmt.Errorf("actualizando defecto: %w", err)
	}

	return defecto, nil
}

// Listar devuelve los defectos de un proyecto en orden de alta.
func (s *DefectoService) Listar(ctx context.Context, proyectoID uint) ([]domain.Defecto, error) {
	if err := s.exigirProyecto(ctx, proyectoID); err != nil {
		return nil, err
	}

	defectos, err := s.defectos.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("listando defectos: %w", err)
	}
	if defectos == nil {
		return []domain.Defecto{}, nil
	}
	return defectos, nil
}

// ListarPorSprint devuelve los defectos del Sprint según el tipo de listado:
// por Sprint de detección o por Sprint de resolución (US-18). El tipo es
// obligatorio y se valida antes de consultar nada.
func (s *DefectoService) ListarPorSprint(ctx context.Context, proyectoID, sprintID uint, tipo domain.TipoListadoDefecto) ([]domain.Defecto, error) {
	if !tipo.Valido() {
		return nil, domain.ErrTipoListadoDefectoInvalido
	}
	if err := s.exigirProyecto(ctx, proyectoID); err != nil {
		return nil, err
	}
	if _, err := s.sprints.ObtenerPorID(ctx, proyectoID, sprintID); err != nil {
		return nil, fmt.Errorf("obteniendo sprint del listado: %w", err)
	}

	defectos, err := s.defectos.ListarPorSprint(ctx, proyectoID, sprintID, tipo)
	if err != nil {
		return nil, fmt.Errorf("listando defectos del sprint: %w", err)
	}
	if defectos == nil {
		return []domain.Defecto{}, nil
	}
	return defectos, nil
}

// limiteDescripcionDefecto es el máximo de caracteres (runas) de la descripción
// de un defecto al editarla (US-18).
const limiteDescripcionDefecto = 1000

// validarDescripcionActualizada normaliza la descripción informada en el PATCH
// (TrimSpace) y valida que no quede vacía ni supere el límite. Devuelve nil si
// no se informó.
func validarDescripcionActualizada(descripcion *string) (*string, error) {
	if descripcion == nil {
		return nil, nil
	}
	normalizada := strings.TrimSpace(*descripcion)
	if len(normalizada) == 0 {
		return nil, domain.ErrDescripcionDefectoObligatoria
	}
	if utf8.RuneCountInString(normalizada) > limiteDescripcionDefecto {
		return nil, domain.ErrDescripcionDefectoDemasiadoLarga
	}
	return &normalizada, nil
}

func (s *DefectoService) exigirProyecto(ctx context.Context, proyectoID uint) error {
	if _, err := s.proyectos.ObtenerPorID(ctx, proyectoID); err != nil {
		return fmt.Errorf("obteniendo proyecto: %w", err)
	}
	return nil
}
