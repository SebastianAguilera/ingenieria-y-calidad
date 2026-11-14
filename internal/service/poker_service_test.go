package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

const (
	proyectoPokerID = uint(1)
	historiaPokerID = uint(5)
	anaPokerID      = uint(2)
	benPokerID      = uint(7)
)

func historiaDePoker(estado domain.EstadoHistoria) *domain.Historia {
	return &domain.Historia{
		ID:         historiaPokerID,
		ProyectoID: proyectoPokerID,
		Numero:     5,
		Codigo:     "US-05",
		Titulo:     "Estimar en equipo",
		Prioridad:  domain.PrioridadMedia,
		Estado:     estado,
	}
}

func nuevoServicioPokerDePrueba(t *testing.T) (*PokerService, *MockProyectoRepository, *MockHistoriaRepository, *MockUsuarioRepository, *MockSesionPokerRepository) {
	t.Helper()
	proyectos := nuevoMockProyectoRepository(t)
	historias := nuevoMockHistoriaRepository(t)
	usuarios := nuevoMockUsuarioRepository(t)
	poker := nuevoMockSesionPokerRepository(t)
	return NewPokerService(proyectos, historias, usuarios, poker), proyectos, historias, usuarios, poker
}

func TestCrearSesionPoker(t *testing.T) {
	inicio := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)

	t.Run("crea la sesión en Votando con ronda 1", func(t *testing.T) {
		servicio, proyectos, historias, usuarios, poker := nuevoServicioPokerDePrueba(t)
		ctx := context.Background()

		proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).
			Return(proyectoDePrueba(proyectoPokerID, "Metrics", inicio, nil, domain.EstadoActivo), nil)
		historias.On("ObtenerPorID", mock.Anything, proyectoPokerID, historiaPokerID).
			Return(historiaDePoker(domain.EstadoHistoriaBacklog), nil)
		poker.On("ObtenerActivaPorHistoria", mock.Anything, historiaPokerID).Return(nil, nil)
		poker.On("CrearSesion", mock.Anything, mock.MatchedBy(func(s *domain.SesionPoker) bool {
			return s.ProyectoID == proyectoPokerID && s.HistoriaID == historiaPokerID && s.CreadoPor == anaPokerID
		})).Return(nil).Run(func(args mock.Arguments) {
			sesion := args.Get(1).(*domain.SesionPoker)
			sesion.ID = 11
			sesion.Estado = domain.EstadoSesionVotando
			sesion.RondaActual = 1
		})
		usuarios.On("ObtenerPorID", mock.Anything, anaPokerID).
			Return(&domain.Usuario{ID: anaPokerID, Nombre: "Ana"}, nil)
		poker.On("ContarMiembros", mock.Anything, proyectoPokerID).Return(3, nil)

		detalle, err := servicio.CrearSesion(ctx, proyectoPokerID, historiaPokerID, anaPokerID)

		require.NoError(t, err)
		assert.Equal(t, uint(11), detalle.Sesion.ID)
		assert.Equal(t, domain.EstadoSesionVotando, detalle.Sesion.Estado)
		assert.Equal(t, 1, detalle.Sesion.RondaActual)
		assert.Equal(t, "Ana", detalle.Creador.Nombre)
		assert.Equal(t, 3, detalle.TotalMiembros)
		assert.False(t, detalle.YaVote)
		assert.Empty(t, detalle.Votos)
	})

	t.Run("rechaza una historia en estado Hecho", func(t *testing.T) {
		servicio, proyectos, historias, _, poker := nuevoServicioPokerDePrueba(t)
		ctx := context.Background()

		proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).
			Return(proyectoDePrueba(proyectoPokerID, "Metrics", inicio, nil, domain.EstadoActivo), nil)
		historias.On("ObtenerPorID", mock.Anything, proyectoPokerID, historiaPokerID).
			Return(historiaDePoker(domain.EstadoHistoriaHecho), nil)

		_, err := servicio.CrearSesion(ctx, proyectoPokerID, historiaPokerID, anaPokerID)

		assert.ErrorIs(t, err, domain.ErrHistoriaNoEstimable)
		poker.AssertNotCalled(t, "CrearSesion", mock.Anything, mock.Anything)
	})

	t.Run("rechaza cuando ya hay una sesión activa", func(t *testing.T) {
		servicio, proyectos, historias, _, poker := nuevoServicioPokerDePrueba(t)
		ctx := context.Background()

		proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).
			Return(proyectoDePrueba(proyectoPokerID, "Metrics", inicio, nil, domain.EstadoActivo), nil)
		historias.On("ObtenerPorID", mock.Anything, proyectoPokerID, historiaPokerID).
			Return(historiaDePoker(domain.EstadoHistoriaBacklog), nil)
		poker.On("ObtenerActivaPorHistoria", mock.Anything, historiaPokerID).
			Return(&domain.SesionPoker{ID: 9, Estado: domain.EstadoSesionVotando}, nil)

		_, err := servicio.CrearSesion(ctx, proyectoPokerID, historiaPokerID, anaPokerID)

		assert.ErrorIs(t, err, domain.ErrSesionPokerActiva)
		poker.AssertNotCalled(t, "CrearSesion", mock.Anything, mock.Anything)
	})

	t.Run("propaga la carrera del índice único parcial", func(t *testing.T) {
		servicio, proyectos, historias, usuarios, poker := nuevoServicioPokerDePrueba(t)
		ctx := context.Background()

		proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).
			Return(proyectoDePrueba(proyectoPokerID, "Metrics", inicio, nil, domain.EstadoActivo), nil)
		historias.On("ObtenerPorID", mock.Anything, proyectoPokerID, historiaPokerID).
			Return(historiaDePoker(domain.EstadoHistoriaBacklog), nil)
		poker.On("ObtenerActivaPorHistoria", mock.Anything, historiaPokerID).Return(nil, nil)
		usuarios.On("ObtenerPorID", mock.Anything, anaPokerID).
			Return(&domain.Usuario{ID: anaPokerID, Nombre: "Ana"}, nil)
		poker.On("CrearSesion", mock.Anything, mock.Anything).Return(domain.ErrSesionPokerActiva)

		_, err := servicio.CrearSesion(ctx, proyectoPokerID, historiaPokerID, anaPokerID)

		assert.ErrorIs(t, err, domain.ErrSesionPokerActiva)
	})

	t.Run("propaga historia inexistente", func(t *testing.T) {
		servicio, proyectos, historias, _, _ := nuevoServicioPokerDePrueba(t)
		ctx := context.Background()

		proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).
			Return(proyectoDePrueba(proyectoPokerID, "Metrics", inicio, nil, domain.EstadoActivo), nil)
		historias.On("ObtenerPorID", mock.Anything, proyectoPokerID, historiaPokerID).
			Return(nil, domain.ErrHistoriaNoEncontrada)

		_, err := servicio.CrearSesion(ctx, proyectoPokerID, historiaPokerID, anaPokerID)

		assert.ErrorIs(t, err, domain.ErrHistoriaNoEncontrada)
	})
}

func TestListarSesionesDeHistoria(t *testing.T) {
	inicio := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)

	t.Run("devuelve el historial completo", func(t *testing.T) {
		servicio, proyectos, historias, _, poker := nuevoServicioPokerDePrueba(t)
		ctx := context.Background()

		proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).
			Return(proyectoDePrueba(proyectoPokerID, "Metrics", inicio, nil, domain.EstadoActivo), nil)
		historias.On("ObtenerPorID", mock.Anything, proyectoPokerID, historiaPokerID).
			Return(historiaDePoker(domain.EstadoHistoriaBacklog), nil)
		poker.On("ListarPorHistoria", mock.Anything, historiaPokerID).Return([]domain.SesionPoker{
			{ID: 2, Estado: domain.EstadoSesionVotando},
			{ID: 1, Estado: domain.EstadoSesionCerrada},
		}, nil)
		poker.On("ObtenerStoryEstimatePorSesion", mock.Anything, uint(1)).Return(&domain.StoryEstimate{
			ID: 12, HistoriaID: historiaPokerID, SesionID: 1, RondaID: 3, Puntos: 5, Fuente: "planning_poker",
		}, nil)

		sesiones, err := servicio.ListarPorHistoria(ctx, proyectoPokerID, historiaPokerID)

		require.NoError(t, err)
		require.Len(t, sesiones, 2)
		assert.Equal(t, domain.EstadoSesionCerrada, sesiones[1].Estado)
		require.NotNil(t, sesiones[1].StoryEstimate)
		assert.Equal(t, 5.0, sesiones[1].StoryEstimate.Puntos)
		assert.Equal(t, uint(3), sesiones[1].StoryEstimate.RondaID)
	})

	t.Run("devuelve lista vacía cuando no hay sesiones", func(t *testing.T) {
		servicio, proyectos, historias, _, poker := nuevoServicioPokerDePrueba(t)
		ctx := context.Background()

		proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).
			Return(proyectoDePrueba(proyectoPokerID, "Metrics", inicio, nil, domain.EstadoActivo), nil)
		historias.On("ObtenerPorID", mock.Anything, proyectoPokerID, historiaPokerID).
			Return(historiaDePoker(domain.EstadoHistoriaBacklog), nil)
		poker.On("ListarPorHistoria", mock.Anything, historiaPokerID).Return(nil, nil)

		sesiones, err := servicio.ListarPorHistoria(ctx, proyectoPokerID, historiaPokerID)

		require.NoError(t, err)
		assert.NotNil(t, sesiones)
		assert.Empty(t, sesiones)
	})
}

func TestDetalleSesionPoker(t *testing.T) {
	inicio := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)

	t.Run("en Votando no expone votos ajenos", func(t *testing.T) {
		servicio, proyectos, _, _, poker := nuevoServicioPokerDePrueba(t)
		ctx := context.Background()

		proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).
			Return(proyectoDePrueba(proyectoPokerID, "Metrics", inicio, nil, domain.EstadoActivo), nil)
		poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).Return(&domain.SesionPoker{
			ID: 9, ProyectoID: proyectoPokerID, HistoriaID: historiaPokerID, CreadoPor: anaPokerID,
			CreadorNombre: "Ana", RondaActual: 1, Estado: domain.EstadoSesionVotando,
		}, nil)
		poker.On("ObtenerRonda", mock.Anything, uint(9), 1).Return(&domain.RondaPoker{ID: 3, SesionID: 9, NumeroRonda: 1}, nil)
		poker.On("ContarMiembros", mock.Anything, proyectoPokerID).Return(2, nil)
		poker.On("ContarVotosPorRonda", mock.Anything, uint(3)).Return(1, nil)
		poker.On("ObtenerVotoUsuario", mock.Anything, uint(3), anaPokerID).Return(&domain.VotoPoker{
			RondaID: 3, UsuarioID: anaPokerID, Nombre: "Ana", Puntos: 3,
		}, nil)
		poker.On("ListarRondasPorSesion", mock.Anything, uint(9)).Return([]domain.RondaPoker{
			{ID: 3, SesionID: 9, NumeroRonda: 1},
		}, nil)

		detalle, err := servicio.Detalle(ctx, proyectoPokerID, 9, anaPokerID)

		require.NoError(t, err)
		assert.Equal(t, 2, detalle.TotalMiembros)
		assert.Equal(t, 1, detalle.VotosRegistrados)
		assert.True(t, detalle.YaVote)
		require.NotNil(t, detalle.MiVoto)
		assert.Equal(t, 3.0, *detalle.MiVoto)
		assert.Empty(t, detalle.Votos, "en Votando no se exponen votos ajenos")
		assert.Nil(t, detalle.Dispersion, "en Votando no se exponen estadísticas")
		require.Len(t, detalle.HistorialRondas, 1)
		assert.Empty(t, detalle.HistorialRondas[0].Votos)
		assert.Nil(t, detalle.HistorialRondas[0].Dispersion)
		poker.AssertNotCalled(t, "ListarVotosPorRonda", mock.Anything, mock.Anything)
	})

	t.Run("en Revelada expone votos con nombre y puntos", func(t *testing.T) {
		servicio, proyectos, _, _, poker := nuevoServicioPokerDePrueba(t)
		ctx := context.Background()

		proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).
			Return(proyectoDePrueba(proyectoPokerID, "Metrics", inicio, nil, domain.EstadoActivo), nil)
		poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).Return(&domain.SesionPoker{
			ID: 9, ProyectoID: proyectoPokerID, HistoriaID: historiaPokerID, CreadoPor: anaPokerID,
			CreadorNombre: "Ana", RondaActual: 1, Estado: domain.EstadoSesionRevelada,
		}, nil)
		poker.On("ObtenerRonda", mock.Anything, uint(9), 1).Return(&domain.RondaPoker{ID: 3, SesionID: 9, NumeroRonda: 1}, nil)
		poker.On("ContarMiembros", mock.Anything, proyectoPokerID).Return(2, nil)
		poker.On("ContarVotosPorRonda", mock.Anything, uint(3)).Return(2, nil)
		poker.On("ObtenerVotoUsuario", mock.Anything, uint(3), anaPokerID).Return(&domain.VotoPoker{
			RondaID: 3, UsuarioID: anaPokerID, Nombre: "Ana", Puntos: 3,
		}, nil)
		poker.On("ListarVotosPorRonda", mock.Anything, uint(3)).Return([]domain.VotoPoker{
			{UsuarioID: anaPokerID, Nombre: "Ana", Puntos: 3},
			{UsuarioID: benPokerID, Nombre: "Ben", Puntos: 5},
		}, nil)
		reveladaEn := time.Date(2026, 11, 13, 18, 0, 0, 0, time.UTC)
		poker.On("ListarRondasPorSesion", mock.Anything, uint(9)).Return([]domain.RondaPoker{
			{ID: 3, SesionID: 9, NumeroRonda: 1, ReveladaEn: &reveladaEn},
		}, nil)

		detalle, err := servicio.Detalle(ctx, proyectoPokerID, 9, anaPokerID)

		require.NoError(t, err)
		require.Len(t, detalle.Votos, 2)
		assert.Equal(t, "Ben", detalle.Votos[1].Nombre)
		assert.Equal(t, 5.0, detalle.Votos[1].Puntos)
		assert.Equal(t, 1, detalle.Votos[1].Ronda)
		require.NotNil(t, detalle.Dispersion)
		assert.Equal(t, 4.0, detalle.Dispersion.Promedio)
		assert.True(t, detalle.Dispersion.HayDispersion)
		require.Len(t, detalle.HistorialRondas, 1)
		assert.Equal(t, domain.EstadoSesionRevelada, detalle.HistorialRondas[0].Estado)
		require.NotNil(t, detalle.HistorialRondas[0].Dispersion)
	})

	t.Run("propaga sesión inexistente", func(t *testing.T) {
		servicio, proyectos, _, _, poker := nuevoServicioPokerDePrueba(t)
		ctx := context.Background()

		proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).
			Return(proyectoDePrueba(proyectoPokerID, "Metrics", inicio, nil, domain.EstadoActivo), nil)
		poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).
			Return(nil, domain.ErrSesionPokerNoEncontrada)

		_, err := servicio.Detalle(ctx, proyectoPokerID, 9, anaPokerID)

		assert.ErrorIs(t, err, domain.ErrSesionPokerNoEncontrada)
	})
}

func TestSesionesActivasPoker(t *testing.T) {
	inicio := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)

	t.Run("devuelve conteos y resumen de historia sin votos", func(t *testing.T) {
		servicio, proyectos, historias, _, poker := nuevoServicioPokerDePrueba(t)
		ctx := context.Background()

		proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).
			Return(proyectoDePrueba(proyectoPokerID, "Metrics", inicio, nil, domain.EstadoActivo), nil)
		poker.On("ListarActivasPorProyecto", mock.Anything, proyectoPokerID).Return([]domain.SesionPoker{
			{ID: 9, ProyectoID: proyectoPokerID, HistoriaID: historiaPokerID, CreadoPor: anaPokerID,
				CreadorNombre: "Ana", RondaActual: 1, Estado: domain.EstadoSesionRevelada},
		}, nil)
		poker.On("ContarMiembros", mock.Anything, proyectoPokerID).Return(3, nil)
		historias.On("ObtenerPorID", mock.Anything, proyectoPokerID, historiaPokerID).
			Return(historiaDePoker(domain.EstadoHistoriaBacklog), nil)
		poker.On("ObtenerRonda", mock.Anything, uint(9), 1).Return(&domain.RondaPoker{ID: 3, SesionID: 9, NumeroRonda: 1}, nil)
		poker.On("ContarVotosPorRonda", mock.Anything, uint(3)).Return(2, nil)
		poker.On("ObtenerVotoUsuario", mock.Anything, uint(3), benPokerID).Return(nil, nil)

		activas, err := servicio.Activas(ctx, proyectoPokerID, benPokerID)

		require.NoError(t, err)
		require.Len(t, activas, 1)
		assert.Equal(t, "US-05", activas[0].Historia.Codigo)
		assert.Equal(t, 3, activas[0].TotalMiembros)
		assert.Equal(t, 2, activas[0].VotosRegistrados)
		assert.False(t, activas[0].YaVote)
	})

	t.Run("devuelve lista vacía cuando no hay activas", func(t *testing.T) {
		servicio, proyectos, _, _, poker := nuevoServicioPokerDePrueba(t)
		ctx := context.Background()

		proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).
			Return(proyectoDePrueba(proyectoPokerID, "Metrics", inicio, nil, domain.EstadoActivo), nil)
		poker.On("ListarActivasPorProyecto", mock.Anything, proyectoPokerID).Return(nil, nil)
		poker.On("ContarMiembros", mock.Anything, proyectoPokerID).Return(3, nil)

		activas, err := servicio.Activas(ctx, proyectoPokerID, anaPokerID)

		require.NoError(t, err)
		assert.NotNil(t, activas)
		assert.Empty(t, activas)
	})
}

func TestVotarPoker(t *testing.T) {
	inicio := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)
	votadoEn := time.Date(2026, 11, 13, 18, 0, 0, 0, time.UTC)

	preparar := func(t *testing.T) (*PokerService, *MockSesionPokerRepository) {
		servicio, proyectos, _, _, poker := nuevoServicioPokerDePrueba(t)
		proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).
			Return(proyectoDePrueba(proyectoPokerID, "Metrics", inicio, nil, domain.EstadoActivo), nil)
		return servicio, poker
	}
	sesionVotando := &domain.SesionPoker{
		ID: 9, ProyectoID: proyectoPokerID, HistoriaID: historiaPokerID, CreadoPor: anaPokerID,
		RondaActual: 1, Estado: domain.EstadoSesionVotando,
	}

	t.Run("E-01 registra un voto válido y devuelve solo conteos", func(t *testing.T) {
		servicio, poker := preparar(t)
		poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).Return(sesionVotando, nil)
		poker.On("RegistrarVoto", mock.Anything, uint(9), anaPokerID, 5.0).
			Return(&domain.VotoPoker{RondaID: 3, UsuarioID: anaPokerID, Puntos: 5, VotadoEn: votadoEn}, nil)
		poker.On("ContarVotosPorRonda", mock.Anything, uint(3)).Return(1, nil)
		poker.On("ContarMiembros", mock.Anything, proyectoPokerID).Return(3, nil)

		resultado, err := servicio.Votar(context.Background(), proyectoPokerID, 9, anaPokerID, 5)

		require.NoError(t, err)
		assert.Equal(t, uint(9), resultado.SesionID)
		assert.Equal(t, 5.0, resultado.Puntos)
		assert.Equal(t, votadoEn, resultado.VotadoEn)
		assert.Equal(t, 1, resultado.VotosRegistrados)
		assert.Equal(t, 3, resultado.TotalMiembros)
	})

	t.Run("E-05 acepta toda la escala de US-05", func(t *testing.T) {
		for _, puntos := range []float64{0, 0.5, 1, 2, 3, 5, 8, 13, 21} {
			t.Run(fmt.Sprintf("%v", puntos), func(t *testing.T) {
				servicio, poker := preparar(t)
				poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).Return(sesionVotando, nil)
				poker.On("RegistrarVoto", mock.Anything, uint(9), anaPokerID, puntos).
					Return(&domain.VotoPoker{RondaID: 3, UsuarioID: anaPokerID, Puntos: puntos}, nil)
				poker.On("ContarVotosPorRonda", mock.Anything, uint(3)).Return(1, nil)
				poker.On("ContarMiembros", mock.Anything, proyectoPokerID).Return(2, nil)

				resultado, err := servicio.Votar(context.Background(), proyectoPokerID, 9, anaPokerID, puntos)

				require.NoError(t, err)
				assert.Equal(t, puntos, resultado.Puntos)
			})
		}
	})

	t.Run("E-07 y E-08 rechaza puntos fuera de escala sin tocar la base", func(t *testing.T) {
		for _, puntos := range []float64{4, -1, 22, 0.25, 100} {
			t.Run(fmt.Sprintf("%v", puntos), func(t *testing.T) {
				servicio, poker := preparar(t)
				poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).Return(sesionVotando, nil)

				_, err := servicio.Votar(context.Background(), proyectoPokerID, 9, anaPokerID, puntos)

				assert.ErrorIs(t, err, domain.ErrPuntosPokerInvalidos)
				poker.AssertNotCalled(t, "RegistrarVoto", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
		}
	})

	t.Run("E-06 propaga que la sesión ya no está Votando", func(t *testing.T) {
		servicio, poker := preparar(t)
		poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).Return(sesionVotando, nil)
		poker.On("RegistrarVoto", mock.Anything, uint(9), anaPokerID, 5.0).
			Return(nil, domain.ErrSesionPokerNoVotando)

		_, err := servicio.Votar(context.Background(), proyectoPokerID, 9, anaPokerID, 5)

		assert.ErrorIs(t, err, domain.ErrSesionPokerNoVotando)
		poker.AssertNotCalled(t, "ContarVotosPorRonda", mock.Anything, mock.Anything)
	})

	t.Run("E-12 sesión inexistente o de otro proyecto", func(t *testing.T) {
		servicio, poker := preparar(t)
		poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).
			Return(nil, domain.ErrSesionPokerNoEncontrada)

		_, err := servicio.Votar(context.Background(), proyectoPokerID, 9, anaPokerID, 5)

		assert.ErrorIs(t, err, domain.ErrSesionPokerNoEncontrada)
		poker.AssertNotCalled(t, "RegistrarVoto", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("proyecto inexistente", func(t *testing.T) {
		servicio, proyectos, _, _, poker := nuevoServicioPokerDePrueba(t)
		proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).Return(nil, domain.ErrProyectoNoEncontrado)

		_, err := servicio.Votar(context.Background(), proyectoPokerID, 9, anaPokerID, 5)

		assert.ErrorIs(t, err, domain.ErrProyectoNoEncontrado)
		poker.AssertNotCalled(t, "RegistrarVoto", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestRevelarSesionPoker(t *testing.T) {
	inicio := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)
	reveladaEn := time.Date(2026, 11, 13, 18, 0, 0, 0, time.UTC)

	preparar := func(t *testing.T) (*PokerService, *MockSesionPokerRepository) {
		servicio, proyectos, _, _, poker := nuevoServicioPokerDePrueba(t)
		proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).
			Return(proyectoDePrueba(proyectoPokerID, "Metrics", inicio, nil, domain.EstadoActivo), nil)
		return servicio, poker
	}
	sesionVotando := func() *domain.SesionPoker {
		return &domain.SesionPoker{
			ID: 9, ProyectoID: proyectoPokerID, HistoriaID: historiaPokerID, CreadoPor: anaPokerID,
			CreadorNombre: "Ana", RondaActual: 1, Estado: domain.EstadoSesionVotando,
		}
	}

	t.Run("E-01 revela una ronda con votos y calcula la dispersión", func(t *testing.T) {
		servicio, poker := preparar(t)
		poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).Return(sesionVotando(), nil)
		poker.On("RevelarSesion", mock.Anything, uint(9)).Return(nil)
		poker.On("ObtenerRonda", mock.Anything, uint(9), 1).Return(&domain.RondaPoker{ID: 3, SesionID: 9, NumeroRonda: 1}, nil)
		poker.On("ContarMiembros", mock.Anything, proyectoPokerID).Return(3, nil)
		poker.On("ContarVotosPorRonda", mock.Anything, uint(3)).Return(3, nil)
		poker.On("ObtenerVotoUsuario", mock.Anything, uint(3), anaPokerID).Return(&domain.VotoPoker{
			RondaID: 3, UsuarioID: anaPokerID, Nombre: "Ana", Puntos: 1,
		}, nil)
		poker.On("ListarVotosPorRonda", mock.Anything, uint(3)).Return([]domain.VotoPoker{
			{UsuarioID: anaPokerID, Nombre: "Ana", Puntos: 1},
			{UsuarioID: benPokerID, Nombre: "Ben", Puntos: 3},
			{UsuarioID: 9, Nombre: "Cris", Puntos: 8},
		}, nil)
		poker.On("ListarRondasPorSesion", mock.Anything, uint(9)).Return([]domain.RondaPoker{
			{ID: 3, SesionID: 9, NumeroRonda: 1, ReveladaEn: &reveladaEn},
		}, nil)

		detalle, err := servicio.Revelar(context.Background(), proyectoPokerID, 9, anaPokerID)

		require.NoError(t, err)
		assert.Equal(t, domain.EstadoSesionRevelada, detalle.Sesion.Estado)
		require.Len(t, detalle.Votos, 3)
		assert.Equal(t, 1, detalle.Votos[0].Ronda)
		require.NotNil(t, detalle.Dispersion)
		assert.Equal(t, 4.0, detalle.Dispersion.Promedio)
		assert.Equal(t, 3.0, detalle.Dispersion.Mediana)
		assert.Equal(t, 1.0, detalle.Dispersion.Minimo)
		assert.Equal(t, 8.0, detalle.Dispersion.Maximo)
		assert.Equal(t, 7.0, detalle.Dispersion.Rango)
		assert.True(t, detalle.Dispersion.HayDispersion)
		assert.Equal(t, 3, detalle.Dispersion.TotalVotos)
		require.Len(t, detalle.HistorialRondas, 1)
		require.NotNil(t, detalle.HistorialRondas[0].Dispersion)
	})

	t.Run("E-02 revelar sin votos devuelve SESION_POKER_SIN_VOTOS", func(t *testing.T) {
		servicio, poker := preparar(t)
		poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).Return(sesionVotando(), nil)
		poker.On("RevelarSesion", mock.Anything, uint(9)).Return(domain.ErrSesionPokerSinVotos)

		_, err := servicio.Revelar(context.Background(), proyectoPokerID, 9, anaPokerID)

		assert.ErrorIs(t, err, domain.ErrSesionPokerSinVotos)
		poker.AssertNotCalled(t, "ObtenerRonda", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("E-03 revelar fuera de Votando devuelve SESION_POKER_NO_VOTANDO", func(t *testing.T) {
		servicio, poker := preparar(t)
		poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).Return(sesionVotando(), nil)
		poker.On("RevelarSesion", mock.Anything, uint(9)).Return(domain.ErrSesionPokerNoVotando)

		_, err := servicio.Revelar(context.Background(), proyectoPokerID, 9, anaPokerID)

		assert.ErrorIs(t, err, domain.ErrSesionPokerNoVotando)
	})

	t.Run("sesión inexistente", func(t *testing.T) {
		servicio, poker := preparar(t)
		poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).
			Return(nil, domain.ErrSesionPokerNoEncontrada)

		_, err := servicio.Revelar(context.Background(), proyectoPokerID, 9, anaPokerID)

		assert.ErrorIs(t, err, domain.ErrSesionPokerNoEncontrada)
		poker.AssertNotCalled(t, "RevelarSesion", mock.Anything, mock.Anything)
	})

	t.Run("proyecto inexistente", func(t *testing.T) {
		servicio, proyectos, _, _, poker := nuevoServicioPokerDePrueba(t)
		proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).Return(nil, domain.ErrProyectoNoEncontrado)

		_, err := servicio.Revelar(context.Background(), proyectoPokerID, 9, anaPokerID)

		assert.ErrorIs(t, err, domain.ErrProyectoNoEncontrado)
		poker.AssertNotCalled(t, "RevelarSesion", mock.Anything, mock.Anything)
	})
}

func TestAbrirNuevaRondaPoker(t *testing.T) {
	inicio := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)
	reveladaEn := time.Date(2026, 11, 13, 18, 0, 0, 0, time.UTC)

	preparar := func(t *testing.T) (*PokerService, *MockSesionPokerRepository) {
		servicio, proyectos, _, _, poker := nuevoServicioPokerDePrueba(t)
		proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).
			Return(proyectoDePrueba(proyectoPokerID, "Metrics", inicio, nil, domain.EstadoActivo), nil)
		return servicio, poker
	}
	sesionRevelada := &domain.SesionPoker{
		ID: 9, ProyectoID: proyectoPokerID, HistoriaID: historiaPokerID, CreadoPor: anaPokerID,
		CreadorNombre: "Ana", RondaActual: 1, Estado: domain.EstadoSesionRevelada,
	}

	t.Run("E-08 abre la ronda siguiente y conserva el historial", func(t *testing.T) {
		servicio, poker := preparar(t)
		poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).Return(sesionRevelada, nil)
		poker.On("AbrirNuevaRonda", mock.Anything, uint(9)).Return(&domain.RondaPoker{
			ID: 4, SesionID: 9, NumeroRonda: 2,
		}, nil)
		poker.On("ObtenerRonda", mock.Anything, uint(9), 2).Return(&domain.RondaPoker{ID: 4, SesionID: 9, NumeroRonda: 2}, nil)
		poker.On("ContarMiembros", mock.Anything, proyectoPokerID).Return(3, nil)
		poker.On("ContarVotosPorRonda", mock.Anything, uint(4)).Return(0, nil)
		poker.On("ObtenerVotoUsuario", mock.Anything, uint(4), anaPokerID).Return(nil, nil)
		poker.On("ListarRondasPorSesion", mock.Anything, uint(9)).Return([]domain.RondaPoker{
			{ID: 3, SesionID: 9, NumeroRonda: 1, ReveladaEn: &reveladaEn},
			{ID: 4, SesionID: 9, NumeroRonda: 2},
		}, nil)
		poker.On("ListarVotosPorRonda", mock.Anything, uint(3)).Return([]domain.VotoPoker{
			{UsuarioID: anaPokerID, Nombre: "Ana", Puntos: 3},
			{UsuarioID: benPokerID, Nombre: "Ben", Puntos: 5},
		}, nil)

		detalle, err := servicio.NuevaRonda(context.Background(), proyectoPokerID, 9, anaPokerID)

		require.NoError(t, err)
		assert.Equal(t, 2, detalle.Sesion.RondaActual)
		assert.Equal(t, domain.EstadoSesionVotando, detalle.Sesion.Estado)
		assert.Empty(t, detalle.Votos)
		assert.Nil(t, detalle.Dispersion)
		require.Len(t, detalle.HistorialRondas, 2)
		assert.Equal(t, domain.EstadoSesionRevelada, detalle.HistorialRondas[0].Estado)
		require.NotNil(t, detalle.HistorialRondas[0].Dispersion)
		assert.Equal(t, domain.EstadoSesionVotando, detalle.HistorialRondas[1].Estado)
		assert.Empty(t, detalle.HistorialRondas[1].Votos)
	})

	t.Run("E-09 abrir ronda fuera de Revelada devuelve SESION_POKER_NO_REVELADA", func(t *testing.T) {
		servicio, poker := preparar(t)
		poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).Return(sesionRevelada, nil)
		poker.On("AbrirNuevaRonda", mock.Anything, uint(9)).Return(nil, domain.ErrSesionPokerNoRevelada)

		_, err := servicio.NuevaRonda(context.Background(), proyectoPokerID, 9, anaPokerID)

		assert.ErrorIs(t, err, domain.ErrSesionPokerNoRevelada)
	})

	t.Run("sesión inexistente", func(t *testing.T) {
		servicio, poker := preparar(t)
		poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).
			Return(nil, domain.ErrSesionPokerNoEncontrada)

		_, err := servicio.NuevaRonda(context.Background(), proyectoPokerID, 9, anaPokerID)

		assert.ErrorIs(t, err, domain.ErrSesionPokerNoEncontrada)
		poker.AssertNotCalled(t, "AbrirNuevaRonda", mock.Anything, mock.Anything)
	})

	t.Run("proyecto inexistente", func(t *testing.T) {
		servicio, proyectos, _, _, poker := nuevoServicioPokerDePrueba(t)
		proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).Return(nil, domain.ErrProyectoNoEncontrado)

		_, err := servicio.NuevaRonda(context.Background(), proyectoPokerID, 9, anaPokerID)

		assert.ErrorIs(t, err, domain.ErrProyectoNoEncontrado)
		poker.AssertNotCalled(t, "AbrirNuevaRonda", mock.Anything, mock.Anything)
	})
}

func TestAcordarSesionPoker(t *testing.T) {
	inicio := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)
	estimadoEn := time.Date(2026, 11, 13, 18, 0, 0, 0, time.UTC)

	preparar := func(t *testing.T) (*PokerService, *MockSesionPokerRepository) {
		t.Helper()
		servicio, proyectos, _, _, poker := nuevoServicioPokerDePrueba(t)
		proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).
			Return(proyectoDePrueba(proyectoPokerID, "Metrics", inicio, nil, domain.EstadoActivo), nil)
		return servicio, poker
	}
	sesionRevelada := func() *domain.SesionPoker {
		return &domain.SesionPoker{
			ID: 9, ProyectoID: proyectoPokerID, HistoriaID: historiaPokerID, CreadoPor: anaPokerID,
			CreadorNombre: "Ana", RondaActual: 1, Estado: domain.EstadoSesionRevelada,
		}
	}
	estimacion := func(puntos float64) *domain.StoryEstimate {
		return &domain.StoryEstimate{
			ID: 12, HistoriaID: historiaPokerID, SesionID: 9, RondaID: 3,
			Puntos: puntos, Fuente: "planning_poker", EstimadoPor: punteroIDPoker(anaPokerID), EstimadoEn: estimadoEn,
		}
	}
	configurarDetalleCerrado := func(poker *MockSesionPokerRepository) {
		t.Helper()
		poker.On("ObtenerRonda", mock.Anything, uint(9), 1).
			Return(&domain.RondaPoker{ID: 3, SesionID: 9, NumeroRonda: 1, ReveladaEn: &estimadoEn}, nil)
		poker.On("ContarMiembros", mock.Anything, proyectoPokerID).Return(3, nil)
		poker.On("ContarVotosPorRonda", mock.Anything, uint(3)).Return(3, nil)
		poker.On("ObtenerVotoUsuario", mock.Anything, uint(3), anaPokerID).Return(&domain.VotoPoker{
			RondaID: 3, UsuarioID: anaPokerID, Nombre: "Ana", Puntos: 3,
		}, nil)
		poker.On("ListarVotosPorRonda", mock.Anything, uint(3)).Return([]domain.VotoPoker{
			{UsuarioID: anaPokerID, Nombre: "Ana", Puntos: 3},
			{UsuarioID: benPokerID, Nombre: "Ben", Puntos: 5},
			{UsuarioID: 9, Nombre: "Cris", Puntos: 8},
		}, nil)
		poker.On("ListarRondasPorSesion", mock.Anything, uint(9)).Return([]domain.RondaPoker{
			{ID: 3, SesionID: 9, NumeroRonda: 1, ReveladaEn: &estimadoEn},
		}, nil)
	}

	t.Run("E-01 y E-02 acuerda, cierra y devuelve la estimación completa", func(t *testing.T) {
		servicio, poker := preparar(t)
		poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).Return(sesionRevelada(), nil)
		poker.On("AcordarSesion", mock.Anything, uint(9), anaPokerID, 5.0).Return(estimacion(5), nil)
		configurarDetalleCerrado(poker)

		detalle, err := servicio.Acordar(context.Background(), proyectoPokerID, 9, anaPokerID, 5)

		require.NoError(t, err)
		assert.Equal(t, domain.EstadoSesionCerrada, detalle.Sesion.Estado)
		require.NotNil(t, detalle.Sesion.PuntosAcordados)
		assert.Equal(t, 5.0, *detalle.Sesion.PuntosAcordados)
		require.NotNil(t, detalle.Sesion.CerradoEn)
		assert.Equal(t, estimadoEn, *detalle.Sesion.CerradoEn)
		require.NotNil(t, detalle.Sesion.StoryEstimate)
		assert.Equal(t, uint(9), detalle.Sesion.StoryEstimate.SesionID)
		assert.Equal(t, uint(3), detalle.Sesion.StoryEstimate.RondaID)
		assert.Equal(t, "planning_poker", detalle.Sesion.StoryEstimate.Fuente)
		require.NotNil(t, detalle.Sesion.StoryEstimate.EstimadoPor)
		assert.Equal(t, anaPokerID, *detalle.Sesion.StoryEstimate.EstimadoPor)
	})

	t.Run("E-03 acepta 0 y 0.5", func(t *testing.T) {
		for _, puntos := range []float64{0, 0.5} {
			t.Run(fmt.Sprintf("%v", puntos), func(t *testing.T) {
				servicio, poker := preparar(t)
				poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).Return(sesionRevelada(), nil)
				poker.On("AcordarSesion", mock.Anything, uint(9), anaPokerID, puntos).Return(estimacion(puntos), nil)
				configurarDetalleCerrado(poker)

				detalle, err := servicio.Acordar(context.Background(), proyectoPokerID, 9, anaPokerID, puntos)

				require.NoError(t, err)
				require.NotNil(t, detalle.Sesion.PuntosAcordados)
				assert.Equal(t, puntos, *detalle.Sesion.PuntosAcordados)
			})
		}
	})

	t.Run("E-07 rechaza puntos fuera de escala sin llamar al repositorio", func(t *testing.T) {
		for _, puntos := range []float64{4, -1, 22, 0.25, 100} {
			t.Run(fmt.Sprintf("%v", puntos), func(t *testing.T) {
				servicio, poker := preparar(t)
				poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).Return(sesionRevelada(), nil)

				_, err := servicio.Acordar(context.Background(), proyectoPokerID, 9, anaPokerID, puntos)

				assert.ErrorIs(t, err, domain.ErrPuntosPokerInvalidos)
				poker.AssertNotCalled(t, "AcordarSesion", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
		}
	})

	t.Run("E-04 acordar desde Votando devuelve SESION_POKER_NO_REVELADA", func(t *testing.T) {
		servicio, poker := preparar(t)
		sesion := sesionRevelada()
		sesion.Estado = domain.EstadoSesionVotando
		poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).Return(sesion, nil)
		poker.On("AcordarSesion", mock.Anything, uint(9), anaPokerID, 5.0).
			Return(nil, domain.ErrSesionPokerNoRevelada)

		_, err := servicio.Acordar(context.Background(), proyectoPokerID, 9, anaPokerID, 5)

		assert.ErrorIs(t, err, domain.ErrSesionPokerNoRevelada)
		poker.AssertNotCalled(t, "ObtenerRonda", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("E-05 y E-06 acordar una sesión cerrada o cancelada devuelve SESION_POKER_CERRADA", func(t *testing.T) {
		for _, estado := range []domain.EstadoSesionPoker{domain.EstadoSesionCerrada, domain.EstadoSesionCancelada} {
			t.Run(string(estado), func(t *testing.T) {
				servicio, poker := preparar(t)
				sesion := sesionRevelada()
				sesion.Estado = estado
				poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).Return(sesion, nil)
				poker.On("AcordarSesion", mock.Anything, uint(9), anaPokerID, 5.0).
					Return(nil, domain.ErrSesionPokerCerrada)

				_, err := servicio.Acordar(context.Background(), proyectoPokerID, 9, anaPokerID, 5)

				assert.ErrorIs(t, err, domain.ErrSesionPokerCerrada)
			})
		}
	})

	t.Run("sesión inexistente y proyecto inexistente", func(t *testing.T) {
		t.Run("sesión", func(t *testing.T) {
			servicio, poker := preparar(t)
			poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).
				Return(nil, domain.ErrSesionPokerNoEncontrada)

			_, err := servicio.Acordar(context.Background(), proyectoPokerID, 9, anaPokerID, 5)

			assert.ErrorIs(t, err, domain.ErrSesionPokerNoEncontrada)
			poker.AssertNotCalled(t, "AcordarSesion", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
		t.Run("proyecto", func(t *testing.T) {
			servicio, proyectos, _, _, poker := nuevoServicioPokerDePrueba(t)
			proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).Return(nil, domain.ErrProyectoNoEncontrado)

			_, err := servicio.Acordar(context.Background(), proyectoPokerID, 9, anaPokerID, 5)

			assert.ErrorIs(t, err, domain.ErrProyectoNoEncontrado)
			poker.AssertNotCalled(t, "AcordarSesion", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
	})
}

func TestCancelarSesionPoker(t *testing.T) {
	inicio := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)
	canceladoEn := time.Date(2026, 11, 13, 18, 0, 0, 0, time.UTC)
	reveladaEn := time.Date(2026, 11, 13, 17, 0, 0, 0, time.UTC)

	preparar := func(t *testing.T) (*PokerService, *MockSesionPokerRepository) {
		t.Helper()
		servicio, proyectos, _, _, poker := nuevoServicioPokerDePrueba(t)
		proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).
			Return(proyectoDePrueba(proyectoPokerID, "Metrics", inicio, nil, domain.EstadoActivo), nil)
		return servicio, poker
	}
	sesionRevelada := func() *domain.SesionPoker {
		return &domain.SesionPoker{
			ID: 9, ProyectoID: proyectoPokerID, HistoriaID: historiaPokerID, CreadoPor: anaPokerID,
			CreadorNombre: "Ana", RondaActual: 1, Estado: domain.EstadoSesionRevelada,
		}
	}
	configurarDetalleCancelado := func(poker *MockSesionPokerRepository) {
		t.Helper()
		poker.On("ObtenerRonda", mock.Anything, uint(9), 1).
			Return(&domain.RondaPoker{ID: 3, SesionID: 9, NumeroRonda: 1, ReveladaEn: &reveladaEn}, nil)
		poker.On("ContarMiembros", mock.Anything, proyectoPokerID).Return(3, nil)
		poker.On("ContarVotosPorRonda", mock.Anything, uint(3)).Return(2, nil)
		poker.On("ObtenerVotoUsuario", mock.Anything, uint(3), benPokerID).Return(&domain.VotoPoker{
			RondaID: 3, UsuarioID: benPokerID, Nombre: "Ben", Puntos: 5,
		}, nil)
		poker.On("ListarVotosPorRonda", mock.Anything, uint(3)).Return([]domain.VotoPoker{
			{UsuarioID: anaPokerID, Nombre: "Ana", Puntos: 3},
			{UsuarioID: benPokerID, Nombre: "Ben", Puntos: 5},
		}, nil)
		poker.On("ListarRondasPorSesion", mock.Anything, uint(9)).Return([]domain.RondaPoker{
			{ID: 3, SesionID: 9, NumeroRonda: 1, ReveladaEn: &reveladaEn},
		}, nil)
	}

	t.Run("E-10 y E-11 cancela sin estimar y conserva votos y rondas", func(t *testing.T) {
		servicio, poker := preparar(t)
		poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).Return(sesionRevelada(), nil)
		actualizada := sesionRevelada()
		actualizada.Estado = domain.EstadoSesionCancelada
		actualizada.CanceladoEn = &canceladoEn
		poker.On("CancelarSesion", mock.Anything, uint(9)).Return(actualizada, nil)
		configurarDetalleCancelado(poker)

		detalle, err := servicio.Cancelar(context.Background(), proyectoPokerID, 9, benPokerID)

		require.NoError(t, err)
		assert.Equal(t, domain.EstadoSesionCancelada, detalle.Sesion.Estado)
		require.NotNil(t, detalle.Sesion.CanceladoEn)
		assert.Equal(t, canceladoEn, *detalle.Sesion.CanceladoEn)
		assert.Nil(t, detalle.Sesion.PuntosAcordados)
		assert.Nil(t, detalle.Sesion.CerradoEn)
		assert.Nil(t, detalle.Sesion.StoryEstimate)
		require.Len(t, detalle.HistorialRondas, 1)
		assert.Len(t, detalle.HistorialRondas[0].Votos, 2, "los votos de la ronda revelada se conservan")
		poker.AssertNotCalled(t, "AcordarSesion", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("E-13 y E-14 cancelar una sesión cerrada o cancelada devuelve SESION_POKER_CERRADA", func(t *testing.T) {
		for _, estado := range []domain.EstadoSesionPoker{domain.EstadoSesionCerrada, domain.EstadoSesionCancelada} {
			t.Run(string(estado), func(t *testing.T) {
				servicio, poker := preparar(t)
				sesion := sesionRevelada()
				sesion.Estado = estado
				poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).Return(sesion, nil)
				poker.On("CancelarSesion", mock.Anything, uint(9)).Return(nil, domain.ErrSesionPokerCerrada)

				_, err := servicio.Cancelar(context.Background(), proyectoPokerID, 9, benPokerID)

				assert.ErrorIs(t, err, domain.ErrSesionPokerCerrada)
				poker.AssertNotCalled(t, "ObtenerRonda", mock.Anything, mock.Anything, mock.Anything)
			})
		}
	})

	t.Run("sesión inexistente y proyecto inexistente", func(t *testing.T) {
		t.Run("sesión", func(t *testing.T) {
			servicio, poker := preparar(t)
			poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).
				Return(nil, domain.ErrSesionPokerNoEncontrada)

			_, err := servicio.Cancelar(context.Background(), proyectoPokerID, 9, benPokerID)

			assert.ErrorIs(t, err, domain.ErrSesionPokerNoEncontrada)
			poker.AssertNotCalled(t, "CancelarSesion", mock.Anything, mock.Anything)
		})
		t.Run("proyecto", func(t *testing.T) {
			servicio, proyectos, _, _, poker := nuevoServicioPokerDePrueba(t)
			proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).Return(nil, domain.ErrProyectoNoEncontrado)

			_, err := servicio.Cancelar(context.Background(), proyectoPokerID, 9, benPokerID)

			assert.ErrorIs(t, err, domain.ErrProyectoNoEncontrado)
			poker.AssertNotCalled(t, "CancelarSesion", mock.Anything, mock.Anything)
		})
	})
}

func TestDetalleSesionPokerCerradaIncluyeEstimacion(t *testing.T) {
	inicio := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)
	cerradoEn := time.Date(2026, 11, 13, 18, 0, 0, 0, time.UTC)
	servicio, proyectos, _, _, poker := nuevoServicioPokerDePrueba(t)

	proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).
		Return(proyectoDePrueba(proyectoPokerID, "Metrics", inicio, nil, domain.EstadoActivo), nil)
	sesion := &domain.SesionPoker{
		ID: 9, ProyectoID: proyectoPokerID, HistoriaID: historiaPokerID, CreadoPor: anaPokerID,
		CreadorNombre: "Ana", RondaActual: 1, Estado: domain.EstadoSesionCerrada,
		PuntosAcordados: punteroFloat64(5), CerradoEn: &cerradoEn,
	}
	poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, uint(9)).Return(sesion, nil)
	poker.On("ObtenerRonda", mock.Anything, uint(9), 1).
		Return(&domain.RondaPoker{ID: 3, SesionID: 9, NumeroRonda: 1, ReveladaEn: &cerradoEn}, nil)
	poker.On("ContarMiembros", mock.Anything, proyectoPokerID).Return(3, nil)
	poker.On("ContarVotosPorRonda", mock.Anything, uint(3)).Return(2, nil)
	poker.On("ObtenerVotoUsuario", mock.Anything, uint(3), anaPokerID).Return(nil, nil)
	poker.On("ListarVotosPorRonda", mock.Anything, uint(3)).Return([]domain.VotoPoker{
		{UsuarioID: anaPokerID, Nombre: "Ana", Puntos: 3},
		{UsuarioID: benPokerID, Nombre: "Ben", Puntos: 5},
	}, nil)
	poker.On("ListarRondasPorSesion", mock.Anything, uint(9)).Return([]domain.RondaPoker{
		{ID: 3, SesionID: 9, NumeroRonda: 1, ReveladaEn: &cerradoEn},
	}, nil)
	poker.On("ObtenerStoryEstimatePorSesion", mock.Anything, uint(9)).Return(&domain.StoryEstimate{
		ID: 12, HistoriaID: historiaPokerID, SesionID: 9, RondaID: 3,
		Puntos: 5, Fuente: "planning_poker", EstimadoPor: punteroIDPoker(anaPokerID), EstimadoEn: cerradoEn,
	}, nil)

	detalle, err := servicio.Detalle(context.Background(), proyectoPokerID, 9, anaPokerID)

	require.NoError(t, err)
	require.NotNil(t, detalle.Sesion.StoryEstimate)
	assert.Equal(t, 5.0, detalle.Sesion.StoryEstimate.Puntos)
	assert.Equal(t, uint(9), detalle.Sesion.StoryEstimate.SesionID)
}

func punteroFloat64(valor float64) *float64 {
	return &valor
}

func punteroIDPoker(valor uint) *uint {
	return &valor
}
