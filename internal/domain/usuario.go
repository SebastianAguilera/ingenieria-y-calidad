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

	Proyectos []Proyecto `json:"-"`
}

// UsuarioInput es el dato de entrada de un usuario en las operaciones de
// alta de proyecto y de asociación.
type UsuarioInput struct {
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
