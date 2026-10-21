package domain

import (
	"context"
	"time"
)

// SprintRepository define las operaciones de persistencia que el service de
// Sprints necesita. La implementacion vive en internal/repository.
type SprintRepository interface {
	// Crear persiste un Sprint nuevo dentro del proyecto indicado. Cuando el
	// identificador ya fue tomado en el mismo proyecto, el indice unico
	// aborta el INSERT y el error devuelto es ErrIdentificadorSprintDuplicado.
	Crear(ctx context.Context, s *Sprint) error

	// ObtenerPorID busca un Sprint dentro de un proyecto concreto. Un Sprint
	// de otro proyecto se comporta como inexistente.
	ObtenerPorID(ctx context.Context, proyectoID, id uint) (*Sprint, error)

	// ListarPorProyecto devuelve los Sprints del proyecto ordenados por
	// fecha_inicio ASC con desempate por id ASC. Cuando no hay Sprints
	// devuelve un slice vacio, nunca nil.
	ListarPorProyecto(ctx context.Context, proyectoID uint) ([]Sprint, error)

	// ActualizarEstado escribe el estado del Sprint y refresca la marca de
	// auditoria.
	ActualizarEstado(ctx context.Context, s *Sprint) error

	// ObtenerActivo devuelve el Sprint en estado Activo del proyecto, o
	// (nil, nil) si el proyecto no tiene ningun Sprint activo.
	ObtenerActivo(ctx context.Context, proyectoID uint) (*Sprint, error)

	// HaySolapamiento responde si existe un Sprint del proyecto cuyo rango de
	// fechas se interseca estrictamente con [inicio, fin]: los Sprints
	// contiguos (la fin de uno es igual a la inicio del siguiente) no se
	// consideran solapados.
	HaySolapamiento(ctx context.Context, proyectoID uint, inicio, fin time.Time) (bool, error)
}
