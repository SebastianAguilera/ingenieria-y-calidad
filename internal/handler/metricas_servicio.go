package handler

import (
	"context"

	"ingenieria-y-calidad/internal/domain"
)

// ServicioMetricas es el puerto que el handler necesita del service de métricas.
// Solo expone operaciones de lectura (US-20, RN-01).
type ServicioMetricas interface {
	CalcularPorSprint(ctx context.Context, proyectoID, sprintID uint) (domain.MetricasSprint, error)
	CalcularPorProyecto(ctx context.Context, proyectoID uint) (domain.MetricasProyecto, error)
}
