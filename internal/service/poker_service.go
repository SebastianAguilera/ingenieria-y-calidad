package service

import (
	"context"
	"fmt"

	"ingenieria-y-calidad/internal/domain"
)

// PokerService implementa las reglas de negocio de las sesiones de Planning
// Poker de US-14. Depende únicamente de las interfaces definidas en
// internal/domain.
type PokerService struct {
	proyectos domain.ProyectoRepository
	historias domain.HistoriaRepository
	usuarios  domain.UsuarioRepository
	poker     domain.SesionPokerRepository
}

// NewPokerService construye el servicio de Planning Poker con inyección de
// dependencias.
func NewPokerService(
	proyectos domain.ProyectoRepository,
	historias domain.HistoriaRepository,
	usuarios domain.UsuarioRepository,
	poker domain.SesionPokerRepository,
) *PokerService {
	return &PokerService{
		proyectos: proyectos,
		historias: historias,
		usuarios:  usuarios,
		poker:     poker,
	}
}

// CrearSesion valida que la historia exista en el proyecto y sea estimable,
// que no haya otra sesión activa y crea la sesión en Votando con la ronda 1.
func (s *PokerService) CrearSesion(ctx context.Context, proyectoID, historiaID, usuarioID uint) (*domain.DetalleSesionPoker, error) {
	if err := s.exigirProyecto(ctx, proyectoID); err != nil {
		return nil, err
	}

	historia, err := s.historias.ObtenerPorID(ctx, proyectoID, historiaID)
	if err != nil {
		return nil, fmt.Errorf("obteniendo historia para Planning Poker: %w", err)
	}
	if historia.Estado == domain.EstadoHistoriaHecho {
		return nil, domain.ErrHistoriaNoEstimable
	}

	activa, err := s.poker.ObtenerActivaPorHistoria(ctx, historiaID)
	if err != nil {
		return nil, fmt.Errorf("verificando la sesión activa de la historia: %w", err)
	}
	if activa != nil {
		return nil, domain.ErrSesionPokerActiva
	}

	creador, err := s.usuarios.ObtenerPorID(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("obteniendo el creador de la sesión: %w", err)
	}

	sesion := &domain.SesionPoker{
		ProyectoID: proyectoID,
		HistoriaID: historiaID,
		CreadoPor:  usuarioID,
	}
	if err := s.poker.CrearSesion(ctx, sesion); err != nil {
		return nil, fmt.Errorf("creando la sesión de Planning Poker: %w", err)
	}

	totalMiembros, err := s.poker.ContarMiembros(ctx, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("contando los miembros del proyecto: %w", err)
	}

	return &domain.DetalleSesionPoker{
		Sesion:           *sesion,
		Creador:          domain.ReferenciaUsuario{ID: creador.ID, Nombre: creador.Nombre},
		TotalMiembros:    totalMiembros,
		VotosRegistrados: 0,
		YaVote:           false,
		Votos:            []domain.VotoPoker{},
	}, nil
}

// ListarPorHistoria devuelve el historial de sesiones de la historia, incluidas
// las cerradas y canceladas.
func (s *PokerService) ListarPorHistoria(ctx context.Context, proyectoID, historiaID uint) ([]domain.SesionPoker, error) {
	if err := s.exigirProyecto(ctx, proyectoID); err != nil {
		return nil, err
	}
	if _, err := s.historias.ObtenerPorID(ctx, proyectoID, historiaID); err != nil {
		return nil, fmt.Errorf("obteniendo historia para Planning Poker: %w", err)
	}

	sesiones, err := s.poker.ListarPorHistoria(ctx, historiaID)
	if err != nil {
		return nil, fmt.Errorf("listando las sesiones de la historia: %w", err)
	}
	if sesiones == nil {
		return []domain.SesionPoker{}, nil
	}
	return sesiones, nil
}

// Detalle devuelve la sesión con sus conteos y, solo en estado Revelada, los
// votos con nombre y puntos. En Votando nunca expone votos ajenos.
func (s *PokerService) Detalle(ctx context.Context, proyectoID, sesionID, usuarioID uint) (*domain.DetalleSesionPoker, error) {
	if err := s.exigirProyecto(ctx, proyectoID); err != nil {
		return nil, err
	}

	sesion, err := s.poker.ObtenerPorID(ctx, proyectoID, sesionID)
	if err != nil {
		return nil, fmt.Errorf("obteniendo la sesión de Planning Poker: %w", err)
	}

	ronda, err := s.poker.ObtenerRonda(ctx, sesion.ID, sesion.RondaActual)
	if err != nil {
		return nil, fmt.Errorf("obteniendo la ronda actual: %w", err)
	}

	totalMiembros, err := s.poker.ContarMiembros(ctx, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("contando los miembros del proyecto: %w", err)
	}

	votosRegistrados, err := s.poker.ContarVotosPorRonda(ctx, ronda.ID)
	if err != nil {
		return nil, fmt.Errorf("contando los votos de la ronda: %w", err)
	}

	voto, err := s.poker.ObtenerVotoUsuario(ctx, ronda.ID, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("obteniendo el voto del usuario: %w", err)
	}

	detalle := &domain.DetalleSesionPoker{
		Sesion:           *sesion,
		Creador:          domain.ReferenciaUsuario{ID: sesion.CreadoPor, Nombre: sesion.CreadorNombre},
		TotalMiembros:    totalMiembros,
		VotosRegistrados: votosRegistrados,
		YaVote:           voto != nil,
		Votos:            []domain.VotoPoker{},
	}
	if voto != nil {
		puntos := voto.Puntos
		detalle.MiVoto = &puntos
	}

	if sesion.Estado == domain.EstadoSesionRevelada {
		votos, err := s.poker.ListarVotosPorRonda(ctx, ronda.ID)
		if err != nil {
			return nil, fmt.Errorf("listando los votos de la ronda: %w", err)
		}
		if votos != nil {
			detalle.Votos = votos
		}
	}

	return detalle, nil
}

// Activas devuelve las sesiones Votando/Revelada del proyecto con sus conteos y
// el resumen de la historia. Nunca incluye votos, ni siquiera en Revelada.
func (s *PokerService) Activas(ctx context.Context, proyectoID, usuarioID uint) ([]domain.SesionPokerActiva, error) {
	if err := s.exigirProyecto(ctx, proyectoID); err != nil {
		return nil, err
	}

	sesiones, err := s.poker.ListarActivasPorProyecto(ctx, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("listando las sesiones activas: %w", err)
	}

	totalMiembros, err := s.poker.ContarMiembros(ctx, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("contando los miembros del proyecto: %w", err)
	}

	activas := make([]domain.SesionPokerActiva, 0, len(sesiones))
	for _, sesion := range sesiones {
		historia, err := s.historias.ObtenerPorID(ctx, proyectoID, sesion.HistoriaID)
		if err != nil {
			return nil, fmt.Errorf("obteniendo la historia de la sesión activa: %w", err)
		}

		ronda, err := s.poker.ObtenerRonda(ctx, sesion.ID, sesion.RondaActual)
		if err != nil {
			return nil, fmt.Errorf("obteniendo la ronda actual: %w", err)
		}

		votosRegistrados, err := s.poker.ContarVotosPorRonda(ctx, ronda.ID)
		if err != nil {
			return nil, fmt.Errorf("contando los votos de la ronda: %w", err)
		}

		voto, err := s.poker.ObtenerVotoUsuario(ctx, ronda.ID, usuarioID)
		if err != nil {
			return nil, fmt.Errorf("obteniendo el voto del usuario: %w", err)
		}

		activas = append(activas, domain.SesionPokerActiva{
			Sesion:           sesion,
			Creador:          domain.ReferenciaUsuario{ID: sesion.CreadoPor, Nombre: sesion.CreadorNombre},
			Historia:         domain.ResumenHistoria{ID: historia.ID, Codigo: historia.Codigo, Titulo: historia.Titulo},
			TotalMiembros:    totalMiembros,
			VotosRegistrados: votosRegistrados,
			YaVote:           voto != nil,
		})
	}

	return activas, nil
}

// Votar registra o modifica el voto del usuario en la ronda actual. Valida la
// escala de US-05 y delega en el repositorio la comprobación del estado Votando
// y la escritura, que ocurren bajo bloqueo de fila. Devuelve solo el voto propio
// y conteos: nunca votos ajenos.
func (s *PokerService) Votar(ctx context.Context, proyectoID, sesionID, usuarioID uint, puntos float64) (*domain.ResultadoVotoPoker, error) {
	if err := s.exigirProyecto(ctx, proyectoID); err != nil {
		return nil, err
	}

	sesion, err := s.poker.ObtenerPorID(ctx, proyectoID, sesionID)
	if err != nil {
		return nil, fmt.Errorf("obteniendo la sesión de Planning Poker: %w", err)
	}

	if !domain.StoryPointsValidos(puntos) {
		return nil, domain.ErrPuntosPokerInvalidos
	}

	voto, err := s.poker.RegistrarVoto(ctx, sesion.ID, usuarioID, puntos)
	if err != nil {
		return nil, fmt.Errorf("registrando el voto: %w", err)
	}

	votosRegistrados, err := s.poker.ContarVotosPorRonda(ctx, voto.RondaID)
	if err != nil {
		return nil, fmt.Errorf("contando los votos de la ronda: %w", err)
	}

	totalMiembros, err := s.poker.ContarMiembros(ctx, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("contando los miembros del proyecto: %w", err)
	}

	return &domain.ResultadoVotoPoker{
		SesionID:         sesion.ID,
		Puntos:           voto.Puntos,
		VotadoEn:         voto.VotadoEn,
		VotosRegistrados: votosRegistrados,
		TotalMiembros:    totalMiembros,
	}, nil
}

// exigirProyecto confirma que el proyecto existe antes de operar sobre sus
// sesiones de Planning Poker.
func (s *PokerService) exigirProyecto(ctx context.Context, proyectoID uint) error {
	if _, err := s.proyectos.ObtenerPorID(ctx, proyectoID); err != nil {
		return fmt.Errorf("obteniendo proyecto: %w", err)
	}
	return nil
}
