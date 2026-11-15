package handler

import (
	"context"

	"ingenieria-y-calidad/internal/domain"
)

type ServicioReporte interface {
	ObtenerReporteProyecto(ctx context.Context, proyectoID uint) (domain.ReporteProyecto, error)
	ObtenerReporteSprint(ctx context.Context, proyectoID, sprintID uint) (domain.ReporteSprint, error)
}
