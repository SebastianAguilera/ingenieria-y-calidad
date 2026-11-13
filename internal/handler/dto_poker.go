package handler

import "ingenieria-y-calidad/internal/domain"

// respuestaReferenciaUsuario es el creador de una sesión en las respuestas.
type respuestaReferenciaUsuario struct {
	ID     uint   `json:"id"`
	Nombre string `json:"nombre"`
}

// respuestaVotoPoker es un voto revelado con el nombre del votante y la ronda.
type respuestaVotoPoker struct {
	UsuarioID uint    `json:"usuario_id"`
	Nombre    string  `json:"usuario_nombre"`
	Puntos    float64 `json:"puntos"`
	Ronda     int     `json:"ronda"`
}

// respuestaDispersionPoker expone las estadísticas de dispersión de una ronda.
type respuestaDispersionPoker struct {
	Promedio           float64        `json:"promedio"`
	Mediana            float64        `json:"mediana"`
	Minimo             float64        `json:"minimo"`
	Maximo             float64        `json:"maximo"`
	Rango              float64        `json:"rango"`
	DesviacionEstandar float64        `json:"desviacion_estandar"`
	HayDispersion      bool           `json:"hay_dispersion"`
	TotalVotos         int            `json:"total_votos"`
	VotosPorValor      map[string]int `json:"votos_por_valor"`
}

// respuestaHistorialRondaPoker es una ronda del historial de la sesión.
type respuestaHistorialRondaPoker struct {
	Numero     int                       `json:"numero"`
	Estado     domain.EstadoSesionPoker  `json:"estado"`
	ReveladaEn *string                   `json:"revelada_en"`
	Votos      []respuestaVotoPoker      `json:"votos"`
	Dispersion *respuestaDispersionPoker `json:"dispersion,omitempty"`
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
	ID               uint                           `json:"id"`
	HistoriaID       uint                           `json:"historia_id"`
	Estado           domain.EstadoSesionPoker       `json:"estado"`
	RondaActual      int                            `json:"ronda_actual"`
	CreadoPor        respuestaReferenciaUsuario     `json:"creado_por"`
	CreadoEn         string                         `json:"creado_en"`
	CerradoEn        *string                        `json:"cerrado_en,omitempty"`
	PuntosAcordados  *float64                       `json:"puntos_acordados"`
	TotalMiembros    int                            `json:"total_miembros"`
	VotosRegistrados int                            `json:"votos_registrados"`
	YaVote           bool                           `json:"ya_vote"`
	MiVoto           *float64                       `json:"mi_voto"`
	Votos            []respuestaVotoPoker           `json:"votos"`
	Dispersion       *respuestaDispersionPoker      `json:"dispersion,omitempty"`
	HistorialRondas  []respuestaHistorialRondaPoker `json:"historial_rondas"`
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
	votos := construirRespuestaVotos(d.Votos)

	historial := make([]respuestaHistorialRondaPoker, 0, len(d.HistorialRondas))
	for _, ronda := range d.HistorialRondas {
		item := respuestaHistorialRondaPoker{
			Numero:     ronda.NumeroRonda,
			Estado:     ronda.Estado,
			ReveladaEn: formatearTimestampOpcional(ronda.ReveladaEn),
			Votos:      construirRespuestaVotos(ronda.Votos),
		}
		if ronda.Dispersion != nil {
			d := construirRespuestaDispersion(ronda.Dispersion)
			item.Dispersion = &d
		}
		historial = append(historial, item)
	}

	resp := respuestaDetalleSesion{
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
		HistorialRondas:  historial,
	}
	if d.Dispersion != nil {
		disp := construirRespuestaDispersion(d.Dispersion)
		resp.Dispersion = &disp
	}
	return resp
}

func construirRespuestaVotos(votos []domain.VotoPoker) []respuestaVotoPoker {
	resultado := make([]respuestaVotoPoker, 0, len(votos))
	for _, voto := range votos {
		resultado = append(resultado, respuestaVotoPoker{
			UsuarioID: voto.UsuarioID,
			Nombre:    voto.Nombre,
			Puntos:    voto.Puntos,
			Ronda:     voto.Ronda,
		})
	}
	return resultado
}

func construirRespuestaDispersion(d *domain.DispersionPoker) respuestaDispersionPoker {
	votos := d.VotosPorValor
	if votos == nil {
		votos = map[string]int{}
	}
	return respuestaDispersionPoker{
		Promedio:           d.Promedio,
		Mediana:            d.Mediana,
		Minimo:             d.Minimo,
		Maximo:             d.Maximo,
		Rango:              d.Rango,
		DesviacionEstandar: d.DesviacionEstandar,
		HayDispersion:      d.HayDispersion,
		TotalVotos:         d.TotalVotos,
		VotosPorValor:      votos,
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
