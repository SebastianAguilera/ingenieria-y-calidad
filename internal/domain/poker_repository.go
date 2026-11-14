package domain

import "context"

// SesionPokerRepository es el puerto de persistencia de las sesiones de
// Planning Poker. Lo implementa internal/repository con GORM.
type SesionPokerRepository interface {
	// CrearSesion inserta la sesión y su ronda número 1 en una única
	// transacción. Si ya existe una sesión activa para la historia, el índice
	// único parcial aborta el INSERT y devuelve ErrSesionPokerActiva.
	CrearSesion(ctx context.Context, s *SesionPoker) error
	// ObtenerPorID busca una sesión dentro de un proyecto concreto. Una sesión
	// de otro proyecto se comporta como inexistente.
	ObtenerPorID(ctx context.Context, proyectoID, sesionID uint) (*SesionPoker, error)
	// ObtenerActivaPorHistoria devuelve la sesión Votando/Revelada de la
	// historia, o nil si no hay ninguna.
	ObtenerActivaPorHistoria(ctx context.Context, historiaID uint) (*SesionPoker, error)
	// ListarPorHistoria devuelve el historial completo de la historia,
	// incluidas Cerrada y Cancelada, en orden de creación descendente.
	ListarPorHistoria(ctx context.Context, historiaID uint) ([]SesionPoker, error)
	// ListarActivasPorProyecto devuelve las sesiones Votando/Revelada del
	// proyecto en orden de creación ascendente.
	ListarActivasPorProyecto(ctx context.Context, proyectoID uint) ([]SesionPoker, error)
	// ObtenerRonda devuelve la ronda con el número indicado de la sesión.
	ObtenerRonda(ctx context.Context, sesionID uint, numero int) (*RondaPoker, error)
	// ContarMiembros cuenta los vínculos del proyecto en proyecto_usuarios.
	ContarMiembros(ctx context.Context, proyectoID uint) (int, error)
	// ContarVotosPorRonda cuenta los votos registrados en la ronda.
	ContarVotosPorRonda(ctx context.Context, rondaID uint) (int, error)
	// ListarVotosPorRonda devuelve los votos de la ronda con el nombre del
	// votante.
	ListarVotosPorRonda(ctx context.Context, rondaID uint) ([]VotoPoker, error)
	// ObtenerVotoUsuario devuelve el voto del usuario en la ronda, o nil si
	// todavía no votó.
	ObtenerVotoUsuario(ctx context.Context, rondaID, usuarioID uint) (*VotoPoker, error)
	// RegistrarVoto inserta o actualiza el voto del usuario en la ronda actual
	// de la sesión. Bloquea la fila de la sesión (SELECT ... FOR UPDATE) y
	// verifica el estado dentro de la misma transacción: si la sesión no está
	// Votando devuelve ErrSesionPokerNoVotando sin escribir, y si no existe
	// devuelve ErrSesionPokerNoEncontrada.
	RegistrarVoto(ctx context.Context, sesionID, usuarioID uint, puntos float64) (*VotoPoker, error)
	// RevelarSesion pasa la sesión a Revelada y marca la ronda actual como
	// revelada. Bloquea la fila de la sesión y valida dentro de la transacción
	// que esté Votando y que la ronda tenga al menos un voto.
	RevelarSesion(ctx context.Context, sesionID uint) error
	// AbrirNuevaRonda crea la ronda siguiente y devuelve la sesión a Votando.
	// Solo se permite desde Revelada; bloquea la fila de la sesión durante la
	// transición.
	AbrirNuevaRonda(ctx context.Context, sesionID uint) (*RondaPoker, error)
	// ListarRondasPorSesion devuelve todas las rondas de la sesión ordenadas
	// por número ascendente.
	ListarRondasPorSesion(ctx context.Context, sesionID uint) ([]RondaPoker, error)
	// CerrarSesion registra el acuerdo (US-17): bloquea la fila de la sesión,
	// valida que esté Revelada (ValidarAcuerdoSesionPoker), fija
	// puntos_acordados, estado Cerrada y cerrado_en, y devuelve la ronda
	// actual para el story_estimate. Debe ejecutarse dentro de la
	// UnidadDeTrabajo del acuerdo para que el cierre sea atómico con la
	// actualización de la historia.
	CerrarSesion(ctx context.Context, sesionID uint, puntos float64) (*RondaPoker, error)
	// CancelarSesion bloquea la fila de la sesión, valida que esté Votando o
	// Revelada (ValidarCancelacionSesionPoker) y fija estado Cancelada y
	// cancelado_en. No toca votos, rondas ni la historia.
	CancelarSesion(ctx context.Context, sesionID uint) error
}

// StoryEstimateRepository es el puerto de persistencia de las estimaciones
// acordadas de las historias (tabla story_estimates).
type StoryEstimateRepository interface {
	// Crear inserta la estimación y escribe en ella el identificador generado.
	Crear(ctx context.Context, e *StoryEstimate) error
	// ListarPorSesiones devuelve las estimaciones originadas en las sesiones
	// indicadas. Con una lista vacía devuelve un slice vacío.
	ListarPorSesiones(ctx context.Context, sesionIDs []uint) ([]StoryEstimate, error)
}

// RepositoriosTransaccion expone los repositorios ligados a una misma
// transacción de base de datos.
type RepositoriosTransaccion interface {
	Historias() HistoriaRepository
	Poker() SesionPokerRepository
	Estimaciones() StoryEstimateRepository
}

// UnidadDeTrabajo ejecuta varias operaciones de persistencia de forma
// atómica (US-17 RB-06). Si fn devuelve un error (o entra en pánico) se hace
// rollback de todo lo escrito; si no, se confirma.
type UnidadDeTrabajo interface {
	Ejecutar(ctx context.Context, fn func(repos RepositoriosTransaccion) error) error
}
