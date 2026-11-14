package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ingenieria-y-calidad/internal/domain"
)

// PokerService implementa las reglas de negocio de las sesiones de Planning
// Poker (US-14 a US-17). Depende únicamente de las interfaces definidas en
// internal/domain.
type PokerService struct {
	proyectos    domain.ProyectoRepository
	historias    domain.HistoriaRepository
	usuarios     domain.UsuarioRepository
	poker        domain.SesionPokerRepository
	estimaciones domain.StoryEstimateRepository
	uow          domain.UnidadDeTrabajo
}

// NewPokerService construye el servicio de Planning Poker con inyección de
// dependencias. uow ejecuta el acuerdo de US-17 de forma atómica.
func NewPokerService(
	proyectos domain.ProyectoRepository,
	historias domain.HistoriaRepository,
	usuarios domain.UsuarioRepository,
	poker domain.SesionPokerRepository,
	estimaciones domain.StoryEstimateRepository,
	uow domain.UnidadDeTrabajo,
) *PokerService {
	return &PokerService{
		proyectos:    proyectos,
		historias:    historias,
		usuarios:     usuarios,
		poker:        poker,
		estimaciones: estimaciones,
		uow:          uow,
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
	if err := s.adjuntarEstimaciones(ctx, sesiones); err != nil {
		return nil, err
	}
	return sesiones, nil
}

// adjuntarEstimaciones completa el story_estimate de las sesiones Cerradas
// (US-17 RB-13/RB-14) con una sola consulta. Las demás no tienen estimación.
func (s *PokerService) adjuntarEstimaciones(ctx context.Context, sesiones []domain.SesionPoker) error {
	var cerradas []uint
	for _, sesion := range sesiones {
		if sesion.Estado == domain.EstadoSesionCerrada {
			cerradas = append(cerradas, sesion.ID)
		}
	}
	if len(cerradas) == 0 {
		return nil
	}

	estimaciones, err := s.estimaciones.ListarPorSesiones(ctx, cerradas)
	if err != nil {
		return fmt.Errorf("listando los story_estimate de las sesiones: %w", err)
	}
	porSesion := make(map[uint]domain.StoryEstimate, len(estimaciones))
	for _, estimacion := range estimaciones {
		if estimacion.SesionID != nil {
			porSesion[*estimacion.SesionID] = estimacion
		}
	}
	for idx := range sesiones {
		if estimacion, ok := porSesion[sesiones[idx].ID]; ok {
			sesiones[idx].Estimacion = &estimacion
		}
	}
	return nil
}

// Acordar registra la estimación acordada de una sesión Revelada (US-17). En
// una única transacción cierra la sesión, fija los story points de la historia
// por el camino de US-05 y crea el story_estimate; si cualquiera de los pasos
// falla no queda ningún cambio. Devuelve el detalle de la sesión Cerrada.
func (s *PokerService) Acordar(ctx context.Context, proyectoID, sesionID, usuarioID uint, puntos float64) (*domain.DetalleSesionPoker, error) {
	if err := s.exigirProyecto(ctx, proyectoID); err != nil {
		return nil, err
	}

	sesion, err := s.poker.ObtenerPorID(ctx, proyectoID, sesionID)
	if err != nil {
		return nil, fmt.Errorf("obteniendo la sesión de Planning Poker: %w", err)
	}

	// Validación de US-05 sobre la misma actualización que se aplicará a la
	// historia. RB-16: se evalúa después del 404 y antes del estado (409).
	actualizacion := domain.ActualizacionHistoria{StoryPoints: &puntos}
	if err := validarActualizacionHistoria(actualizacion); err != nil {
		if errors.Is(err, domain.ErrStoryPointsInvalido) {
			return nil, domain.ErrPuntosPokerInvalidos
		}
		return nil, err
	}

	err = s.uow.Ejecutar(ctx, func(repos domain.RepositoriosTransaccion) error {
		// Primero se bloquea y cierra la sesión: un acuerdo concurrente espera
		// el bloqueo y luego ve la sesión Cerrada (RB-07).
		ronda, err := repos.Poker().CerrarSesion(ctx, sesion.ID, puntos)
		if err != nil {
			return fmt.Errorf("cerrando la sesión de Planning Poker: %w", err)
		}

		historia, err := repos.Historias().ObtenerPorID(ctx, proyectoID, sesion.HistoriaID)
		if err != nil {
			return fmt.Errorf("obteniendo la historia acordada: %w", err)
		}
		aplicarActualizacionHistoria(historia, actualizacion)
		if err := repos.Historias().Actualizar(ctx, historia); err != nil {
			return fmt.Errorf("actualizando los story points de la historia: %w", err)
		}

		estimacion := domain.NuevaEstimacionPlanningPoker(*sesion, ronda.ID, usuarioID, puntos, time.Now().UTC())
		if err := repos.Estimaciones().Crear(ctx, &estimacion); err != nil {
			return fmt.Errorf("creando el story_estimate: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.detalleActualizado(ctx, proyectoID, sesion.ID, usuarioID)
}

// Cancelar pasa una sesión Votando o Revelada a Cancelada (US-17). No fija
// puntos, no crea story_estimate ni modifica la historia, y conserva votos y
// rondas. Al dejar de estar activa, la historia queda libre para otra sesión.
func (s *PokerService) Cancelar(ctx context.Context, proyectoID, sesionID, usuarioID uint) (*domain.DetalleSesionPoker, error) {
	if err := s.exigirProyecto(ctx, proyectoID); err != nil {
		return nil, err
	}

	sesion, err := s.poker.ObtenerPorID(ctx, proyectoID, sesionID)
	if err != nil {
		return nil, fmt.Errorf("obteniendo la sesión de Planning Poker: %w", err)
	}

	if err := s.poker.CancelarSesion(ctx, sesion.ID); err != nil {
		return nil, fmt.Errorf("cancelando la sesión de Planning Poker: %w", err)
	}

	return s.detalleActualizado(ctx, proyectoID, sesion.ID, usuarioID)
}

// detalleActualizado vuelve a leer la sesión tras una transición para
// informar las marcas de tiempo y los puntos tal como quedaron persistidos.
func (s *PokerService) detalleActualizado(ctx context.Context, proyectoID, sesionID, usuarioID uint) (*domain.DetalleSesionPoker, error) {
	sesion, err := s.poker.ObtenerPorID(ctx, proyectoID, sesionID)
	if err != nil {
		return nil, fmt.Errorf("obteniendo la sesión de Planning Poker: %w", err)
	}
	return s.detalleDeSesion(ctx, proyectoID, sesion, usuarioID)
}

// Detalle devuelve la sesión con sus conteos y, solo después del revelado, los
// votos con nombre y puntos, la dispersión y el historial de rondas.
func (s *PokerService) Detalle(ctx context.Context, proyectoID, sesionID, usuarioID uint) (*domain.DetalleSesionPoker, error) {
	if err := s.exigirProyecto(ctx, proyectoID); err != nil {
		return nil, err
	}

	sesion, err := s.poker.ObtenerPorID(ctx, proyectoID, sesionID)
	if err != nil {
		return nil, fmt.Errorf("obteniendo la sesión de Planning Poker: %w", err)
	}

	return s.detalleDeSesion(ctx, proyectoID, sesion, usuarioID)
}

// Revelar pasa una sesión Votando con votos a Revelada y devuelve el detalle
// con los votos y la dispersión de la ronda.
func (s *PokerService) Revelar(ctx context.Context, proyectoID, sesionID, usuarioID uint) (*domain.DetalleSesionPoker, error) {
	if err := s.exigirProyecto(ctx, proyectoID); err != nil {
		return nil, err
	}

	sesion, err := s.poker.ObtenerPorID(ctx, proyectoID, sesionID)
	if err != nil {
		return nil, fmt.Errorf("obteniendo la sesión de Planning Poker: %w", err)
	}

	if err := s.poker.RevelarSesion(ctx, sesion.ID); err != nil {
		return nil, fmt.Errorf("revelando la sesión de Planning Poker: %w", err)
	}
	sesion.Estado = domain.EstadoSesionRevelada

	return s.detalleDeSesion(ctx, proyectoID, sesion, usuarioID)
}

// NuevaRonda abre la ronda siguiente de una sesión Revelada y devuelve el
// detalle en Votando conservando el historial anterior.
func (s *PokerService) NuevaRonda(ctx context.Context, proyectoID, sesionID, usuarioID uint) (*domain.DetalleSesionPoker, error) {
	if err := s.exigirProyecto(ctx, proyectoID); err != nil {
		return nil, err
	}

	sesion, err := s.poker.ObtenerPorID(ctx, proyectoID, sesionID)
	if err != nil {
		return nil, fmt.Errorf("obteniendo la sesión de Planning Poker: %w", err)
	}

	ronda, err := s.poker.AbrirNuevaRonda(ctx, sesion.ID)
	if err != nil {
		return nil, fmt.Errorf("abriendo una nueva ronda de Planning Poker: %w", err)
	}
	sesion.RondaActual = ronda.NumeroRonda
	sesion.Estado = domain.EstadoSesionVotando

	return s.detalleDeSesion(ctx, proyectoID, sesion, usuarioID)
}

func (s *PokerService) detalleDeSesion(ctx context.Context, proyectoID uint, sesion *domain.SesionPoker, usuarioID uint) (*domain.DetalleSesionPoker, error) {
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
		HistorialRondas:  []domain.HistorialRondaPoker{},
	}
	if voto != nil {
		puntos := voto.Puntos
		detalle.MiVoto = &puntos
	}

	if sesion.Estado == domain.EstadoSesionCerrada && sesion.Estimacion == nil {
		sesiones := []domain.SesionPoker{*sesion}
		if err := s.adjuntarEstimaciones(ctx, sesiones); err != nil {
			return nil, err
		}
		detalle.Sesion.Estimacion = sesiones[0].Estimacion
	}

	votosPorRonda := map[uint][]domain.VotoPoker{}
	if votosDeRondaActualVisibles(sesion.Estado, ronda) {
		votos, err := s.poker.ListarVotosPorRonda(ctx, ronda.ID)
		if err != nil {
			return nil, fmt.Errorf("listando los votos de la ronda: %w", err)
		}
		votos = votosConRonda(votos, ronda.NumeroRonda)
		votosPorRonda[ronda.ID] = votos
		dispersion := domain.CalcularDispersionPoker(votos)
		detalle.Votos = votos
		detalle.Dispersion = &dispersion
	}

	rondas, err := s.poker.ListarRondasPorSesion(ctx, sesion.ID)
	if err != nil {
		return nil, fmt.Errorf("listando el historial de rondas: %w", err)
	}
	for _, rondaHistorial := range rondas {
		item := domain.HistorialRondaPoker{
			NumeroRonda: rondaHistorial.NumeroRonda,
			Estado:      rondaHistorial.EstadoRonda(),
			ReveladaEn:  rondaHistorial.ReveladaEn,
			Votos:       []domain.VotoPoker{},
		}
		if rondaHistorial.ReveladaEn != nil {
			votos, cargados := votosPorRonda[rondaHistorial.ID]
			if !cargados {
				votos, err = s.poker.ListarVotosPorRonda(ctx, rondaHistorial.ID)
				if err != nil {
					return nil, fmt.Errorf("listando los votos de la ronda %d: %w", rondaHistorial.NumeroRonda, err)
				}
				votos = votosConRonda(votos, rondaHistorial.NumeroRonda)
			}
			dispersion := domain.CalcularDispersionPoker(votos)
			item.Votos = votos
			item.Dispersion = &dispersion
		}
		detalle.HistorialRondas = append(detalle.HistorialRondas, item)
	}

	return detalle, nil
}

// votosDeRondaActualVisibles decide si el detalle expone los votos de la ronda
// actual. En Revelada y Cerrada siempre se ven; una sesión Cancelada solo los
// muestra si la ronda llegó a revelarse, para no filtrar votos ocultos de una
// sesión cancelada mientras se votaba (US-16 RB-08, US-17 RB-13).
func votosDeRondaActualVisibles(estado domain.EstadoSesionPoker, ronda *domain.RondaPoker) bool {
	switch estado {
	case domain.EstadoSesionRevelada, domain.EstadoSesionCerrada:
		return true
	case domain.EstadoSesionCancelada:
		return ronda.ReveladaEn != nil
	default:
		return false
	}
}

func votosConRonda(votos []domain.VotoPoker, numeroRonda int) []domain.VotoPoker {
	resultado := make([]domain.VotoPoker, len(votos))
	for idx, voto := range votos {
		voto.Ronda = numeroRonda
		resultado[idx] = voto
	}
	return resultado
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
