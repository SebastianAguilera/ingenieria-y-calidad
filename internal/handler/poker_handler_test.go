package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func nuevoRouterDePoker(t *testing.T, usuarioID uint) (*gin.Engine, *mockServicioPoker) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	servicio := nuevoMockServicioPoker(t)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		if usuarioID != 0 {
			c.Set(claveContextoUsuario, usuarioID)
		}
		c.Next()
	})
	RegistrarRutasPoker(router, servicio)
	return router, servicio
}

func detallePokerDeEjemplo(estado domain.EstadoSesionPoker) *domain.DetalleSesionPoker {
	creado := time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC)
	return &domain.DetalleSesionPoker{
		Sesion: domain.SesionPoker{
			ID:            9,
			ProyectoID:    1,
			HistoriaID:    5,
			CreadoPor:     2,
			RondaActual:   1,
			Estado:        estado,
			CreadoEn:      creado,
			ActualizadoEn: creado,
		},
		Creador:          domain.ReferenciaUsuario{ID: 2, Nombre: "Ana"},
		TotalMiembros:    3,
		VotosRegistrados: 0,
		YaVote:           false,
		Votos:            []domain.VotoPoker{},
		HistorialRondas:  []domain.HistorialRondaPoker{},
	}
}

func decodificarDetallePoker(t *testing.T, w *httptest.ResponseRecorder) respuestaDetalleSesion {
	t.Helper()
	var respuesta respuestaDetalleSesion
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta), "cuerpo: %s", w.Body.String())
	return respuesta
}

func TestCrearSesionPokerHandler(t *testing.T) {
	t.Run("E-01 crea la sesión", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("CrearSesion", mock.Anything, uint(1), uint(5), uint(2)).
			Return(detallePokerDeEjemplo(domain.EstadoSesionVotando), nil)

		w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/historias/5/poker-sessions", nil)

		require.Equal(t, http.StatusCreated, w.Code)
		respuesta := decodificarDetallePoker(t, w)
		assert.Equal(t, uint(9), respuesta.ID)
		assert.Equal(t, domain.EstadoSesionVotando, respuesta.Estado)
		assert.Equal(t, 1, respuesta.RondaActual)
		assert.Equal(t, "Ana", respuesta.CreadoPor.Nombre)
		assert.NotNil(t, respuesta.Votos)
		assert.Empty(t, respuesta.Votos)
	})

	t.Run("E-13 sesión activa existente", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("CrearSesion", mock.Anything, uint(1), uint(5), uint(2)).
			Return(nil, domain.ErrSesionPokerActiva)

		w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/historias/5/poker-sessions", nil)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Equal(t, codigoSesionPokerActiva, decodificarError(t, w).Error)
	})

	t.Run("E-14 historia en Hecho", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("CrearSesion", mock.Anything, uint(1), uint(5), uint(2)).
			Return(nil, domain.ErrHistoriaNoEstimable)

		w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/historias/5/poker-sessions", nil)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Equal(t, codigoHistoriaNoEstimable, decodificarError(t, w).Error)
	})

	t.Run("E-16 historia inexistente", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("CrearSesion", mock.Anything, uint(1), uint(5), uint(2)).
			Return(nil, domain.ErrHistoriaNoEncontrada)

		w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/historias/5/poker-sessions", nil)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, codigoHistoriaNoEncontrada, decodificarError(t, w).Error)
	})

	t.Run("sin identidad devuelve 401", func(t *testing.T) {
		router, _ := nuevoRouterDePoker(t, 0)

		w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/historias/5/poker-sessions", nil)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Equal(t, codigoTokenInvalido, decodificarError(t, w).Error)
	})

	t.Run("id inválido devuelve 400", func(t *testing.T) {
		router, _ := nuevoRouterDePoker(t, 2)

		w := ejecutar(t, router, http.MethodPost, "/api/proyectos/abc/historias/5/poker-sessions", nil)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, codigoParametroInvalido, decodificarError(t, w).Error)
	})
}

func TestListarHistorialPokerHandler(t *testing.T) {
	t.Run("E-12 y E-16 incluye cerradas con estimación y canceladas", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		creado := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
		cancelado := time.Date(2026, 11, 12, 9, 0, 0, 0, time.UTC)
		acordados := 5.0
		servicio.On("ListarPorHistoria", mock.Anything, uint(1), uint(5)).Return([]domain.SesionPoker{
			{
				ID: 3, Estado: domain.EstadoSesionCerrada, RondaActual: 2, CreadoPor: 2, CreadoEn: creado,
				PuntosAcordados: &acordados,
				StoryEstimate: &domain.StoryEstimate{
					ID: 12, HistoriaID: 5, SesionID: 3, RondaID: 4, Puntos: 5,
					Fuente: "planning_poker", EstimadoPor: punteroUintPoker(2), EstimadoEn: creado,
				},
			},
			{
				ID: 2, Estado: domain.EstadoSesionCancelada, RondaActual: 1, CreadoPor: 2, CreadoEn: creado,
				CanceladoEn: &cancelado,
			},
		}, nil)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/1/historias/5/poker-sessions", nil)

		require.Equal(t, http.StatusOK, w.Code)
		var respuestas []respuestaHistorialSesion
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuestas))
		require.Len(t, respuestas, 2)
		assert.Equal(t, domain.EstadoSesionCerrada, respuestas[0].Estado)
		require.NotNil(t, respuestas[0].StoryEstimate)
		assert.Equal(t, 5.0, respuestas[0].StoryEstimate.Puntos)
		assert.Equal(t, uint(3), respuestas[0].StoryEstimate.SesionID)
		assert.Equal(t, uint(4), respuestas[0].StoryEstimate.RondaID)
		assert.Equal(t, domain.EstadoSesionCancelada, respuestas[1].Estado)
		require.NotNil(t, respuestas[1].CanceladoEn)
		assert.Nil(t, respuestas[1].StoryEstimate)
	})

	t.Run("sin sesiones responde []", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("ListarPorHistoria", mock.Anything, uint(1), uint(5)).
			Return([]domain.SesionPoker{}, nil)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/1/historias/5/poker-sessions", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "[]", strings.TrimSpace(w.Body.String()))
	})
}

func TestDetallePokerHandler(t *testing.T) {
	t.Run("E-11 en Votando no expone votos ajenos", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		detalle := detallePokerDeEjemplo(domain.EstadoSesionVotando)
		detalle.VotosRegistrados = 1
		detalle.YaVote = true
		miVoto := 3.0
		detalle.MiVoto = &miVoto
		servicio.On("Detalle", mock.Anything, uint(1), uint(9), uint(2)).Return(detalle, nil)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/1/poker-sessions/9", nil)

		require.Equal(t, http.StatusOK, w.Code)
		respuesta := decodificarDetallePoker(t, w)
		assert.True(t, respuesta.YaVote)
		require.NotNil(t, respuesta.MiVoto)
		assert.Equal(t, 3.0, *respuesta.MiVoto)
		assert.Empty(t, respuesta.Votos)
		assert.Equal(t, 3, respuesta.TotalMiembros)
	})

	t.Run("E-05 en Revelada expone votos", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		detalle := detallePokerDeEjemplo(domain.EstadoSesionRevelada)
		detalle.Votos = []domain.VotoPoker{
			{UsuarioID: 2, Nombre: "Ana", Puntos: 3},
			{UsuarioID: 7, Nombre: "Ben", Puntos: 5},
		}
		servicio.On("Detalle", mock.Anything, uint(1), uint(9), uint(2)).Return(detalle, nil)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/1/poker-sessions/9", nil)

		require.Equal(t, http.StatusOK, w.Code)
		respuesta := decodificarDetallePoker(t, w)
		require.Len(t, respuesta.Votos, 2)
		assert.Equal(t, "Ben", respuesta.Votos[1].Nombre)
		assert.Equal(t, 5.0, respuesta.Votos[1].Puntos)
	})

	t.Run("E-15 en Cerrada incluye story_estimate completo", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		detalle := detallePokerDeEjemplo(domain.EstadoSesionCerrada)
		acordados := 5.0
		estimadoEn := time.Date(2026, 11, 13, 18, 0, 0, 0, time.UTC)
		detalle.Sesion.PuntosAcordados = &acordados
		detalle.Sesion.CerradoEn = &estimadoEn
		detalle.Sesion.StoryEstimate = &domain.StoryEstimate{
			ID: 12, HistoriaID: 5, SesionID: 9, RondaID: 3, Puntos: 5,
			Fuente: "planning_poker", EstimadoPor: punteroUintPoker(2), EstimadoEn: estimadoEn,
		}
		servicio.On("Detalle", mock.Anything, uint(1), uint(9), uint(2)).Return(detalle, nil)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/1/poker-sessions/9", nil)

		require.Equal(t, http.StatusOK, w.Code)
		respuesta := decodificarDetallePoker(t, w)
		require.NotNil(t, respuesta.StoryEstimate)
		assert.Equal(t, 5.0, respuesta.StoryEstimate.Puntos)
		assert.Equal(t, uint(9), respuesta.StoryEstimate.SesionID)
		assert.Equal(t, uint(3), respuesta.StoryEstimate.RondaID)
		require.NotNil(t, respuesta.StoryEstimate.EstimadoPor)
		assert.Equal(t, uint(2), *respuesta.StoryEstimate.EstimadoPor)
	})

	t.Run("E-17 sesión inexistente", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("Detalle", mock.Anything, uint(1), uint(9), uint(2)).
			Return(nil, domain.ErrSesionPokerNoEncontrada)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/1/poker-sessions/9", nil)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, codigoSesionPokerNoEncontrada, decodificarError(t, w).Error)
	})
}

func TestActivasPokerHandler(t *testing.T) {
	t.Run("E-04 devuelve conteos y resumen", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		creado := time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC)
		servicio.On("Activas", mock.Anything, uint(1), uint(2)).Return([]domain.SesionPokerActiva{
			{
				Sesion: domain.SesionPoker{
					ID: 9, ProyectoID: 1, HistoriaID: 5, CreadoPor: 2, CreadorNombre: "Ana",
					RondaActual: 1, Estado: domain.EstadoSesionRevelada, CreadoEn: creado,
				},
				Creador:          domain.ReferenciaUsuario{ID: 2, Nombre: "Ana"},
				Historia:         domain.ResumenHistoria{ID: 5, Codigo: "US-05", Titulo: "Estimar"},
				TotalMiembros:    3,
				VotosRegistrados: 2,
				YaVote:           true,
			},
		}, nil)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/1/poker-sessions/activas", nil)

		require.Equal(t, http.StatusOK, w.Code)
		var respuestas []respuestaSesionActiva
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuestas))
		require.Len(t, respuestas, 1)
		assert.Equal(t, "US-05", respuestas[0].Historia.Codigo)
		assert.Equal(t, 3, respuestas[0].TotalMiembros)
		assert.True(t, respuestas[0].YaVote)
		assert.NotContains(t, w.Body.String(), `"votos"`, "activas nunca expone votos")
	})

	t.Run("E-08 sin activas responde []", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("Activas", mock.Anything, uint(1), uint(2)).
			Return([]domain.SesionPokerActiva{}, nil)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/1/poker-sessions/activas", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "[]", strings.TrimSpace(w.Body.String()))
	})
}

func TestRevelarPokerHandler(t *testing.T) {
	const ruta = "/api/proyectos/1/poker-sessions/9/revelar"

	reveladaEn := time.Date(2026, 11, 13, 18, 0, 0, 0, time.UTC)
	detalleRevelado := func() *domain.DetalleSesionPoker {
		d := detallePokerDeEjemplo(domain.EstadoSesionRevelada)
		miVoto := 1.0
		d.MiVoto = &miVoto
		d.YaVote = true
		d.VotosRegistrados = 3
		d.Votos = []domain.VotoPoker{
			{UsuarioID: 2, Nombre: "Ana", Puntos: 1, Ronda: 1},
			{UsuarioID: 7, Nombre: "Ben", Puntos: 3, Ronda: 1},
			{UsuarioID: 9, Nombre: "Cris", Puntos: 8, Ronda: 1},
		}
		disp := domain.CalcularDispersionPoker(d.Votos)
		d.Dispersion = &disp
		d.HistorialRondas = []domain.HistorialRondaPoker{
			{
				NumeroRonda: 1,
				Estado:      domain.EstadoSesionRevelada,
				ReveladaEn:  &reveladaEn,
				Votos:       d.Votos,
				Dispersion:  &disp,
			},
		}
		return d
	}

	t.Run("E-01 revela y devuelve votos y dispersión", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("Revelar", mock.Anything, uint(1), uint(9), uint(2)).Return(detalleRevelado(), nil)

		w := ejecutar(t, router, http.MethodPost, ruta, nil)

		require.Equal(t, http.StatusOK, w.Code)
		respuesta := decodificarDetallePoker(t, w)
		assert.Equal(t, domain.EstadoSesionRevelada, respuesta.Estado)
		require.Len(t, respuesta.Votos, 3)
		assert.Equal(t, "Ana", respuesta.Votos[0].Nombre)
		assert.Equal(t, 1.0, respuesta.Votos[0].Puntos)
		assert.Equal(t, 1, respuesta.Votos[0].Ronda)
		require.NotNil(t, respuesta.Dispersion)
		assert.Equal(t, 4.0, respuesta.Dispersion.Promedio)
		assert.Equal(t, 3.0, respuesta.Dispersion.Mediana)
		assert.Equal(t, 7.0, respuesta.Dispersion.Rango)
		assert.True(t, respuesta.Dispersion.HayDispersion)
		assert.Equal(t, 3, respuesta.Dispersion.TotalVotos)
		require.Len(t, respuesta.HistorialRondas, 1)
		assert.Equal(t, domain.EstadoSesionRevelada, respuesta.HistorialRondas[0].Estado)
		assert.NotNil(t, respuesta.HistorialRondas[0].Dispersion)
	})

	t.Run("E-02 sin votos responde 422 SESION_POKER_SIN_VOTOS", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("Revelar", mock.Anything, uint(1), uint(9), uint(2)).Return(nil, domain.ErrSesionPokerSinVotos)

		w := ejecutar(t, router, http.MethodPost, ruta, nil)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		assert.Equal(t, codigoSesionPokerSinVotos, decodificarError(t, w).Error)
	})

	t.Run("E-03 fuera de Votando responde 409 SESION_POKER_NO_VOTANDO", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("Revelar", mock.Anything, uint(1), uint(9), uint(2)).Return(nil, domain.ErrSesionPokerNoVotando)

		w := ejecutar(t, router, http.MethodPost, ruta, nil)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Equal(t, codigoSesionPokerNoVotando, decodificarError(t, w).Error)
	})

	t.Run("sesión inexistente responde 404", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("Revelar", mock.Anything, uint(1), uint(9), uint(2)).Return(nil, domain.ErrSesionPokerNoEncontrada)

		w := ejecutar(t, router, http.MethodPost, ruta, nil)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, codigoSesionPokerNoEncontrada, decodificarError(t, w).Error)
	})

	t.Run("sesionId inválido responde 400", func(t *testing.T) {
		router, _ := nuevoRouterDePoker(t, 2)

		w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/poker-sessions/abc/revelar", nil)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, codigoParametroInvalido, decodificarError(t, w).Error)
	})

	t.Run("sin identidad responde 401", func(t *testing.T) {
		router, _ := nuevoRouterDePoker(t, 0)

		w := ejecutar(t, router, http.MethodPost, ruta, nil)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestNuevaRondaPokerHandler(t *testing.T) {
	const ruta = "/api/proyectos/1/poker-sessions/9/rondas"

	reveladaEn := time.Date(2026, 11, 13, 18, 0, 0, 0, time.UTC)
	votosRonda1 := []domain.VotoPoker{
		{UsuarioID: 2, Nombre: "Ana", Puntos: 3, Ronda: 1},
		{UsuarioID: 7, Nombre: "Ben", Puntos: 5, Ronda: 1},
	}
	disp := domain.CalcularDispersionPoker(votosRonda1)
	detalleNuevaRonda := func() *domain.DetalleSesionPoker {
		d := detallePokerDeEjemplo(domain.EstadoSesionVotando)
		d.Sesion.RondaActual = 2
		d.Votos = []domain.VotoPoker{}
		d.HistorialRondas = []domain.HistorialRondaPoker{
			{
				NumeroRonda: 1,
				Estado:      domain.EstadoSesionRevelada,
				ReveladaEn:  &reveladaEn,
				Votos:       votosRonda1,
				Dispersion:  &disp,
			},
			{
				NumeroRonda: 2,
				Estado:      domain.EstadoSesionVotando,
				Votos:       []domain.VotoPoker{},
			},
		}
		return d
	}

	t.Run("E-08 abre nueva ronda y conserva historial", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("NuevaRonda", mock.Anything, uint(1), uint(9), uint(2)).Return(detalleNuevaRonda(), nil)

		w := ejecutar(t, router, http.MethodPost, ruta, nil)

		require.Equal(t, http.StatusOK, w.Code)
		respuesta := decodificarDetallePoker(t, w)
		assert.Equal(t, domain.EstadoSesionVotando, respuesta.Estado)
		assert.Equal(t, 2, respuesta.RondaActual)
		assert.Empty(t, respuesta.Votos)
		assert.Nil(t, respuesta.Dispersion)
		require.Len(t, respuesta.HistorialRondas, 2)
		assert.Equal(t, domain.EstadoSesionRevelada, respuesta.HistorialRondas[0].Estado)
		assert.NotNil(t, respuesta.HistorialRondas[0].Dispersion)
		assert.Equal(t, domain.EstadoSesionVotando, respuesta.HistorialRondas[1].Estado)
		assert.Empty(t, respuesta.HistorialRondas[1].Votos)
		assert.Nil(t, respuesta.HistorialRondas[1].Dispersion)
	})

	t.Run("E-09 fuera de Revelada responde 409 SESION_POKER_NO_REVELADA", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("NuevaRonda", mock.Anything, uint(1), uint(9), uint(2)).Return(nil, domain.ErrSesionPokerNoRevelada)

		w := ejecutar(t, router, http.MethodPost, ruta, nil)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Equal(t, codigoSesionPokerNoRevelada, decodificarError(t, w).Error)
	})

	t.Run("sesión inexistente responde 404", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("NuevaRonda", mock.Anything, uint(1), uint(9), uint(2)).Return(nil, domain.ErrSesionPokerNoEncontrada)

		w := ejecutar(t, router, http.MethodPost, ruta, nil)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("sin identidad responde 401", func(t *testing.T) {
		router, _ := nuevoRouterDePoker(t, 0)

		w := ejecutar(t, router, http.MethodPost, ruta, nil)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func resultadoVotoDeEjemplo() *domain.ResultadoVotoPoker {
	return &domain.ResultadoVotoPoker{
		SesionID:         9,
		Puntos:           5,
		VotadoEn:         time.Date(2026, 11, 13, 18, 0, 0, 0, time.UTC),
		VotosRegistrados: 2,
		TotalMiembros:    3,
	}
}

func TestVotarPokerHandler(t *testing.T) {
	const ruta = "/api/proyectos/1/poker-sessions/9/votos"

	t.Run("E-01 registra el voto y responde solo conteos", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("Votar", mock.Anything, uint(1), uint(9), uint(2), 5.0).Return(resultadoVotoDeEjemplo(), nil)

		w := ejecutar(t, router, http.MethodPost, ruta, `{"puntos": 5}`)

		require.Equal(t, http.StatusOK, w.Code)
		var cuerpo map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cuerpo))
		assert.Equal(t, float64(9), cuerpo["sesion_id"])
		assert.Equal(t, 5.0, cuerpo["mi_voto"])
		assert.Equal(t, float64(2), cuerpo["votos_registrados"])
		assert.Equal(t, float64(3), cuerpo["total_miembros"])
		assert.NotEmpty(t, cuerpo["votado_en"])
		assert.NotContains(t, cuerpo, "votos", "la respuesta nunca incluye votos ajenos")
	})

	t.Run("E-05 acepta 0 y 0.5", func(t *testing.T) {
		for _, cuerpo := range []string{`{"puntos": 0}`, `{"puntos": 0.5}`} {
			router, servicio := nuevoRouterDePoker(t, 2)
			servicio.On("Votar", mock.Anything, uint(1), uint(9), uint(2), mock.AnythingOfType("float64")).
				Return(resultadoVotoDeEjemplo(), nil)

			w := ejecutar(t, router, http.MethodPost, ruta, cuerpo)

			assert.Equal(t, http.StatusOK, w.Code, cuerpo)
		}
	})

	t.Run("E-09 puntos no numéricos o ausentes responden 422 sin llamar al servicio", func(t *testing.T) {
		for _, cuerpo := range []string{`{"puntos": "cinco"}`, `{}`, `{"puntos": null}`, `{"puntos": true}`} {
			router, _ := nuevoRouterDePoker(t, 2)

			w := ejecutar(t, router, http.MethodPost, ruta, cuerpo)

			require.Equal(t, http.StatusUnprocessableEntity, w.Code, cuerpo)
			respuesta := decodificarError(t, w)
			assert.Equal(t, codigoPuntosPokerInvalidos, respuesta.Error)
			assert.Equal(t, "puntos", respuesta.Campo)
		}
	})

	t.Run("E-07 puntos fuera de escala rechazados por el servicio", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("Votar", mock.Anything, uint(1), uint(9), uint(2), 4.0).Return(nil, domain.ErrPuntosPokerInvalidos)

		w := ejecutar(t, router, http.MethodPost, ruta, `{"puntos": 4}`)

		require.Equal(t, http.StatusUnprocessableEntity, w.Code)
		respuesta := decodificarError(t, w)
		assert.Equal(t, codigoPuntosPokerInvalidos, respuesta.Error)
		assert.Equal(t, "puntos", respuesta.Campo)
	})

	t.Run("E-10 cuerpo malformado responde 400", func(t *testing.T) {
		router, _ := nuevoRouterDePoker(t, 2)

		w := ejecutar(t, router, http.MethodPost, ruta, `{"puntos": `)

		require.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, codigoJSONInvalido, decodificarError(t, w).Error)
	})

	t.Run("E-06 sesión que no está Votando responde 409", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("Votar", mock.Anything, uint(1), uint(9), uint(2), 5.0).Return(nil, domain.ErrSesionPokerNoVotando)

		w := ejecutar(t, router, http.MethodPost, ruta, `{"puntos": 5}`)

		require.Equal(t, http.StatusConflict, w.Code)
		assert.Equal(t, codigoSesionPokerNoVotando, decodificarError(t, w).Error)
	})

	t.Run("E-12 sesión inexistente responde 404", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("Votar", mock.Anything, uint(1), uint(9), uint(2), 5.0).Return(nil, domain.ErrSesionPokerNoEncontrada)

		w := ejecutar(t, router, http.MethodPost, ruta, `{"puntos": 5}`)

		require.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, codigoSesionPokerNoEncontrada, decodificarError(t, w).Error)
	})

	t.Run("sesionId inválido responde 400", func(t *testing.T) {
		router, _ := nuevoRouterDePoker(t, 2)

		w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/poker-sessions/abc/votos", `{"puntos": 5}`)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("E-13 sin identidad en el contexto responde 401", func(t *testing.T) {
		router, _ := nuevoRouterDePoker(t, 0)

		w := ejecutar(t, router, http.MethodPost, ruta, `{"puntos": 5}`)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func detallePokerCerrado() *domain.DetalleSesionPoker {
	detalle := detallePokerDeEjemplo(domain.EstadoSesionCerrada)
	acordados := 5.0
	cerradoEn := time.Date(2026, 11, 13, 18, 0, 0, 0, time.UTC)
	detalle.Sesion.PuntosAcordados = &acordados
	detalle.Sesion.CerradoEn = &cerradoEn
	detalle.Sesion.StoryEstimate = &domain.StoryEstimate{
		ID: 12, HistoriaID: 5, SesionID: 9, RondaID: 3, Puntos: 5,
		Fuente: "planning_poker", EstimadoPor: punteroUintPoker(2), EstimadoEn: cerradoEn,
	}
	return detalle
}

func TestAcordarPokerHandler(t *testing.T) {
	const ruta = "/api/proyectos/1/poker-sessions/9/acuerdo"

	t.Run("E-01 y E-02 acuerda y devuelve la estimación", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("Acordar", mock.Anything, uint(1), uint(9), uint(2), 5.0).Return(detallePokerCerrado(), nil)

		w := ejecutar(t, router, http.MethodPost, ruta, `{"puntos": 5}`)

		require.Equal(t, http.StatusOK, w.Code)
		respuesta := decodificarDetallePoker(t, w)
		assert.Equal(t, domain.EstadoSesionCerrada, respuesta.Estado)
		require.NotNil(t, respuesta.PuntosAcordados)
		assert.Equal(t, 5.0, *respuesta.PuntosAcordados)
		assert.NotNil(t, respuesta.CerradoEn)
		require.NotNil(t, respuesta.StoryEstimate)
		assert.Equal(t, 5.0, respuesta.StoryEstimate.Puntos)
		assert.Equal(t, "planning_poker", respuesta.StoryEstimate.Fuente)
	})

	t.Run("E-03 acepta 0 y 0.5", func(t *testing.T) {
		for _, cuerpo := range []string{`{"puntos": 0}`, `{"puntos": 0.5}`} {
			router, servicio := nuevoRouterDePoker(t, 2)
			servicio.On("Acordar", mock.Anything, uint(1), uint(9), uint(2), mock.AnythingOfType("float64")).
				Return(detallePokerCerrado(), nil)

			w := ejecutar(t, router, http.MethodPost, ruta, cuerpo)

			assert.Equal(t, http.StatusOK, w.Code, cuerpo)
		}
	})

	t.Run("E-04 acordar desde Votando responde 409 SESION_POKER_NO_REVELADA", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("Acordar", mock.Anything, uint(1), uint(9), uint(2), 5.0).
			Return(nil, domain.ErrSesionPokerNoRevelada)

		w := ejecutar(t, router, http.MethodPost, ruta, `{"puntos": 5}`)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Equal(t, codigoSesionPokerNoRevelada, decodificarError(t, w).Error)
	})

	t.Run("E-05 y E-06 acordar cerrada o cancelada responde 409 SESION_POKER_CERRADA", func(t *testing.T) {
		for _, estado := range []domain.EstadoSesionPoker{domain.EstadoSesionCerrada, domain.EstadoSesionCancelada} {
			t.Run(string(estado), func(t *testing.T) {
				router, servicio := nuevoRouterDePoker(t, 2)
				servicio.On("Acordar", mock.Anything, uint(1), uint(9), uint(2), 5.0).
					Return(nil, domain.ErrSesionPokerCerrada)

				w := ejecutar(t, router, http.MethodPost, ruta, `{"puntos": 5}`)

				assert.Equal(t, http.StatusConflict, w.Code)
				assert.Equal(t, codigoSesionPokerCerrada, decodificarError(t, w).Error)
			})
		}
	})

	t.Run("E-07 puntos ausentes, no numéricos o fuera de escala", func(t *testing.T) {
		for _, cuerpo := range []string{`{"puntos": "cinco"}`, `{}`, `{"puntos": null}`, `{"puntos": true}`} {
			router, _ := nuevoRouterDePoker(t, 2)

			w := ejecutar(t, router, http.MethodPost, ruta, cuerpo)

			require.Equal(t, http.StatusUnprocessableEntity, w.Code, cuerpo)
			respuesta := decodificarError(t, w)
			assert.Equal(t, codigoPuntosPokerInvalidos, respuesta.Error)
			assert.Equal(t, "puntos", respuesta.Campo)
		}
	})

	t.Run("E-07 puntos fuera de escala rechazados por el servicio", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("Acordar", mock.Anything, uint(1), uint(9), uint(2), 4.0).
			Return(nil, domain.ErrPuntosPokerInvalidos)

		w := ejecutar(t, router, http.MethodPost, ruta, `{"puntos": 4}`)

		require.Equal(t, http.StatusUnprocessableEntity, w.Code)
		respuesta := decodificarError(t, w)
		assert.Equal(t, codigoPuntosPokerInvalidos, respuesta.Error)
		assert.Equal(t, "puntos", respuesta.Campo)
	})

	t.Run("E-08 JSON malformado responde 400", func(t *testing.T) {
		router, _ := nuevoRouterDePoker(t, 2)

		w := ejecutar(t, router, http.MethodPost, ruta, `{"puntos": `)

		require.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, codigoJSONInvalido, decodificarError(t, w).Error)
	})

	t.Run("sesión inexistente responde 404", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("Acordar", mock.Anything, uint(1), uint(9), uint(2), 5.0).
			Return(nil, domain.ErrSesionPokerNoEncontrada)

		w := ejecutar(t, router, http.MethodPost, ruta, `{"puntos": 5}`)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, codigoSesionPokerNoEncontrada, decodificarError(t, w).Error)
	})

	t.Run("sesionId inválido responde 400", func(t *testing.T) {
		router, _ := nuevoRouterDePoker(t, 2)

		w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/poker-sessions/abc/acuerdo", `{"puntos": 5}`)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, codigoParametroInvalido, decodificarError(t, w).Error)
	})

	t.Run("sin identidad responde 401", func(t *testing.T) {
		router, _ := nuevoRouterDePoker(t, 0)

		w := ejecutar(t, router, http.MethodPost, ruta, `{"puntos": 5}`)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestCancelarPokerHandler(t *testing.T) {
	const ruta = "/api/proyectos/1/poker-sessions/9/cancelar"

	canceladoEn := time.Date(2026, 11, 13, 18, 0, 0, 0, time.UTC)
	detalleCancelado := func() *domain.DetalleSesionPoker {
		detalle := detallePokerDeEjemplo(domain.EstadoSesionCancelada)
		detalle.Sesion.CanceladoEn = &canceladoEn
		return detalle
	}

	t.Run("E-10 y E-11 cancela sin estimar", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("Cancelar", mock.Anything, uint(1), uint(9), uint(2)).Return(detalleCancelado(), nil)

		w := ejecutar(t, router, http.MethodPost, ruta, nil)

		require.Equal(t, http.StatusOK, w.Code)
		respuesta := decodificarDetallePoker(t, w)
		assert.Equal(t, domain.EstadoSesionCancelada, respuesta.Estado)
		assert.NotNil(t, respuesta.CanceladoEn)
		assert.Nil(t, respuesta.PuntosAcordados)
		assert.Nil(t, respuesta.CerradoEn)
		assert.Nil(t, respuesta.StoryEstimate)
	})

	t.Run("E-13 y E-14 cancelar cerrada o cancelada responde 409 SESION_POKER_CERRADA", func(t *testing.T) {
		for _, estado := range []domain.EstadoSesionPoker{domain.EstadoSesionCerrada, domain.EstadoSesionCancelada} {
			t.Run(string(estado), func(t *testing.T) {
				router, servicio := nuevoRouterDePoker(t, 2)
				servicio.On("Cancelar", mock.Anything, uint(1), uint(9), uint(2)).
					Return(nil, domain.ErrSesionPokerCerrada)

				w := ejecutar(t, router, http.MethodPost, ruta, nil)

				assert.Equal(t, http.StatusConflict, w.Code)
				assert.Equal(t, codigoSesionPokerCerrada, decodificarError(t, w).Error)
			})
		}
	})

	t.Run("sesión inexistente responde 404", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		servicio.On("Cancelar", mock.Anything, uint(1), uint(9), uint(2)).
			Return(nil, domain.ErrSesionPokerNoEncontrada)

		w := ejecutar(t, router, http.MethodPost, ruta, nil)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, codigoSesionPokerNoEncontrada, decodificarError(t, w).Error)
	})

	t.Run("sesionId inválido responde 400", func(t *testing.T) {
		router, _ := nuevoRouterDePoker(t, 2)

		w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/poker-sessions/abc/cancelar", nil)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, codigoParametroInvalido, decodificarError(t, w).Error)
	})

	t.Run("sin identidad responde 401", func(t *testing.T) {
		router, _ := nuevoRouterDePoker(t, 0)

		w := ejecutar(t, router, http.MethodPost, ruta, nil)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func punteroUintPoker(valor uint) *uint {
	return &valor
}
