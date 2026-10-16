package domain

import "context"

// HistoriaRepository es el puerto de persistencia de las historias del Product
// Backlog. Lo implementa internal/repository con GORM.
type HistoriaRepository interface {
	// Crear inserta una historia y escribe en ella el identificador y las
	// marcas de auditoría generados por la base.
	Crear(ctx context.Context, h *Historia) error
	// SiguienteNumero devuelve el número consecutivo que corresponde a la
	// próxima historia del proyecto.
	SiguienteNumero(ctx context.Context, proyectoID uint) (int, error)
	// ObtenerPorID busca una historia dentro de un proyecto concreto.
	ObtenerPorID(ctx context.Context, proyectoID, id uint) (*Historia, error)
	// ListarPorProyecto devuelve el backlog de un proyecto en orden de alta.
	ListarPorProyecto(ctx context.Context, proyectoID uint) ([]Historia, error)
	// Actualizar escribe título, descripción, prioridad y estado de la historia.
	Actualizar(ctx context.Context, h *Historia) error
}
