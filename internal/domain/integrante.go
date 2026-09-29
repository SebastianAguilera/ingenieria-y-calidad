package domain

import (
	"net/mail"
	"strings"
	"time"
)

const longitudMaximaEmail = 150

// Integrante identifica a una persona. El email es su clave natural y es
// único en todo el sistema.
type Integrante struct {
	ID       uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Nombre   string    `gorm:"column:nombre;type:varchar(100);not null" json:"nombre"`
	Email    string    `gorm:"column:email;type:varchar(150);not null;uniqueIndex" json:"email"`
	CreadoEn time.Time `gorm:"column:creado_en;type:timestamptz;not null;default:now()" json:"creado_en"`

	Proyectos []Proyecto `gorm:"many2many:proyecto_integrantes;joinForeignKey:IntegranteID;joinReferences:ProyectoID" json:"-"`
}

// IntegranteInput es el dato de entrada de un integrante en las operaciones de
// alta de proyecto y de asociación.
type IntegranteInput struct {
	Nombre string
	Email  string
}

// NormalizarEmail canonicaliza el email para que las comparaciones y el índice
// único sean insensibles a mayúsculas y espacios.
func NormalizarEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// EmailValido valida la forma del email sin exigir dominios reales y rechaza
// la forma "Nombre <correo>" que net/mail aceptaría.
func EmailValido(email string) bool {
	normalizado := NormalizarEmail(email)
	if normalizado == "" || len(normalizado) > longitudMaximaEmail || strings.ContainsAny(normalizado, " \t") {
		return false
	}
	direccion, err := mail.ParseAddress(normalizado)
	if err != nil {
		return false
	}
	return direccion.Address == normalizado
}
