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
	t.Run("E-12 incluye cerradas y canceladas", func(t *testing.T) {
		router, servicio := nuevoRouterDePoker(t, 2)
		creado := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
		servicio.On("ListarPorHistoria", mock.Anything, uint(1), uint(5)).Return([]domain.SesionPoker{
			{ID: 3, Estado: domain.EstadoSesionCerrada, RondaActual: 2, CreadoPor: 2, CreadoEn: creado},
			{ID: 2, Estado: domain.EstadoSesionCancelada, RondaActual: 1, CreadoPor: 2, CreadoEn: creado},
		}, nil)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/1/historias/5/poker-sessions", nil)

		require.Equal(t, http.StatusOK, w.Code)
		var respuestas []respuestaHistorialSesion
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuestas))
		require.Len(t, respuestas, 2)
		assert.Equal(t, domain.EstadoSesionCerrada, respuestas[0].Estado)
		assert.Equal(t, domain.EstadoSesionCancelada, respuestas[1].Estado)
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
