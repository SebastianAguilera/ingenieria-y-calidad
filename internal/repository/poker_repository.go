package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ingenieria-y-calidad/internal/domain"
	"ingenieria-y-calidad/internal/repository/models"
)

// sesionConCreador es la fila que devuelven las consultas de sesiones con el
// nombre del creador resuelto por JOIN.
type sesionConCreador struct {
	ID              uint
	ProyectoID      uint
	HistoriaID      uint
	CreadoPor       uint
	CreadorNombre   string
	RondaActual     int
	Estado          string
	PuntosAcordados *float64
	CreadoEn        time.Time
	ActualizadoEn   time.Time
	CerradoEn       *time.Time
	CanceladoEn     *time.Time
}

func (f sesionConCreador) aDominio() domain.SesionPoker {
	return domain.SesionPoker{
		ID:              f.ID,
		ProyectoID:      f.ProyectoID,
		HistoriaID:      f.HistoriaID,
		CreadoPor:       f.CreadoPor,
		CreadorNombre:   f.CreadorNombre,
		RondaActual:     f.RondaActual,
		Estado:          domain.EstadoSesionPoker(f.Estado),
		PuntosAcordados: f.PuntosAcordados,
		CreadoEn:        f.CreadoEn,
		ActualizadoEn:   f.ActualizadoEn,
		CerradoEn:       f.CerradoEn,
		CanceladoEn:     f.CanceladoEn,
	}
}

// votoConNombre es la fila de votos con el nombre del votante resuelto por
// JOIN contra usuarios.
type votoConNombre struct {
	ID        uint
	RondaID   uint
	UsuarioID uint
	Nombre    string
	Puntos    float64
	VotadoEn  time.Time
}

func (f votoConNombre) aDominio() domain.VotoPoker {
	return domain.VotoPoker{
		ID:        f.ID,
		RondaID:   f.RondaID,
		UsuarioID: f.UsuarioID,
		Nombre:    f.Nombre,
		Puntos:    f.Puntos,
		VotadoEn:  f.VotadoEn,
	}
}

type pokerRepository struct {
	db *gorm.DB
}

// NuevoPokerRepository construye el repositorio de Planning Poker sobre GORM.
func NuevoPokerRepository(db *gorm.DB) *pokerRepository {
	return &pokerRepository{db: db}
}

var _ domain.SesionPokerRepository = (*pokerRepository)(nil)

// CrearSesion inserta la sesión en estado Votando y su ronda 1 dentro de una
// transacción. El índice único parcial de historias activas es la red de
// seguridad ante dos creaciones concurrentes (RB-02, RB-13).
func (r *pokerRepository) CrearSesion(ctx context.Context, s *domain.SesionPoker) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ahora := time.Now().UTC()
		s.RondaActual = 1
		s.Estado = domain.EstadoSesionVotando
		s.ActualizadoEn = ahora
		if s.CreadoEn.IsZero() {
			s.CreadoEn = ahora
		}

		m := models.FromDomainSesionPoker(*s)
		if err := tx.Omit("Proyecto", "Historia", "Creador").Create(&m).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return domain.ErrSesionPokerActiva
			}
			return fmt.Errorf("persistiendo sesión de poker: %w", err)
		}
		s.ID = m.ID
		s.CreadoEn = m.CreadoEn
		s.ActualizadoEn = m.ActualizadoEn

		ronda := models.PokerRoundModel{
			SesionID:    s.ID,
			NumeroRonda: 1,
			CreadoEn:    ahora,
		}
		if err := tx.Omit("Sesion").Create(&ronda).Error; err != nil {
			return fmt.Errorf("persistiendo la primera ronda de poker: %w", err)
		}
		return nil
	})
}

// ObtenerPorID busca una sesión dentro de un proyecto concreto. Una sesión de
// otro proyecto se comporta como inexistente.
func (r *pokerRepository) ObtenerPorID(ctx context.Context, proyectoID, sesionID uint) (*domain.SesionPoker, error) {
	var fila sesionConCreador
	err := r.consultaSesiones(ctx).
		Where("poker_sessions.id = ? AND poker_sessions.proyecto_id = ?", sesionID, proyectoID).
		Take(&fila).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSesionPokerNoEncontrada
		}
		return nil, fmt.Errorf("obteniendo sesión de poker %d: %w", sesionID, err)
	}
	sesion := fila.aDominio()
	return &sesion, nil
}

// ObtenerActivaPorHistoria devuelve la sesión Votando/Revelada de la historia,
// o nil si no hay ninguna.
func (r *pokerRepository) ObtenerActivaPorHistoria(ctx context.Context, historiaID uint) (*domain.SesionPoker, error) {
	var fila sesionConCreador
	err := r.consultaSesiones(ctx).
		Where("poker_sessions.historia_id = ? AND poker_sessions.estado IN ?", historiaID, estadosActivos()).
		Order("poker_sessions.id DESC").
		Take(&fila).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("obteniendo la sesión activa de la historia %d: %w", historiaID, err)
	}
	sesion := fila.aDominio()
	return &sesion, nil
}

// ListarPorHistoria devuelve el historial completo de la historia en orden de
// creación descendente (RB-07).
func (r *pokerRepository) ListarPorHistoria(ctx context.Context, historiaID uint) ([]domain.SesionPoker, error) {
	var filas []sesionConCreador
	err := r.consultaSesiones(ctx).
		Where("poker_sessions.historia_id = ?", historiaID).
		Order("poker_sessions.creado_en DESC, poker_sessions.id DESC").
		Scan(&filas).Error
	if err != nil {
		return nil, fmt.Errorf("listando sesiones de la historia %d: %w", historiaID, err)
	}
	return sesionesADominio(filas), nil
}

// ListarActivasPorProyecto devuelve las sesiones Votando/Revelada del proyecto
// en orden de creación ascendente (RB-11).
func (r *pokerRepository) ListarActivasPorProyecto(ctx context.Context, proyectoID uint) ([]domain.SesionPoker, error) {
	var filas []sesionConCreador
	err := r.consultaSesiones(ctx).
		Where("poker_sessions.proyecto_id = ? AND poker_sessions.estado IN ?", proyectoID, estadosActivos()).
		Order("poker_sessions.creado_en ASC, poker_sessions.id ASC").
		Scan(&filas).Error
	if err != nil {
		return nil, fmt.Errorf("listando sesiones activas del proyecto %d: %w", proyectoID, err)
	}
	return sesionesADominio(filas), nil
}

// ObtenerRonda devuelve la ronda con el número indicado de la sesión.
func (r *pokerRepository) ObtenerRonda(ctx context.Context, sesionID uint, numero int) (*domain.RondaPoker, error) {
	var m models.PokerRoundModel
	err := r.db.WithContext(ctx).
		Where("sesion_id = ? AND numero_ronda = ?", sesionID, numero).
		First(&m).Error
	if err != nil {
		return nil, fmt.Errorf("obteniendo la ronda %d de la sesión %d: %w", numero, sesionID, err)
	}
	ronda := models.ToDomainRondaPoker(m)
	return &ronda, nil
}

// ContarMiembros cuenta los vínculos del proyecto en proyecto_usuarios.
func (r *pokerRepository) ContarMiembros(ctx context.Context, proyectoID uint) (int, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Table("proyecto_usuarios").
		Where("proyecto_id = ?", proyectoID).
		Count(&total).Error
	if err != nil {
		return 0, fmt.Errorf("contando miembros del proyecto %d: %w", proyectoID, err)
	}
	return int(total), nil
}

// ContarVotosPorRonda cuenta los votos registrados en la ronda.
func (r *pokerRepository) ContarVotosPorRonda(ctx context.Context, rondaID uint) (int, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&models.PokerVoteModel{}).
		Where("ronda_id = ?", rondaID).
		Count(&total).Error
	if err != nil {
		return 0, fmt.Errorf("contando votos de la ronda %d: %w", rondaID, err)
	}
	return int(total), nil
}

// ListarVotosPorRonda devuelve los votos de la ronda con el nombre del votante.
func (r *pokerRepository) ListarVotosPorRonda(ctx context.Context, rondaID uint) ([]domain.VotoPoker, error) {
	var filas []votoConNombre
	err := r.db.WithContext(ctx).
		Table("poker_votes").
		Select("poker_votes.id, poker_votes.ronda_id, poker_votes.usuario_id, usuarios.nombre AS nombre, poker_votes.puntos, poker_votes.votado_en").
		Joins("JOIN usuarios ON usuarios.id = poker_votes.usuario_id").
		Where("poker_votes.ronda_id = ?", rondaID).
		Order("poker_votes.id ASC").
		Scan(&filas).Error
	if err != nil {
		return nil, fmt.Errorf("listando votos de la ronda %d: %w", rondaID, err)
	}
	votos := make([]domain.VotoPoker, len(filas))
	for idx, fila := range filas {
		votos[idx] = fila.aDominio()
	}
	return votos, nil
}

// ObtenerVotoUsuario devuelve el voto del usuario en la ronda, o nil si todavía
// no votó.
func (r *pokerRepository) ObtenerVotoUsuario(ctx context.Context, rondaID, usuarioID uint) (*domain.VotoPoker, error) {
	var fila votoConNombre
	err := r.db.WithContext(ctx).
		Table("poker_votes").
		Select("poker_votes.id, poker_votes.ronda_id, poker_votes.usuario_id, usuarios.nombre AS nombre, poker_votes.puntos, poker_votes.votado_en").
		Joins("JOIN usuarios ON usuarios.id = poker_votes.usuario_id").
		Where("poker_votes.ronda_id = ? AND poker_votes.usuario_id = ?", rondaID, usuarioID).
		Take(&fila).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("obteniendo el voto del usuario %d en la ronda %d: %w", usuarioID, rondaID, err)
	}
	voto := fila.aDominio()
	return &voto, nil
}

// RegistrarVoto bloquea la fila de la sesión, comprueba que siga Votando y
// hace el upsert del voto en la misma transacción. El bloqueo serializa el voto
// contra el revelado: si el revelado gana, el voto tardío ve el estado nuevo y
// se rechaza sin escribir (US-15, RB-04).
func (r *pokerRepository) RegistrarVoto(ctx context.Context, sesionID, usuarioID uint, puntos float64) (*domain.VotoPoker, error) {
	var voto models.PokerVoteModel
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var sesion models.PokerSessionModel
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id", "estado", "ronda_actual").
			Where("id = ?", sesionID).
			Take(&sesion).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrSesionPokerNoEncontrada
			}
			return fmt.Errorf("bloqueando la sesión de poker %d: %w", sesionID, err)
		}
		if domain.EstadoSesionPoker(sesion.Estado) != domain.EstadoSesionVotando {
			return domain.ErrSesionPokerNoVotando
		}

		var ronda models.PokerRoundModel
		err = tx.Where("sesion_id = ? AND numero_ronda = ?", sesion.ID, sesion.RondaActual).
			Take(&ronda).Error
		if err != nil {
			return fmt.Errorf("obteniendo la ronda %d de la sesión %d: %w", sesion.RondaActual, sesion.ID, err)
		}

		voto = models.PokerVoteModel{
			RondaID:   ronda.ID,
			UsuarioID: usuarioID,
			Puntos:    puntos,
			VotadoEn:  time.Now().UTC(),
		}
		err = tx.Omit("Ronda", "Usuario").Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "ronda_id"}, {Name: "usuario_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"puntos", "votado_en"}),
		}).Create(&voto).Error
		if err != nil {
			return fmt.Errorf("registrando el voto del usuario %d: %w", usuarioID, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &domain.VotoPoker{
		ID:        voto.ID,
		RondaID:   voto.RondaID,
		UsuarioID: voto.UsuarioID,
		Puntos:    voto.Puntos,
		VotadoEn:  voto.VotadoEn,
	}, nil
}

// RevelarSesion bloquea la sesión, valida que esté Votando y que la ronda tenga
// votos, y marca el revelado en la misma transacción.
func (r *pokerRepository) RevelarSesion(ctx context.Context, sesionID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		sesion, err := bloquearSesion(tx, sesionID)
		if err != nil {
			return err
		}
		if domain.EstadoSesionPoker(sesion.Estado) != domain.EstadoSesionVotando {
			return domain.ErrSesionPokerNoVotando
		}

		ronda, err := rondaActual(tx, sesion)
		if err != nil {
			return err
		}

		var total int64
		if err := tx.Model(&models.PokerVoteModel{}).
			Where("ronda_id = ?", ronda.ID).
			Count(&total).Error; err != nil {
			return fmt.Errorf("contando los votos de la ronda %d: %w", ronda.ID, err)
		}
		if total == 0 {
			return domain.ErrSesionPokerSinVotos
		}

		ahora := time.Now().UTC()
		if err := tx.Model(&models.PokerRoundModel{}).
			Where("id = ?", ronda.ID).
			Update("revelada_en", ahora).Error; err != nil {
			return fmt.Errorf("marcando la ronda %d como revelada: %w", ronda.ID, err)
		}
		if err := tx.Model(&models.PokerSessionModel{}).
			Where("id = ?", sesion.ID).
			Updates(map[string]any{
				"estado":         string(domain.EstadoSesionRevelada),
				"actualizado_en": ahora,
			}).Error; err != nil {
			return fmt.Errorf("actualizando la sesión revelada %d: %w", sesion.ID, err)
		}
		return nil
	})
}

// AbrirNuevaRonda bloquea la sesión, valida que esté Revelada y crea la ronda
// siguiente sin tocar las rondas ni los votos anteriores.
func (r *pokerRepository) AbrirNuevaRonda(ctx context.Context, sesionID uint) (*domain.RondaPoker, error) {
	var nueva models.PokerRoundModel
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		sesion, err := bloquearSesion(tx, sesionID)
		if err != nil {
			return err
		}
		if domain.EstadoSesionPoker(sesion.Estado) != domain.EstadoSesionRevelada {
			return domain.ErrSesionPokerNoRevelada
		}

		ahora := time.Now().UTC()
		nueva = models.PokerRoundModel{
			SesionID:    sesion.ID,
			NumeroRonda: sesion.RondaActual + 1,
			CreadoEn:    ahora,
		}
		if err := tx.Omit("Sesion").Create(&nueva).Error; err != nil {
			return fmt.Errorf("creando la ronda %d de la sesión %d: %w", nueva.NumeroRonda, sesion.ID, err)
		}
		if err := tx.Model(&models.PokerSessionModel{}).
			Where("id = ?", sesion.ID).
			Updates(map[string]any{
				"ronda_actual":   nueva.NumeroRonda,
				"estado":         string(domain.EstadoSesionVotando),
				"actualizado_en": ahora,
			}).Error; err != nil {
			return fmt.Errorf("actualizando la sesión %d a la ronda %d: %w", sesion.ID, nueva.NumeroRonda, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	ronda := models.ToDomainRondaPoker(nueva)
	return &ronda, nil
}

// ListarRondasPorSesion devuelve las rondas en orden ascendente.
func (r *pokerRepository) ListarRondasPorSesion(ctx context.Context, sesionID uint) ([]domain.RondaPoker, error) {
	var filas []models.PokerRoundModel
	err := r.db.WithContext(ctx).
		Where("sesion_id = ?", sesionID).
		Order("numero_ronda ASC").
		Find(&filas).Error
	if err != nil {
		return nil, fmt.Errorf("listando las rondas de la sesión %d: %w", sesionID, err)
	}
	rondas := make([]domain.RondaPoker, len(filas))
	for idx, fila := range filas {
		rondas[idx] = models.ToDomainRondaPoker(fila)
	}
	return rondas, nil
}

// AcordarSesion bloquea la sesión, exige el estado Revelada y persiste en una
// única transacción la estimación, el cierre de la sesión y los story points.
func (r *pokerRepository) AcordarSesion(ctx context.Context, sesionID, usuarioID uint, puntos float64) (*domain.StoryEstimate, error) {
	var estimacion models.StoryEstimateModel
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		sesion, err := bloquearSesion(tx, sesionID)
		if err != nil {
			return err
		}
		switch domain.EstadoSesionPoker(sesion.Estado) {
		case domain.EstadoSesionRevelada:
		case domain.EstadoSesionVotando:
			return domain.ErrSesionPokerNoRevelada
		default:
			return domain.ErrSesionPokerCerrada
		}

		ronda, err := rondaActual(tx, sesion)
		if err != nil {
			return err
		}

		ahora := time.Now().UTC()
		resultado := tx.Model(&models.HistoriaModel{}).
			Where("id = ?", sesion.HistoriaID).
			Updates(map[string]any{
				"story_points":   puntos,
				"actualizado_en": ahora,
			})
		if resultado.Error != nil {
			return fmt.Errorf("actualizando los story points de la historia %d: %w", sesion.HistoriaID, resultado.Error)
		}
		if resultado.RowsAffected == 0 {
			return domain.ErrHistoriaNoEncontrada
		}

		if err := tx.Model(&models.PokerSessionModel{}).
			Where("id = ?", sesion.ID).
			Updates(map[string]any{
				"estado":           string(domain.EstadoSesionCerrada),
				"puntos_acordados": puntos,
				"cerrado_en":       ahora,
				"actualizado_en":   ahora,
			}).Error; err != nil {
			return fmt.Errorf("cerrando la sesión de poker %d: %w", sesion.ID, err)
		}

		estimadoPor := usuarioID
		estimacion = models.StoryEstimateModel{
			HistoriaID:  sesion.HistoriaID,
			SesionID:    sesion.ID,
			RondaID:     ronda.ID,
			Puntos:      puntos,
			Fuente:      "planning_poker",
			EstimadoPor: &estimadoPor,
			EstimadoEn:  ahora,
		}
		if err := tx.Omit("Historia", "Sesion", "Ronda").Create(&estimacion).Error; err != nil {
			return fmt.Errorf("registrando la estimación de la sesión %d: %w", sesion.ID, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	resultado := models.ToDomainStoryEstimate(estimacion)
	return &resultado, nil
}

// CancelarSesion bloquea la sesión, exige Votando/Revelada y la marca
// Cancelada sin tocar la historia, la estimación, los votos ni las rondas.
func (r *pokerRepository) CancelarSesion(ctx context.Context, sesionID uint) (*domain.SesionPoker, error) {
	var actualizada domain.SesionPoker
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		sesion, err := bloquearSesion(tx, sesionID)
		if err != nil {
			return err
		}
		switch domain.EstadoSesionPoker(sesion.Estado) {
		case domain.EstadoSesionVotando, domain.EstadoSesionRevelada:
		default:
			return domain.ErrSesionPokerCerrada
		}

		ahora := time.Now().UTC()
		if err := tx.Model(&models.PokerSessionModel{}).
			Where("id = ?", sesion.ID).
			Updates(map[string]any{
				"estado":         string(domain.EstadoSesionCancelada),
				"cancelado_en":   ahora,
				"actualizado_en": ahora,
			}).Error; err != nil {
			return fmt.Errorf("cancelando la sesión de poker %d: %w", sesion.ID, err)
		}

		var fila sesionConCreador
		err = sesionesConCreadorQuery(tx).
			Where("poker_sessions.id = ?", sesion.ID).
			Take(&fila).Error
		if err != nil {
			return fmt.Errorf("obteniendo la sesión cancelada %d: %w", sesion.ID, err)
		}
		actualizada = fila.aDominio()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &actualizada, nil
}

// ObtenerStoryEstimatePorSesion devuelve la estimación acordada de la sesión, o
// nil si la sesión no fue cerrada con acuerdo.
func (r *pokerRepository) ObtenerStoryEstimatePorSesion(ctx context.Context, sesionID uint) (*domain.StoryEstimate, error) {
	var m models.StoryEstimateModel
	err := r.db.WithContext(ctx).
		Where("sesion_id = ?", sesionID).
		Take(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("obteniendo la estimación de la sesión %d: %w", sesionID, err)
	}
	estimacion := models.ToDomainStoryEstimate(m)
	return &estimacion, nil
}

func bloquearSesion(tx *gorm.DB, sesionID uint) (*models.PokerSessionModel, error) {
	var sesion models.PokerSessionModel
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id", "proyecto_id", "historia_id", "creado_por", "estado", "ronda_actual", "puntos_acordados", "cerrado_en", "cancelado_en").
		Where("id = ?", sesionID).
		Take(&sesion).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSesionPokerNoEncontrada
		}
		return nil, fmt.Errorf("bloqueando la sesión de poker %d: %w", sesionID, err)
	}
	return &sesion, nil
}

func rondaActual(tx *gorm.DB, sesion *models.PokerSessionModel) (*models.PokerRoundModel, error) {
	var ronda models.PokerRoundModel
	err := tx.Where("sesion_id = ? AND numero_ronda = ?", sesion.ID, sesion.RondaActual).
		Take(&ronda).Error
	if err != nil {
		return nil, fmt.Errorf("obteniendo la ronda %d de la sesión %d: %w", sesion.RondaActual, sesion.ID, err)
	}
	return &ronda, nil
}

// consultaSesiones arma la consulta base de sesiones con el nombre del creador.
func (r *pokerRepository) consultaSesiones(ctx context.Context) *gorm.DB {
	return sesionesConCreadorQuery(r.db.WithContext(ctx))
}

func sesionesConCreadorQuery(db *gorm.DB) *gorm.DB {
	return db.
		Table("poker_sessions").
		Select("poker_sessions.id, poker_sessions.proyecto_id, poker_sessions.historia_id, poker_sessions.creado_por, usuarios.nombre AS creador_nombre, poker_sessions.ronda_actual, poker_sessions.estado, poker_sessions.puntos_acordados, poker_sessions.creado_en, poker_sessions.actualizado_en, poker_sessions.cerrado_en, poker_sessions.cancelado_en").
		Joins("JOIN usuarios ON usuarios.id = poker_sessions.creado_por")
}

func estadosActivos() []string {
	return []string{
		string(domain.EstadoSesionVotando),
		string(domain.EstadoSesionRevelada),
	}
}

func sesionesADominio(filas []sesionConCreador) []domain.SesionPoker {
	sesiones := make([]domain.SesionPoker, len(filas))
	for idx, fila := range filas {
		sesiones[idx] = fila.aDominio()
	}
	return sesiones
}
