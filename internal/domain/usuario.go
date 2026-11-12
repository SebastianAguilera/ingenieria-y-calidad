package domain

import (
	"net/mail"
	"strings"
	"time"
)

const longitudMaximaEmail = 150

// Usuario identifica a una persona. El email es su clave natural y es
// único en todo el sistema.
type Usuario struct {
	ID       uint      `json:"id"`
	Nombre   string    `json:"nombre"`
	Email    string    `json:"email"`
	CreadoEn time.Time `json:"creado_en"`

	PasswordHash string `json:"-"`

	// Rol es el rol dentro del proyecto en el que se lo vincula. Solo se usa al
	// persistir el vínculo (alta de proyecto); fuera de ese contexto va vacío.
	Rol RolProyecto `json:"-"`

	Proyectos []Proyecto `json:"-"`
}

// UsuarioInput es el dato de entrada de un usuario en las operaciones de
// alta de proyecto y de asociación.
type UsuarioInput struct {
	Nombre string
	Email  string
	Rol    RolProyecto
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

type RolProyecto string

const (
	// RolLider gestiona el equipo y puede eliminar el proyecto (US-11).
	RolLider RolProyecto = "LIDER"
	// RolMiembro trabaja el contenido del proyecto sin gestionar el equipo.
	RolMiembro RolProyecto = "MIEMBRO"
)

// Valido indica si el rol es uno de los valores permitidos por el dominio.
func (r RolProyecto) Valido() bool {
	switch r {
	case RolLider, RolMiembro:
		return true
	default:
		return false
	}
}

// ParseRolProyecto convierte un string al tipo RolProyecto validando el valor.
func ParseRolProyecto(rol string) (RolProyecto, error) {
	r := RolProyecto(rol)
	if !r.Valido() {
		return "", ErrRolInvalido
	}
	return r, nil
}

// AsociacionUsuario describe la asociación de un usuario existente a un proyecto con su rol.
type AsociacionUsuario struct {
	UsuarioID uint
	Rol       RolProyecto
}
