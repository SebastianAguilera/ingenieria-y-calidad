package handler

import (
	"context"

	"ingenieria-y-calidad/internal/domain"
)

// ServicioSprintBacklog es el puerto que el handler necesita del servicio de Sprint Backlog.
type ServicioSprintBacklog interface {
	AsignarASprint(ctx context.Context, proyectoID, sprintID, historiaID uint) error
	DesasignarDeSprint(ctx context.Context, proyectoID, sprintID, historiaID uint) error
	ListarSprintBacklog(ctx context.Context, proyectoID, sprintID uint) (*domain.SprintBacklogDetalle, error)
}
