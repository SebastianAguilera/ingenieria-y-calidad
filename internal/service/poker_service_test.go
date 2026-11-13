package service

import (
	"context"
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

		sesiones, err := servicio.ListarPorHistoria(ctx, proyectoPokerID, historiaPokerID)

		require.NoError(t, err)
		require.Len(t, sesiones, 2)
		assert.Equal(t, domain.EstadoSesionCerrada, sesiones[1].Estado)
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

		detalle, err := servicio.Detalle(ctx, proyectoPokerID, 9, anaPokerID)

		require.NoError(t, err)
		assert.Equal(t, 2, detalle.TotalMiembros)
		assert.Equal(t, 1, detalle.VotosRegistrados)
		assert.True(t, detalle.YaVote)
		require.NotNil(t, detalle.MiVoto)
		assert.Equal(t, 3.0, *detalle.MiVoto)
		assert.Empty(t, detalle.Votos, "en Votando no se exponen votos ajenos")
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

		detalle, err := servicio.Detalle(ctx, proyectoPokerID, 9, anaPokerID)

		require.NoError(t, err)
		require.Len(t, detalle.Votos, 2)
		assert.Equal(t, "Ben", detalle.Votos[1].Nombre)
		assert.Equal(t, 5.0, detalle.Votos[1].Puntos)
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
