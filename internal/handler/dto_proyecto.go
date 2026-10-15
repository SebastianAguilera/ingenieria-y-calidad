package handler

import (
	"time"

	"ingenieria-y-calidad/internal/domain"
)

const formatoFecha = time.DateOnly

type usuarioDTO struct {
	Nombre string `json:"nombre"`
	Email  string `json:"email"`
}

type altaProyectoDTO struct {
	Nombre      string       `json:"nombre"`
	FechaInicio string       `json:"fecha_inicio"`
	FechaFin    *string      `json:"fecha_fin"`
	Usuarios    []usuarioDTO `json:"usuarios"`
}

type actualizacionProyectoDTO struct {
	Nombre      string  `json:"nombre"`
	FechaInicio string  `json:"fecha_inicio"`
	FechaFin    *string `json:"fecha_fin"`
}

type cambioEstadoDTO struct {
	Estado domain.EstadoProyecto `json:"estado"`
}

type respuestaUsuario struct {
	ID     uint   `json:"id"`
	Nombre string `json:"nombre"`
	Email  string `json:"email"`

	ProyectoID uint   `json:"proyecto_id,omitempty"`
	CreadoEn   string `json:"creado_en,omitempty"`
}

type respuestaProyecto struct {
	ID               uint                  `json:"id"`
	Nombre           string                `json:"nombre"`
	FechaInicio      string                `json:"fecha_inicio"`
	FechaFin         *string               `json:"fecha_fin"`
	Estado           domain.EstadoProyecto `json:"estado"`
	CantidadUsuarios int                   `json:"cantidad_usuarios"`
	Usuarios         []respuestaUsuario    `json:"usuarios"`
	CreadoEn         string                `json:"creado_en"`
	ActualizadoEn    string                `json:"actualizado_en"`
}

type respuestaProyectoListado struct {
	ID               uint                  `json:"id"`
	Nombre           string                `json:"nombre"`
	FechaInicio      string                `json:"fecha_inicio"`
	FechaFin         *string               `json:"fecha_fin"`
	Estado           domain.EstadoProyecto `json:"estado"`
	CantidadUsuarios int                   `json:"cantidad_usuarios"`
	CreadoEn         string                `json:"creado_en"`
	ActualizadoEn    string                `json:"actualizado_en"`
}

type respuestaListadoProyectos struct {
	Total     int                        `json:"total"`
	Proyectos []respuestaProyectoListado `json:"proyectos"`
}

type altaUsuarioDTO struct {
	Nombre string `json:"nombre"`
	Email  string `json:"email"`
}

func formatearFecha(t time.Time) string {
	return t.Format(formatoFecha)
}

func formatearFechaOpcional(t *time.Time) *string {
	if t == nil {
		return nil
	}
	formateada := formatearFecha(*t)
	return &formateada
}

func formatearTimestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func construirRespuestaProyecto(p *domain.Proyecto) respuestaProyecto {
	usuarios := construirRespuestaUsuarios(p.Usuarios)
	return respuestaProyecto{
		ID:               p.ID,
		Nombre:           p.Nombre,
		FechaInicio:      formatearFecha(p.FechaInicio),
		FechaFin:         formatearFechaOpcional(p.FechaFin),
		Estado:           p.Estado,
		CantidadUsuarios: len(usuarios),
		Usuarios:         usuarios,
		CreadoEn:         formatearTimestamp(p.CreadoEn),
		ActualizadoEn:    formatearTimestamp(p.ActualizadoEn),
	}
}

func construirRespuestaListado(p *domain.Proyecto) respuestaProyectoListado {
	detalle := construirRespuestaProyecto(p)
	return respuestaProyectoListado{
		ID:               detalle.ID,
		Nombre:           detalle.Nombre,
		FechaInicio:      detalle.FechaInicio,
		FechaFin:         detalle.FechaFin,
		Estado:           detalle.Estado,
		CantidadUsuarios: detalle.CantidadUsuarios,
		CreadoEn:         detalle.CreadoEn,
		ActualizadoEn:    detalle.ActualizadoEn,
	}
}

func construirRespuestaUsuarios(usuarios []domain.Usuario) []respuestaUsuario {
	respuestas := make([]respuestaUsuario, 0, len(usuarios))
	for _, usuario := range usuarios {
		respuestas = append(respuestas, respuestaUsuario{
			ID:       usuario.ID,
			Nombre:   usuario.Nombre,
			Email:    usuario.Email,
			CreadoEn: formatearTimestamp(usuario.CreadoEn),
		})
	}
	return respuestas
}

func construirRespuestaVinculo(u *domain.Usuario, proyectoID uint) respuestaUsuario {
	return respuestaUsuario{
		ID:         u.ID,
		Nombre:     u.Nombre,
		Email:      u.Email,
		ProyectoID: proyectoID,
		CreadoEn:   formatearTimestamp(u.CreadoEn),
	}
}
