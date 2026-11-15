package domain

import "context"

type AcceptanceCriterionRepository interface {
	Crear(ctx context.Context, c *AcceptanceCriterion) error
	ListarPorHistoria(ctx context.Context, historiaID uint) ([]AcceptanceCriterion, error)
	ObtenerPorID(ctx context.Context, historiaID, id uint) (*AcceptanceCriterion, error)
	Actualizar(ctx context.Context, c *AcceptanceCriterion) error
	Eliminar(ctx context.Context, historiaID, id uint) error
	SiguientePosicion(ctx context.Context, historiaID uint) (int, error)
}
