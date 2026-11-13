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

// consultaSesiones arma la consulta base de sesiones con el nombre del creador.
func (r *pokerRepository) consultaSesiones(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Table("poker_sessions").
		Select("poker_sessions.id, poker_sessions.proyecto_id, poker_sessions.historia_id, poker_sessions.creado_por, usuarios.nombre AS creador_nombre, poker_sessions.ronda_actual, poker_sessions.estado, poker_sessions.puntos_acordados, poker_sessions.creado_en, poker_sessions.actualizado_en, poker_sessions.cerrado_en").
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
