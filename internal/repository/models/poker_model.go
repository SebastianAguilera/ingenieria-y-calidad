package models

import "time"

// PokerSessionModel representa la tabla "poker_sessions". El índice único
// parcial (historia_id) WHERE estado IN ('Votando','Revelada') garantiza una
// sola sesión activa por historia; se crea en migraciones.go porque GORM no lo
// expresa de forma portable.
type PokerSessionModel struct {
	ID              uint          `gorm:"column:id;primaryKey;autoIncrement"`
	ProyectoID      uint          `gorm:"column:proyecto_id;not null"`
	HistoriaID      uint          `gorm:"column:historia_id;not null"`
	CreadoPor       uint          `gorm:"column:creado_por;not null"`
	RondaActual     int           `gorm:"column:ronda_actual;not null;default:1"`
	Estado          string        `gorm:"column:estado;type:varchar(20);not null;default:'Votando'"`
	PuntosAcordados *float64      `gorm:"column:puntos_acordados"`
	CreadoEn        time.Time     `gorm:"column:creado_en;type:timestamptz;not null;default:now()"`
	ActualizadoEn   time.Time     `gorm:"column:actualizado_en;type:timestamptz;not null;default:now()"`
	CerradoEn       *time.Time    `gorm:"column:cerrado_en;type:timestamptz"`
	Proyecto        ProyectoModel `gorm:"foreignKey:ProyectoID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Historia        HistoriaModel `gorm:"foreignKey:HistoriaID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Creador         UsuarioModel  `gorm:"foreignKey:CreadoPor;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

// TableName define el nombre de la tabla en PostgreSQL.
func (PokerSessionModel) TableName() string {
	return "poker_sessions"
}

// PokerRoundModel representa una ronda de votación de una sesión.
type PokerRoundModel struct {
	ID          uint              `gorm:"column:id;primaryKey;autoIncrement"`
	SesionID    uint              `gorm:"column:sesion_id;not null;uniqueIndex:ux_poker_rounds_sesion_numero,priority:1"`
	NumeroRonda int               `gorm:"column:numero_ronda;not null;uniqueIndex:ux_poker_rounds_sesion_numero,priority:2"`
	ReveladaEn  *time.Time        `gorm:"column:revelada_en;type:timestamptz"`
	CreadoEn    time.Time         `gorm:"column:creado_en;type:timestamptz;not null;default:now()"`
	Sesion      PokerSessionModel `gorm:"foreignKey:SesionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// TableName define el nombre de la tabla en PostgreSQL.
func (PokerRoundModel) TableName() string {
	return "poker_rounds"
}

// PokerVoteModel representa el voto de un usuario en una ronda.
type PokerVoteModel struct {
	ID        uint            `gorm:"column:id;primaryKey;autoIncrement"`
	RondaID   uint            `gorm:"column:ronda_id;not null;uniqueIndex:ux_poker_votes_ronda_usuario,priority:1"`
	UsuarioID uint            `gorm:"column:usuario_id;not null;uniqueIndex:ux_poker_votes_ronda_usuario,priority:2"`
	Puntos    float64         `gorm:"column:puntos;not null"`
	VotadoEn  time.Time       `gorm:"column:votado_en;type:timestamptz;not null;default:now()"`
	Ronda     PokerRoundModel `gorm:"foreignKey:RondaID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Usuario   UsuarioModel    `gorm:"foreignKey:UsuarioID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

// TableName define el nombre de la tabla en PostgreSQL.
func (PokerVoteModel) TableName() string {
	return "poker_votes"
}

// StoryEstimateModel representa la estimación acordada de una historia. US-14
// solo crea el esquema; la escritura llega con US-16.
type StoryEstimateModel struct {
	ID          uint          `gorm:"column:id;primaryKey;autoIncrement"`
	HistoriaID  uint          `gorm:"column:historia_id;not null"`
	Puntos      float64       `gorm:"column:puntos;not null"`
	Fuente      string        `gorm:"column:fuente;type:varchar(20);not null;default:'planning_poker'"`
	EstimadoPor *uint         `gorm:"column:estimado_por"`
	EstimadoEn  time.Time     `gorm:"column:estimado_en;type:timestamptz;not null;default:now()"`
	Historia    HistoriaModel `gorm:"foreignKey:HistoriaID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// TableName define el nombre de la tabla en PostgreSQL.
func (StoryEstimateModel) TableName() string {
	return "story_estimates"
}
