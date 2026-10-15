package domain

import "context"

type ProyectoRepository interface {
	CrearConUsuarios(ctx context.Context, p *Proyecto, usuarios []Usuario) error
	ObtenerPorID(ctx context.Context, id uint) (*Proyecto, error)
	Listar(ctx context.Context) ([]Proyecto, error)
	Actualizar(ctx context.Context, p *Proyecto) error
	Eliminar(ctx context.Context, id uint) error
	TieneHistorial(ctx context.Context, id uint) (bool, error)
}

type UsuarioRepository interface {
	ObtenerPorEmail(ctx context.Context, email string) (*Usuario, error)
	Crear(ctx context.Context, u *Usuario) error
	EstaAsociado(ctx context.Context, proyectoID, usuarioID uint) (bool, error)

	Vincular(ctx context.Context, proyectoID, usuarioID uint) error
	Desvincular(ctx context.Context, proyectoID, usuarioID uint) error
	ListarPorProyecto(ctx context.Context, proyectoID uint) ([]Usuario, error)
	ActualizarNombre(ctx context.Context, id uint, nombre string) error
}
