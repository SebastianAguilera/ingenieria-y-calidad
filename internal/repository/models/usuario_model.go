package models

import "time"

type UsuarioModel struct {
	ID        uint            `gorm:"column:id;primaryKey;autoIncrement"`
	Nombre    string          `gorm:"column:nombre;type:varchar(100);not null"`
	Email     string          `gorm:"column:email;type:varchar(150);not null;uniqueIndex"`
	CreadoEn  time.Time       `gorm:"column:creado_en;type:timestamptz;not null;default:now()"`
	Proyectos []ProyectoModel `gorm:"many2many:proyecto_usuarios;joinForeignKey:UsuarioID;joinReferences:ProyectoID"`
}

func (UsuarioModel) TableName() string {
	return "usuarios"
}

type ProyectoUsuarioModel struct {
	ProyectoID uint `gorm:"column:proyecto_id;primaryKey"`
	UsuarioID  uint `gorm:"column:usuario_id;primaryKey"`
}

func (ProyectoUsuarioModel) TableName() string {
	return "proyecto_usuarios"
}
