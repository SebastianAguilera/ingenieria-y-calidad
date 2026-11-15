package models

import "time"

// DefectoModel representa la tabla "defectos" en la base de datos PostgreSQL.
// Un defecto se asocia a la historia donde se detectó y al Sprint de detección;
// el Sprint de resolución y la fecha de resolución solo se completan cuando el
// defecto se resuelve o se cierra.
type DefectoModel struct {
	ID                 uint          `gorm:"column:id;primaryKey;autoIncrement"`
	ProyectoID         uint          `gorm:"column:proyecto_id;not null;index:idx_defectos_proyecto"`
	HistoriaID         uint          `gorm:"column:historia_id;not null;index:idx_defectos_historia"`
	SprintDeteccionID  uint          `gorm:"column:sprint_deteccion_id;not null;index:idx_defectos_sprint_deteccion"`
	SprintResolucionID *uint         `gorm:"column:sprint_resolucion_id;index:idx_defectos_sprint_resolucion"`
	ReportadoPor       *uint         `gorm:"column:reportado_por"`
	Descripcion        string        `gorm:"column:descripcion;type:text;not null"`
	Severidad          string        `gorm:"column:severidad;type:varchar(20);not null;default:'minor'"`
	Estado             string        `gorm:"column:estado;type:varchar(20);not null;default:'open'"`
	CreadoEn           time.Time     `gorm:"column:creado_en;type:timestamptz;not null;default:now()"`
	ResueltoEn         *time.Time    `gorm:"column:resuelto_en;type:timestamptz"`
	Proyecto           ProyectoModel `gorm:"foreignKey:ProyectoID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Historia           HistoriaModel `gorm:"foreignKey:HistoriaID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	SprintDeteccion    SprintModel   `gorm:"foreignKey:SprintDeteccionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	SprintResolucion   *SprintModel  `gorm:"foreignKey:SprintResolucionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Reportado          *UsuarioModel `gorm:"foreignKey:ReportadoPor;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

// TableName define el nombre de la tabla en PostgreSQL.
func (DefectoModel) TableName() string {
	return "defectos"
}
