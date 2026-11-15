package domain

import "time"

// EstadoSesionPoker representa el estado de una sesión de Planning Poker.
// Solo admite los valores declarados en las constantes siguientes.
type EstadoSesionPoker string

const (
	EstadoSesionVotando   EstadoSesionPoker = "Votando"
	EstadoSesionRevelada  EstadoSesionPoker = "Revelada"
	EstadoSesionCerrada   EstadoSesionPoker = "Cerrada"
	EstadoSesionCancelada EstadoSesionPoker = "Cancelada"
)

// Valido indica si el estado corresponde a un valor aceptado por el dominio.
func (e EstadoSesionPoker) Valido() bool {
	return e == EstadoSesionVotando || e == EstadoSesionRevelada ||
		e == EstadoSesionCerrada || e == EstadoSesionCancelada
}

// Activa indica si la sesión bloquea la creación de otra sesión para la misma
// historia. Solo Votando y Revelada se consideran activas (US-14, RB-02).
func (e EstadoSesionPoker) Activa() bool {
	return e == EstadoSesionVotando || e == EstadoSesionRevelada
}

// SesionPoker es una sesión de estimación en equipo para una historia.
// CreadorNombre no se persiste: lo completa el repositorio con un JOIN.
type SesionPoker struct {
	ID              uint
	ProyectoID      uint
	HistoriaID      uint
	CreadoPor       uint
	CreadorNombre   string
	RondaActual     int
	Estado          EstadoSesionPoker
	PuntosAcordados *float64
	CreadoEn        time.Time
	ActualizadoEn   time.Time
	CerradoEn       *time.Time
	CanceladoEn     *time.Time
	StoryEstimate   *StoryEstimate
}

// RondaPoker es una ronda de votación dentro de una sesión de Planning Poker.
type RondaPoker struct {
	ID          uint
	SesionID    uint
	NumeroRonda int
	ReveladaEn  *time.Time
	CreadoEn    time.Time
}

// EstadoRonda deriva el estado visible de la ronda a partir del revelado.
func (r RondaPoker) EstadoRonda() EstadoSesionPoker {
	if r.ReveladaEn != nil {
		return EstadoSesionRevelada
	}
	return EstadoSesionVotando
}

// VotoPoker es el voto de un miembro en una ronda. Nombre no se persiste: lo
// completa el repositorio con un JOIN contra usuarios.
type VotoPoker struct {
	ID        uint
	RondaID   uint
	UsuarioID uint
	Nombre    string
	Puntos    float64
	VotadoEn  time.Time
	Ronda     int
}

// DispersionPoker agrupa las estadísticas de los votos de una ronda revelada.
type DispersionPoker struct {
	Promedio           float64
	Mediana            float64
	Minimo             float64
	Maximo             float64
	Rango              float64
	DesviacionEstandar float64
	HayDispersion      bool
	TotalVotos         int
	VotosPorValor      map[string]int
}

// HistorialRondaPoker es una ronda del historial de la sesión. Los votos y la
// dispersión solo se completan cuando la ronda fue revelada.
type HistorialRondaPoker struct {
	NumeroRonda int
	Estado      EstadoSesionPoker
	ReveladaEn  *time.Time
	Votos       []VotoPoker
	Dispersion  *DispersionPoker
}

// StoryEstimate es la estimación acordada para una historia. Es la base de
// US-16; US-14 solo crea su esquema.
type StoryEstimate struct {
	ID          uint
	HistoriaID  uint
	SesionID    uint
	RondaID     uint
	Puntos      float64
	Fuente      string
	EstimadoPor *uint
	EstimadoEn  time.Time
}

// ReferenciaUsuario es la identificación mínima de un usuario en las respuestas
// de Planning Poker.
type ReferenciaUsuario struct {
	ID     uint
	Nombre string
}

// ResumenHistoria es la identificación mínima de la historia en la lista de
// sesiones activas.
type ResumenHistoria struct {
	ID     uint
	Codigo string
	Titulo string
}

// DetalleSesionPoker agrega la sesión con los datos que expone el detalle.
// En estado Votando, Votos queda vacío y MiVoto solo informa los puntos
// propios (RB-08).
type DetalleSesionPoker struct {
	Sesion           SesionPoker
	Creador          ReferenciaUsuario
	TotalMiembros    int
	VotosRegistrados int
	YaVote           bool
	MiVoto           *float64
	Votos            []VotoPoker
	Dispersion       *DispersionPoker
	HistorialRondas  []HistorialRondaPoker
}

// SesionPokerActiva es un ítem de la lista de sesiones activas del proyecto.
// Nunca incluye votos, incluso en estado Revelada (RB-11).
type SesionPokerActiva struct {
	Sesion           SesionPoker
	Creador          ReferenciaUsuario
	Historia         ResumenHistoria
	TotalMiembros    int
	VotosRegistrados int
	YaVote           bool
}

// ResultadoVotoPoker es lo que se informa al votante tras registrar su voto
// (US-15). Solo contiene su propio voto y conteos: nunca votos ajenos.
type ResultadoVotoPoker struct {
	SesionID         uint
	Puntos           float64
	VotadoEn         time.Time
	VotosRegistrados int
	TotalMiembros    int
}
