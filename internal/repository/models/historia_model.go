package models

import "time"

// HistoriaModel representa la tabla "historias" en la base de datos PostgreSQL.
// El indice unico compuesto por (proyecto_id, numero) garantiza que el codigo
// consecutivo US-NN no se repita dentro de un mismo proyecto.
type HistoriaModel struct {
	ID            uint          `gorm:"column:id;primaryKey;autoIncrement"`
	ProyectoID    uint          `gorm:"column:proyecto_id;not null;uniqueIndex:ux_historias_proyecto_numero,priority:1"`
	Numero        int           `gorm:"column:numero;not null;uniqueIndex:ux_historias_proyecto_numero,priority:2"`
	Codigo        string        `gorm:"column:codigo;type:varchar(20);not null"`
	Titulo        string        `gorm:"column:titulo;type:varchar(200);not null"`
	Descripcion   string        `gorm:"column:descripcion;type:text"`
	Prioridad     string        `gorm:"column:prioridad;type:varchar(20);not null;default:'Media'"`
	Estado        string        `gorm:"column:estado;type:varchar(20);not null;default:'Backlog'"`
	StoryPoints   *int          `gorm:"column:story_points"`
	CreadoEn      time.Time     `gorm:"column:creado_en;type:timestamptz;not null;default:now()"`
	ActualizadoEn time.Time     `gorm:"column:actualizado_en;type:timestamptz;not null;default:now()"`
	Proyecto      ProyectoModel `gorm:"foreignKey:ProyectoID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// TableName define el nombre de la tabla en PostgreSQL.
func (HistoriaModel) TableName() string {
	return "historias"
}
