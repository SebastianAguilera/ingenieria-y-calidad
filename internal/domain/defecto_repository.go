package domain

import "context"

// DefectoRepository es el puerto de persistencia de los defectos. Lo
// implementa internal/repository con GORM.
type DefectoRepository interface {
	// Crear inserta un defecto y escribe en él el identificador y las marcas
	// de auditoría generadas por la base.
	Crear(ctx context.Context, d *Defecto) error
	// ObtenerPorID busca un defecto dentro de un proyecto concreto. Un defecto
	// de otro proyecto se comporta como inexistente.
	ObtenerPorID(ctx context.Context, proyectoID, id uint) (*Defecto, error)
	// Actualizar escribe el estado, el Sprint de resolución y la fecha de
	// resolución del defecto.
	Actualizar(ctx context.Context, d *Defecto) error
	// ListarPorProyecto devuelve los defectos del proyecto en orden de alta.
	// Cuando no hay defectos devuelve un slice vacío, nunca nil.
	ListarPorProyecto(ctx context.Context, proyectoID uint) ([]Defecto, error)
	// ListarPorSprint devuelve los defectos del Sprint según el tipo: por
	// sprint de detección o por sprint de resolución. Nunca devuelve nil.
	ListarPorSprint(ctx context.Context, proyectoID, sprintID uint, tipo TipoListadoDefecto) ([]Defecto, error)
}
