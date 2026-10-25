package domain

import "context"

// SprintBacklogRepository es el puerto de persistencia para la asignación
// de historias al Sprint Backlog.
type SprintBacklogRepository interface {
	Asignar(ctx context.Context, item *SprintBacklogItem) error
	Desasignar(ctx context.Context, proyectoID, sprintID, historiaID uint) error
	ExisteAsignacionHistoria(ctx context.Context, historiaID uint) (bool, *SprintBacklogItem, error)
	ObtenerPorHistoria(ctx context.Context, historiaID uint) (*SprintBacklogItem, error)
	ObtenerPorSprint(ctx context.Context, proyectoID, sprintID uint) ([]SprintBacklogItem, error)
	ObtenerMaxOrden(ctx context.Context, proyectoID, sprintID uint) (int, error)
}
