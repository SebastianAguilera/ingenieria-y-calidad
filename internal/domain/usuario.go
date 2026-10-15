package domain

import (
	"net/mail"
	"strings"
	"time"
)

const longitudMaximaEmail = 150

type Usuario struct {
	ID       uint      `json:"id"`
	Nombre   string    `json:"nombre"`
	Email    string    `json:"email"`
	CreadoEn time.Time `json:"creado_en"`

	Proyectos []Proyecto `json:"-"`
}

type UsuarioInput struct {
	Nombre string
	Email  string
}

func NormalizarEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

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
