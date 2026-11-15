package models

import "time"

// SprintBacklogModel representa la tabla "sprint_backlog_items" en PostgreSQL.
// Una historia solo puede pertenecer a un Sprint (unique en historia_id).
type SprintBacklogModel struct {
	ID              uint        `gorm:"column:id;primaryKey;autoIncrement"`
	ProyectoID      uint        `gorm:"column:proyecto_id;not null;index:idx_sprint_backlog_proyecto"`
	SprintID        uint        `gorm:"column:sprint_id;not null;index:idx_sprint_backlog_sprint"`
	HistoriaID      uint        `gorm:"column:historia_id;not null;uniqueIndex:ux_sprint_backlog_historia"`
	Orden           int         `gorm:"column:orden;not null;default:0"`
	FechaAsignacion time.Time   `gorm:"column:fecha_asignacion;type:timestamptz;not null"`
	CreadoEn        time.Time   `gorm:"column:creado_en;type:timestamptz;not null;default:now()"`
	ActualizadoEn   time.Time   `gorm:"column:actualizado_en;type:timestamptz;not null;default:now()"`
	Proyecto        ProyectoModel `gorm:"foreignKey:ProyectoID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Sprint          SprintModel   `gorm:"foreignKey:SprintID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Historia        HistoriaModel `gorm:"foreignKey:HistoriaID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// TableName define el nombre de la tabla en PostgreSQL.
func (SprintBacklogModel) TableName() string {
	return "sprint_backlog_items"
}
