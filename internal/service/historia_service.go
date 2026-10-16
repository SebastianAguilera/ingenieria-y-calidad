package service

import (
	"context"
	"fmt"
	"strings"

	"ingenieria-y-calidad/internal/domain"
)

// HistoriaService implementa las reglas de negocio del Product Backlog. Depende
// únicamente de las interfaces definidas en internal/domain.
type HistoriaService struct {
	proyectos domain.ProyectoRepository
	historias domain.HistoriaRepository
}

// NewHistoriaService construye el servicio de historias con inyección de
// dependencias.
func NewHistoriaService(proyectos domain.ProyectoRepository, historias domain.HistoriaRepository) *HistoriaService {
	return &HistoriaService{
		proyectos: proyectos,
		historias: historias,
	}
}

// Crear valida el alta, asigna el código consecutivo US-NN del proyecto y
// persiste la historia. La prioridad y el estado por defecto son Media y
// Backlog respectivamente.
func (s *HistoriaService) Crear(ctx context.Context, proyectoID uint, nuevo domain.NuevaHistoria) (*domain.Historia, error) {
	prioridad, estado, err := validarAltaHistoria(nuevo)
	if err != nil {
		return nil, err
	}

	if err := s.exigirProyecto(ctx, proyectoID); err != nil {
		return nil, err
	}

	numero, err := s.historias.SiguienteNumero(ctx, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("calculando el consecutivo de la historia: %w", err)
	}

	historia := &domain.Historia{
		ProyectoID:  proyectoID,
		Numero:      numero,
		Codigo:      domain.FormatearCodigoHistoria(numero),
		Titulo:      strings.TrimSpace(nuevo.Titulo),
		Descripcion: nuevo.Descripcion,
		Prioridad:   prioridad,
		Estado:      estado,
	}

	if err := s.historias.Crear(ctx, historia); err != nil {
		return nil, fmt.Errorf("creando historia: %w", err)
	}

	return historia, nil
}

// Listar devuelve el backlog completo de un proyecto en orden de alta.
func (s *HistoriaService) Listar(ctx context.Context, proyectoID uint) ([]domain.Historia, error) {
	if err := s.exigirProyecto(ctx, proyectoID); err != nil {
		return nil, err
	}

	historias, err := s.historias.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("listando historias: %w", err)
	}
	if historias == nil {
		return []domain.Historia{}, nil
	}
	return historias, nil
}

// Actualizar aplica una modificación parcial sobre una historia existente.
// Únicamente los campos informados se modifican.
func (s *HistoriaService) Actualizar(ctx context.Context, proyectoID, historiaID uint, actualizacion domain.ActualizacionHistoria) (*domain.Historia, error) {
	if err := validarActualizacionHistoria(actualizacion); err != nil {
		return nil, err
	}

	if err := s.exigirProyecto(ctx, proyectoID); err != nil {
		return nil, err
	}

	historia, err := s.historias.ObtenerPorID(ctx, proyectoID, historiaID)
	if err != nil {
		return nil, fmt.Errorf("obteniendo historia: %w", err)
	}

	if actualizacion.Titulo != nil {
		historia.Titulo = strings.TrimSpace(*actualizacion.Titulo)
	}
	if actualizacion.Descripcion != nil {
		historia.Descripcion = *actualizacion.Descripcion
	}
	if actualizacion.Prioridad != nil {
		historia.Prioridad = *actualizacion.Prioridad
	}
	if actualizacion.Estado != nil {
		historia.Estado = *actualizacion.Estado
	}

	if err := s.historias.Actualizar(ctx, historia); err != nil {
		return nil, fmt.Errorf("actualizando historia: %w", err)
	}

	return historia, nil
}

// exigirProyecto confirma que el proyecto existe antes de operar sobre su
// backlog.
func (s *HistoriaService) exigirProyecto(ctx context.Context, proyectoID uint) error {
	if _, err := s.proyectos.ObtenerPorID(ctx, proyectoID); err != nil {
		return fmt.Errorf("obteniendo proyecto: %w", err)
	}
	return nil
}

// validarAltaHistoria ejecuta RB-01 a RB-05 y resuelve los valores por defecto
// de prioridad y estado antes de tocar el repositorio.
func validarAltaHistoria(nuevo domain.NuevaHistoria) (domain.PrioridadHistoria, domain.EstadoHistoria, error) {
	if err := validarTituloHistoria(nuevo.Titulo); err != nil {
		return "", "", err
	}

	prioridad := nuevo.Prioridad
	if prioridad == "" {
		prioridad = domain.PrioridadMedia
	}
	if !prioridad.Valido() {
		return "", "", domain.ErrPrioridadHistoriaInvalida
	}

	estado := nuevo.Estado
	if estado == "" {
		estado = domain.EstadoHistoriaBacklog
	}
	if !estado.Valido() {
		return "", "", domain.ErrEstadoHistoriaInvalido
	}

	return prioridad, estado, nil
}

// validarActualizacionHistoria exige al menos un campo y valida los que se
// informan.
func validarActualizacionHistoria(actualizacion domain.ActualizacionHistoria) error {
	if actualizacion.Titulo == nil && actualizacion.Descripcion == nil &&
		actualizacion.Prioridad == nil && actualizacion.Estado == nil {
		return domain.ErrActualizacionHistoriaVacia
	}

	if actualizacion.Titulo != nil {
		if err := validarTituloHistoria(*actualizacion.Titulo); err != nil {
			return err
		}
	}
	if actualizacion.Prioridad != nil && !actualizacion.Prioridad.Valido() {
		return domain.ErrPrioridadHistoriaInvalida
	}
	if actualizacion.Estado != nil && !actualizacion.Estado.Valido() {
		return domain.ErrEstadoHistoriaInvalido
	}

	return nil
}

func validarTituloHistoria(titulo string) error {
	recortado := strings.TrimSpace(titulo)
	if len(recortado) == 0 {
		return domain.ErrTituloHistoriaObligatorio
	}
	if len([]rune(recortado)) > 200 {
		return domain.ErrTituloHistoriaDemasiadoLargo
	}
	return nil
}
