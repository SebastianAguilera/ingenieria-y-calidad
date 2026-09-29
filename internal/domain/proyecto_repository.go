package domain

import "context"

// ProyectoRepository es el puerto de persistencia de proyectos.
// Lo implementa internal/repository con GORM.
type ProyectoRepository interface {
	// CrearConIntegrantes inserta el proyecto y todos sus vínculos en una
	// única transacción: o se persiste todo, o no se persiste nada.
	CrearConIntegrantes(ctx context.Context, p *Proyecto, integrantes []Integrante) error
	ObtenerPorID(ctx context.Context, id uint) (*Proyecto, error)
	// Listar devuelve los proyectos no dados de baja, con integrantes precargados.
	Listar(ctx context.Context) ([]Proyecto, error)
	Actualizar(ctx context.Context, p *Proyecto) error
	// Eliminar realiza la baja lógica: informa la hora en borrado_en.
	Eliminar(ctx context.Context, id uint) error
	// TieneHistorial informa si el proyecto ya tiene entidades dependientes.
	TieneHistorial(ctx context.Context, id uint) (bool, error)
}

// IntegranteRepository es el puerto de persistencia de integrantes.
type IntegranteRepository interface {
	ObtenerPorEmail(ctx context.Context, email string) (*Integrante, error)
	Crear(ctx context.Context, i *Integrante) error
	EstaAsociado(ctx context.Context, proyectoID, integranteID uint) (bool, error)
	// Vincular agrega el vínculo; si el vínculo ya existe devuelve
	// domain.ErrIntegranteYaAsociado (la PK compuesta es la red de seguridad).
	Vincular(ctx context.Context, proyectoID, integranteID uint) error
	Desvincular(ctx context.Context, proyectoID, integranteID uint) error
	ListarPorProyecto(ctx context.Context, proyectoID uint) ([]Integrante, error)
	// ActualizarNombre refresca el nombre de un integrante preexistente.
	ActualizarNombre(ctx context.Context, id uint, nombre string) error
}
