package handler

import (
	"context"
)

// ServicioSprintBacklog es el puerto que el handler necesita del servicio de Sprint Backlog.
type ServicioSprintBacklog interface {
	AsignarASprint(ctx context.Context, proyectoID, sprintID, historiaID uint) error
	DesasignarDeSprint(ctx context.Context, proyectoID, sprintID, historiaID uint) error
}
