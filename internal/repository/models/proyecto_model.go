package models

import "time"

// ProyectoModel representa la tabla "proyectos" en la base de datos PostgreSQL.
type ProyectoModel struct {
	ID            uint                   `gorm:"column:id;primaryKey;autoIncrement"`
	Nombre        string                 `gorm:"column:nombre;type:varchar(150);not null;index"`
	FechaInicio   time.Time              `gorm:"column:fecha_inicio;type:date;not null"`
	FechaFin      *time.Time             `gorm:"column:fecha_fin;type:date"`
	Estado        string                 `gorm:"column:estado;type:varchar(20);not null;default:'Activo';index"`
	BorradoEn     *time.Time             `gorm:"column:borrado_en;type:timestamptz;index"`
	CreadoEn      time.Time              `gorm:"column:creado_en;type:timestamptz;not null;default:now()"`
	ActualizadoEn time.Time              `gorm:"column:actualizado_en;type:timestamptz;not null;default:now()"`
	Usuarios      []UsuarioModel         `gorm:"many2many:proyecto_usuarios;joinForeignKey:ProyectoID;joinReferences:UsuarioID"`
}

// TableName define el nombre de la tabla en PostgreSQL.
func (ProyectoModel) TableName() string {
	return "proyectos"
}
