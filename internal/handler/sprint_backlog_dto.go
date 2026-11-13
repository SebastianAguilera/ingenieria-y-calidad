package handler

import "ingenieria-y-calidad/internal/domain"

type asignarHistoriaSprintDTO struct {
	HistoriaID uint `json:"historia_id" binding:"required"`
}

// respuestaSprintBacklogItem representa una asignación con la historia anidada.
type respuestaSprintBacklogItem struct {
	ID              uint              `json:"id"`
	Orden           int               `json:"orden"`
	FechaAsignacion string            `json:"fecha_asignacion"`
	Historia        respuestaHistoria `json:"historia"`
}

// respuestaSprintBacklog es el read model de US-13: el Sprint consultado y sus
// items ordenados por orden ASC, id ASC.
type respuestaSprintBacklog struct {
	Sprint respuestaSprint              `json:"sprint"`
	Total  int                          `json:"total"`
	Items  []respuestaSprintBacklogItem `json:"items"`
}

func construirRespuestaSprintBacklog(detalle *domain.SprintBacklogDetalle) respuestaSprintBacklog {
	items := make([]respuestaSprintBacklogItem, 0, len(detalle.Items))
	for _, item := range detalle.Items {
		items = append(items, respuestaSprintBacklogItem{
			ID:              item.Item.ID,
			Orden:           item.Item.Orden,
			FechaAsignacion: formatearTimestamp(item.Item.FechaAsignacion),
			Historia:        construirRespuestaHistoria(&item.Historia),
		})
	}

	return respuestaSprintBacklog{
		Sprint: construirRespuestaSprint(detalle.Sprint),
		Total:  len(items),
		Items:  items,
	}
}
