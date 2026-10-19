package models

import "time"

// SprintModel representa la tabla "sprints" en la base de datos PostgreSQL.
// El indice unico compuesto por (proyecto_id, identificador) garantiza que el
// identificador visible del Sprint no se repita dentro de un mismo proyecto.
type SprintModel struct {
	ID            uint          `gorm:"column:id;primaryKey;autoIncrement"`
	ProyectoID    uint          `gorm:"column:proyecto_id;not null;uniqueIndex:ux_sprints_proyecto_identificador,priority:1"`
	Identificador string        `gorm:"column:identificador;type:varchar(50);not null;uniqueIndex:ux_sprints_proyecto_identificador,priority:2"`
	SprintGoal    string        `gorm:"column:sprint_goal;type:text;not null"`
	Estado        string        `gorm:"column:estado;type:varchar(20);not null;default:'Planificado'"`
	FechaInicio   time.Time     `gorm:"column:fecha_inicio;type:timestamptz;not null"`
	FechaFin      time.Time     `gorm:"column:fecha_fin;type:timestamptz;not null"`
	CreadoEn      time.Time     `gorm:"column:creado_en;type:timestamptz;not null;default:now()"`
	ActualizadoEn time.Time     `gorm:"column:actualizado_en;type:timestamptz;not null;default:now()"`
	Proyecto      ProyectoModel `gorm:"foreignKey:ProyectoID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// TableName define el nombre de la tabla en PostgreSQL.
func (SprintModel) TableName() string {
	return "sprints"
}
