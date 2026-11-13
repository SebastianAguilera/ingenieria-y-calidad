package handler

import "ingenieria-y-calidad/internal/domain"

// respuestaReferenciaUsuario es el creador de una sesión en las respuestas.
type respuestaReferenciaUsuario struct {
	ID     uint   `json:"id"`
	Nombre string `json:"nombre"`
}

// respuestaVotoPoker es un voto revelado con el nombre del votante.
type respuestaVotoPoker struct {
	UsuarioID uint    `json:"usuario_id"`
	Nombre    string  `json:"nombre"`
	Puntos    float64 `json:"puntos"`
}

// respuestaHistorialSesion es un ítem del historial de sesiones de una historia.
type respuestaHistorialSesion struct {
	ID              uint                       `json:"id"`
	Estado          domain.EstadoSesionPoker   `json:"estado"`
	RondaActual     int                        `json:"ronda_actual"`
	PuntosAcordados *float64                   `json:"puntos_acordados"`
	CreadoPor       respuestaReferenciaUsuario `json:"creado_por"`
	CreadoEn        string                     `json:"creado_en"`
	CerradoEn       *string                    `json:"cerrado_en,omitempty"`
}

// respuestaDetalleSesion es el detalle de una sesión. En estado Votando, la
// lista votos queda vacía y solo se informan los puntos propios (RB-08).
type respuestaDetalleSesion struct {
	ID               uint                       `json:"id"`
	HistoriaID       uint                       `json:"historia_id"`
	Estado           domain.EstadoSesionPoker   `json:"estado"`
	RondaActual      int                        `json:"ronda_actual"`
	CreadoPor        respuestaReferenciaUsuario `json:"creado_por"`
	CreadoEn         string                     `json:"creado_en"`
	CerradoEn        *string                    `json:"cerrado_en,omitempty"`
	PuntosAcordados  *float64                   `json:"puntos_acordados"`
	TotalMiembros    int                        `json:"total_miembros"`
	VotosRegistrados int                        `json:"votos_registrados"`
	YaVote           bool                       `json:"ya_vote"`
	MiVoto           *float64                   `json:"mi_voto"`
	Votos            []respuestaVotoPoker       `json:"votos"`
}

// respuestaResumenHistoria es la historia asociada a una sesión activa.
type respuestaResumenHistoria struct {
	ID     uint   `json:"id"`
	Codigo string `json:"codigo"`
	Titulo string `json:"titulo"`
}

// respuestaSesionActiva es un ítem de GET .../poker-sessions/activas. Nunca
// incluye votos, solo conteos (RB-11).
type respuestaSesionActiva struct {
	SesionID         uint                       `json:"sesion_id"`
	Estado           domain.EstadoSesionPoker   `json:"estado"`
	RondaActual      int                        `json:"ronda_actual"`
	CreadoPor        respuestaReferenciaUsuario `json:"creado_por"`
	CreadoEn         string                     `json:"creado_en"`
	VotosRegistrados int                        `json:"votos_registrados"`
	TotalMiembros    int                        `json:"total_miembros"`
	YaVote           bool                       `json:"ya_vote"`
	Historia         respuestaResumenHistoria   `json:"historia"`
}

func construirRespuestaHistorialSesion(s *domain.SesionPoker) respuestaHistorialSesion {
	return respuestaHistorialSesion{
		ID:              s.ID,
		Estado:          s.Estado,
		RondaActual:     s.RondaActual,
		PuntosAcordados: s.PuntosAcordados,
		CreadoPor:       respuestaReferenciaUsuario{ID: s.CreadoPor, Nombre: s.CreadorNombre},
		CreadoEn:        formatearTimestamp(s.CreadoEn),
		CerradoEn:       formatearTimestampOpcional(s.CerradoEn),
	}
}

func construirRespuestaDetalleSesion(d *domain.DetalleSesionPoker) respuestaDetalleSesion {
	votos := make([]respuestaVotoPoker, 0, len(d.Votos))
	for _, voto := range d.Votos {
		votos = append(votos, respuestaVotoPoker{
			UsuarioID: voto.UsuarioID,
			Nombre:    voto.Nombre,
			Puntos:    voto.Puntos,
		})
	}

	return respuestaDetalleSesion{
		ID:               d.Sesion.ID,
		HistoriaID:       d.Sesion.HistoriaID,
		Estado:           d.Sesion.Estado,
		RondaActual:      d.Sesion.RondaActual,
		CreadoPor:        respuestaReferenciaUsuario{ID: d.Creador.ID, Nombre: d.Creador.Nombre},
		CreadoEn:         formatearTimestamp(d.Sesion.CreadoEn),
		CerradoEn:        formatearTimestampOpcional(d.Sesion.CerradoEn),
		PuntosAcordados:  d.Sesion.PuntosAcordados,
		TotalMiembros:    d.TotalMiembros,
		VotosRegistrados: d.VotosRegistrados,
		YaVote:           d.YaVote,
		MiVoto:           d.MiVoto,
		Votos:            votos,
	}
}

func construirRespuestaSesionActiva(a *domain.SesionPokerActiva) respuestaSesionActiva {
	return respuestaSesionActiva{
		SesionID:         a.Sesion.ID,
		Estado:           a.Sesion.Estado,
		RondaActual:      a.Sesion.RondaActual,
		CreadoPor:        respuestaReferenciaUsuario{ID: a.Creador.ID, Nombre: a.Creador.Nombre},
		CreadoEn:         formatearTimestamp(a.Sesion.CreadoEn),
		VotosRegistrados: a.VotosRegistrados,
		TotalMiembros:    a.TotalMiembros,
		YaVote:           a.YaVote,
		Historia: respuestaResumenHistoria{
			ID:     a.Historia.ID,
			Codigo: a.Historia.Codigo,
			Titulo: a.Historia.Titulo,
		},
	}
}

// respuestaVotoRegistrado es la respuesta de POST .../votos. Solo informa el
// voto propio y conteos: nunca votos ajenos (US-15, RB-08).
type respuestaVotoRegistrado struct {
	SesionID         uint    `json:"sesion_id"`
	MiVoto           float64 `json:"mi_voto"`
	VotadoEn         string  `json:"votado_en"`
	VotosRegistrados int     `json:"votos_registrados"`
	TotalMiembros    int     `json:"total_miembros"`
}

func construirRespuestaVotoRegistrado(r *domain.ResultadoVotoPoker) respuestaVotoRegistrado {
	return respuestaVotoRegistrado{
		SesionID:         r.SesionID,
		MiVoto:           r.Puntos,
		VotadoEn:         formatearTimestamp(r.VotadoEn),
		VotosRegistrados: r.VotosRegistrados,
		TotalMiembros:    r.TotalMiembros,
	}
}
