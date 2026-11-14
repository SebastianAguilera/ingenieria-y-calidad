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
	// AcordarSesion cierra una sesión Revelada, actualiza los story points de
	// la historia e inserta la estimación en una única transacción. Bloquea la
	// fila de la sesión y valida el estado dentro de la transacción.
	AcordarSesion(ctx context.Context, sesionID, usuarioID uint, puntos float64) (*StoryEstimate, error)
	// CancelarSesion cancela una sesión Votando/Revelada sin estimar la
	// historia ni borrar votos o rondas. Bloquea la fila de la sesión.
	CancelarSesion(ctx context.Context, sesionID uint) (*SesionPoker, error)
	// ObtenerStoryEstimatePorSesion devuelve la estimación asociada a la sesión,
	// o nil si no fue acordada.
	ObtenerStoryEstimatePorSesion(ctx context.Context, sesionID uint) (*StoryEstimate, error)
}
