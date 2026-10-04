package domain

import "context"

// ProyectoRepository es el puerto de persistencia de proyectos.
// Lo implementa internal/repository con GORM.
type ProyectoRepository interface {
	// CrearConUsuarios inserta el proyecto y todos sus vínculos en una
	// única transacción: o se persiste todo, o no se persiste nada.
	CrearConUsuarios(ctx context.Context, p *Proyecto, usuarios []Usuario) error
	ObtenerPorID(ctx context.Context, id uint) (*Proyecto, error)
	// Listar devuelve los proyectos no dados de baja, con usuarios precargados.
	Listar(ctx context.Context) ([]Proyecto, error)
	Actualizar(ctx context.Context, p *Proyecto) error
	// Eliminar realiza la baja lógica: informa la hora en borrado_en.
	Eliminar(ctx context.Context, id uint) error
	// TieneHistorial informa si el proyecto ya tiene entidades dependientes.
	TieneHistorial(ctx context.Context, id uint) (bool, error)
}

// UsuarioRepository es el puerto de persistencia de usuarios.
type UsuarioRepository interface {
	ObtenerPorEmail(ctx context.Context, email string) (*Usuario, error)
	Crear(ctx context.Context, u *Usuario) error
	EstaAsociado(ctx context.Context, proyectoID, usuarioID uint) (bool, error)
	// Vincular agrega el vínculo; si el vínculo ya existe devuelve
	// domain.ErrUsuarioYaAsociado (la PK compuesta es la red de seguridad).
	Vincular(ctx context.Context, proyectoID, usuarioID uint) error
	Desvincular(ctx context.Context, proyectoID, usuarioID uint) error
	ListarPorProyecto(ctx context.Context, proyectoID uint) ([]Usuario, error)
	// ActualizarNombre refresca el nombre de un usuario preexistente.
	ActualizarNombre(ctx context.Context, id uint, nombre string) error
}
