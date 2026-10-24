package handler

import (
	"ingenieria-y-calidad/internal/domain"
)

type respuestaUsuarioDTO struct {
	ID        uint      `json:"id"`
	Nombre    string    `json:"nombre"`
	Email     string    `json:"email"`
	CreadoEn  string    `json:"creado_en"`
}

func construirRespuestaUsuarioDTO(u *domain.Usuario) respuestaUsuarioDTO {
	return respuestaUsuarioDTO{
		ID:       u.ID,
		Nombre:   u.Nombre,
		Email:    u.Email,
		CreadoEn: formatearTimestamp(u.CreadoEn),
	}
}

func construirRespuestaUsuariosDTO(usuarios []domain.Usuario) []respuestaUsuarioDTO {
	respuestas := make([]respuestaUsuarioDTO, 0, len(usuarios))
	for _, usuario := range usuarios {
		respuestas = append(respuestas, construirRespuestaUsuarioDTO(&usuario))
	}
	return respuestas
}
