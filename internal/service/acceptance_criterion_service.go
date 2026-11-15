package service

import (
	"context"
	"fmt"
	"strings"

	"ingenieria-y-calidad/internal/domain"
)

type AcceptanceCriterionService struct {
	proyectos domain.ProyectoRepository
	historias domain.HistoriaRepository
	criterios domain.AcceptanceCriterionRepository
}

func NewAcceptanceCriterionService(
	proyectos domain.ProyectoRepository,
	historias domain.HistoriaRepository,
	criterios domain.AcceptanceCriterionRepository,
) *AcceptanceCriterionService {
	return &AcceptanceCriterionService{
		proyectos: proyectos,
		historias: historias,
		criterios: criterios,
	}
}

func (s *AcceptanceCriterionService) Crear(ctx context.Context, proyectoID, historiaID uint, nuevo domain.NuevoAcceptanceCriterion) (*domain.AcceptanceCriterion, error) {
	if err := domain.ValidarDescripcionCriterio(nuevo.Descripcion); err != nil {
		return nil, err
	}
	if err := s.exigirHistoria(ctx, proyectoID, historiaID); err != nil {
		return nil, err
	}

	pos, err := s.criterios.SiguientePosicion(ctx, historiaID)
	if err != nil {
		return nil, fmt.Errorf("calculando posición: %w", err)
	}

	criterio := &domain.AcceptanceCriterion{
		HistoriaID:  historiaID,
		Descripcion: strings.TrimSpace(nuevo.Descripcion),
		Position:    pos,
		IsMet:       false,
	}

	if err := s.criterios.Crear(ctx, criterio); err != nil {
		return nil, fmt.Errorf("creando criterio de aceptación: %w", err)
	}

	return criterio, nil
}

func (s *AcceptanceCriterionService) Listar(ctx context.Context, proyectoID, historiaID uint) ([]domain.AcceptanceCriterion, error) {
	if err := s.exigirHistoria(ctx, proyectoID, historiaID); err != nil {
		return nil, err
	}

	criterios, err := s.criterios.ListarPorHistoria(ctx, historiaID)
	if err != nil {
		return nil, fmt.Errorf("listando criterios: %w", err)
	}
	if criterios == nil {
		return []domain.AcceptanceCriterion{}, nil
	}
	return criterios, nil
}

func (s *AcceptanceCriterionService) Actualizar(ctx context.Context, proyectoID, historiaID, criterioID uint, actualizacion domain.ActualizacionAcceptanceCriterion) (*domain.AcceptanceCriterion, error) {
	if actualizacion.Descripcion == nil && actualizacion.IsMet == nil {
		return nil, domain.ErrActualizacionCriterioVacia
	}
	if actualizacion.Descripcion != nil {
		if err := domain.ValidarDescripcionCriterio(*actualizacion.Descripcion); err != nil {
			return nil, err
		}
	}

	if err := s.exigirHistoria(ctx, proyectoID, historiaID); err != nil {
		return nil, err
	}

	criterio, err := s.criterios.ObtenerPorID(ctx, historiaID, criterioID)
	if err != nil {
		return nil, fmt.Errorf("obteniendo criterio: %w", err)
	}

	if actualizacion.Descripcion != nil {
		criterio.Descripcion = strings.TrimSpace(*actualizacion.Descripcion)
	}
	if actualizacion.IsMet != nil {
		criterio.IsMet = *actualizacion.IsMet
	}

	if err := s.criterios.Actualizar(ctx, criterio); err != nil {
		return nil, fmt.Errorf("actualizando criterio: %w", err)
	}

	return criterio, nil
}

func (s *AcceptanceCriterionService) Eliminar(ctx context.Context, proyectoID, historiaID, criterioID uint) error {
	if err := s.exigirHistoria(ctx, proyectoID, historiaID); err != nil {
		return err
	}
	if err := s.criterios.Eliminar(ctx, historiaID, criterioID); err != nil {
		return fmt.Errorf("eliminando criterio: %w", err)
	}
	return nil
}

func (s *AcceptanceCriterionService) exigirHistoria(ctx context.Context, proyectoID, historiaID uint) error {
	if _, err := s.proyectos.ObtenerPorID(ctx, proyectoID); err != nil {
		return fmt.Errorf("obteniendo proyecto: %w", err)
	}
	if _, err := s.historias.ObtenerPorID(ctx, proyectoID, historiaID); err != nil {
		return fmt.Errorf("obteniendo historia: %w", err)
	}
	return nil
}
