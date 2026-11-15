package domain

// SprintBacklogDetalle es el read model del Sprint Backlog: el Sprint
// consultado junto a sus items enriquecidos con la historia anidada.
type SprintBacklogDetalle struct {
	Sprint *Sprint
	Items  []SprintBacklogItemDetalle
}

// SprintBacklogItemDetalle combina la asignación al Sprint con la historia de
// usuario referenciada, sin modificar el modelo SprintBacklogItem.
type SprintBacklogItemDetalle struct {
	Item     SprintBacklogItem
	Historia Historia
}
