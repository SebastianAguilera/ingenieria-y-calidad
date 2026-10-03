package models

import "time"

// UsuarioModel representa la tabla "usuarios" en la base de datos PostgreSQL.
type UsuarioModel struct {
	ID        uint            `gorm:"column:id;primaryKey;autoIncrement"`
	Nombre    string          `gorm:"column:nombre;type:varchar(100);not null"`
	Email     string          `gorm:"column:email;type:varchar(150);not null;uniqueIndex"`
	CreadoEn  time.Time       `gorm:"column:creado_en;type:timestamptz;not null;default:now()"`
	Proyectos []ProyectoModel `gorm:"many2many:proyecto_usuarios;joinForeignKey:UsuarioID;joinReferences:ProyectoID"`
}

// TableName define el nombre de la tabla en PostgreSQL.
func (UsuarioModel) TableName() string {
	return "usuarios"
}

// ProyectoUsuarioModel modela la tabla puente de la relación many-to-many entre proyectos y usuarios.
type ProyectoUsuarioModel struct {
	ProyectoID uint `gorm:"column:proyecto_id;primaryKey"`
	UsuarioID  uint `gorm:"column:usuario_id;primaryKey"`
}

// TableName define el nombre de la tabla puente en PostgreSQL.
func (ProyectoUsuarioModel) TableName() string {
	return "proyecto_usuarios"
}
