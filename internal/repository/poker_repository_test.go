package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"ingenieria-y-calidad/internal/domain"
)

// prepararPoker siembra un proyecto con dos miembros y una historia para los
// tests de persistencia de US-14.
func prepararPoker(t *testing.T, db *gorm.DB) (proyectoID, historiaID, anaID, benID uint) {
	t.Helper()
	ctx := context.Background()

	proyectos := NuevoProyectoRepository(db)
	usuarios := NuevoUsuarioRepository(db)
	historias := NuevoHistoriaRepository(db)

	ana := &domain.Usuario{Nombre: "Ana", Email: "ana.poker@utn.edu.ar"}
	ben := &domain.Usuario{Nombre: "Ben", Email: "ben.poker@utn.edu.ar"}
	require.NoError(t, usuarios.Crear(ctx, ana))
	require.NoError(t, usuarios.Crear(ctx, ben))

	inicio := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)
	proyecto := &domain.Proyecto{Nombre: "Poker", FechaInicio: inicio, Estado: domain.EstadoActivo}
	require.NoError(t, proyectos.CrearConUsuarios(ctx, proyecto, nil))
	require.NoError(t, usuarios.Vincular(ctx, proyecto.ID, ana.ID, domain.RolLider))
	require.NoError(t, usuarios.Vincular(ctx, proyecto.ID, ben.ID, domain.RolMiembro))

	historia := &domain.Historia{
		ProyectoID: proyecto.ID,
		Numero:     1,
		Codigo:     "US-01",
		Titulo:     "Estimar en equipo",
		Prioridad:  domain.PrioridadMedia,
		Estado:     domain.EstadoHistoriaBacklog,
	}
	require.NoError(t, historias.Crear(ctx, historia))

	return proyecto.ID, historia.ID, ana.ID, ben.ID
}

func TestCrearSesionPersisteEstadoYRondaUno(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoPokerRepository(db)
	proyectoID, historiaID, anaID, _ := prepararPoker(t, db)
	ctx := context.Background()

	sesion := &domain.SesionPoker{ProyectoID: proyectoID, HistoriaID: historiaID, CreadoPor: anaID}
	require.NoError(t, repo.CrearSesion(ctx, sesion))

	assert.NotZero(t, sesion.ID)
	assert.Equal(t, domain.EstadoSesionVotando, sesion.Estado)
	assert.Equal(t, 1, sesion.RondaActual)
	assert.False(t, sesion.CreadoEn.IsZero())

	ronda, err := repo.ObtenerRonda(ctx, sesion.ID, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, ronda.NumeroRonda)
	assert.Equal(t, sesion.ID, ronda.SesionID)
}

func TestIndiceParcialBloqueaSesionActivaYPermiteTrasCerrar(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoPokerRepository(db)
	proyectoID, historiaID, anaID, _ := prepararPoker(t, db)
	ctx := context.Background()

	primera := &domain.SesionPoker{ProyectoID: proyectoID, HistoriaID: historiaID, CreadoPor: anaID}
	require.NoError(t, repo.CrearSesion(ctx, primera))

	segunda := &domain.SesionPoker{ProyectoID: proyectoID, HistoriaID: historiaID, CreadoPor: anaID}
	err := repo.CrearSesion(ctx, segunda)
	assert.ErrorIs(t, err, domain.ErrSesionPokerActiva)

	var total int64
	require.NoError(t, db.Table("poker_sessions").
		Where("historia_id = ?", historiaID).
		Count(&total).Error)
	assert.Equal(t, int64(1), total, "la creación rechazada no debe persistir nada")

	require.NoError(t, db.Exec("UPDATE poker_sessions SET estado = 'Cerrada' WHERE id = ?", primera.ID).Error)

	tercera := &domain.SesionPoker{ProyectoID: proyectoID, HistoriaID: historiaID, CreadoPor: anaID}
	require.NoError(t, repo.CrearSesion(ctx, tercera))
	assert.NotZero(t, tercera.ID)
}

func TestDetalleActivasYVotos(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoPokerRepository(db)
	proyectoID, historiaID, anaID, benID := prepararPoker(t, db)
	ctx := context.Background()

	sesion := &domain.SesionPoker{ProyectoID: proyectoID, HistoriaID: historiaID, CreadoPor: anaID}
	require.NoError(t, repo.CrearSesion(ctx, sesion))

	ronda, err := repo.ObtenerRonda(ctx, sesion.ID, 1)
	require.NoError(t, err)
	require.NoError(t, db.Exec(
		"INSERT INTO poker_votes (ronda_id, usuario_id, puntos) VALUES (?, ?, ?), (?, ?, ?)",
		ronda.ID, anaID, 3.0, ronda.ID, benID, 5.0).Error)

	obtenida, err := repo.ObtenerPorID(ctx, proyectoID, sesion.ID)
	require.NoError(t, err)
	assert.Equal(t, "Ana", obtenida.CreadorNombre)

	totalMiembros, err := repo.ContarMiembros(ctx, proyectoID)
	require.NoError(t, err)
	assert.Equal(t, 2, totalMiembros)

	totalVotos, err := repo.ContarVotosPorRonda(ctx, ronda.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, totalVotos)

	votos, err := repo.ListarVotosPorRonda(ctx, ronda.ID)
	require.NoError(t, err)
	require.Len(t, votos, 2)
	assert.Equal(t, "Ana", votos[0].Nombre)
	assert.Equal(t, 3.0, votos[0].Puntos)
	assert.Equal(t, "Ben", votos[1].Nombre)

	votoBen, err := repo.ObtenerVotoUsuario(ctx, ronda.ID, benID)
	require.NoError(t, err)
	require.NotNil(t, votoBen)
	assert.Equal(t, 5.0, votoBen.Puntos)

	sinVoto, err := repo.ObtenerVotoUsuario(ctx, ronda.ID, anaID+benID+1000)
	require.NoError(t, err)
	assert.Nil(t, sinVoto)

	activa, err := repo.ObtenerActivaPorHistoria(ctx, historiaID)
	require.NoError(t, err)
	require.NotNil(t, activa)
	assert.Equal(t, sesion.ID, activa.ID)

	activas, err := repo.ListarActivasPorProyecto(ctx, proyectoID)
	require.NoError(t, err)
	require.Len(t, activas, 1)

	require.NoError(t, db.Exec("UPDATE poker_sessions SET estado = 'Revelada' WHERE id = ?", sesion.ID).Error)
	activas, err = repo.ListarActivasPorProyecto(ctx, proyectoID)
	require.NoError(t, err)
	assert.Len(t, activas, 1, "Revelada sigue siendo activa")

	require.NoError(t, db.Exec("UPDATE poker_sessions SET estado = 'Cancelada' WHERE id = ?", sesion.ID).Error)
	activas, err = repo.ListarActivasPorProyecto(ctx, proyectoID)
	require.NoError(t, err)
	assert.Empty(t, activas)

	historial, err := repo.ListarPorHistoria(ctx, historiaID)
	require.NoError(t, err)
	require.Len(t, historial, 1, "el historial incluye Cancelada")
}

func TestObtenerSesionDeOtroProyectoNoExiste(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoPokerRepository(db)
	proyectoID, historiaID, anaID, _ := prepararPoker(t, db)
	ctx := context.Background()

	sesion := &domain.SesionPoker{ProyectoID: proyectoID, HistoriaID: historiaID, CreadoPor: anaID}
	require.NoError(t, repo.CrearSesion(ctx, sesion))

	_, err := repo.ObtenerPorID(ctx, proyectoID+1, sesion.ID)
	assert.ErrorIs(t, err, domain.ErrSesionPokerNoEncontrada)
}

// sesionVotandoConRonda crea una sesión Votando y devuelve su id y el de la
// ronda 1 para los tests de votación de US-15.
func sesionVotandoConRonda(t *testing.T, repo *pokerRepository, proyectoID, historiaID, creadorID uint) (sesionID, rondaID uint) {
	t.Helper()
	ctx := context.Background()
	sesion := &domain.SesionPoker{ProyectoID: proyectoID, HistoriaID: historiaID, CreadoPor: creadorID}
	require.NoError(t, repo.CrearSesion(ctx, sesion))
	ronda, err := repo.ObtenerRonda(ctx, sesion.ID, 1)
	require.NoError(t, err)
	return sesion.ID, ronda.ID
}

func TestRegistrarVotoInsertaYActualizaSinDuplicar(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoPokerRepository(db)
	proyectoID, historiaID, anaID, benID := prepararPoker(t, db)
	ctx := context.Background()
	sesionID, rondaID := sesionVotandoConRonda(t, repo, proyectoID, historiaID, anaID)

	voto, err := repo.RegistrarVoto(ctx, sesionID, anaID, 5)
	require.NoError(t, err)
	assert.Equal(t, rondaID, voto.RondaID)
	assert.Equal(t, anaID, voto.UsuarioID)
	assert.Equal(t, 5.0, voto.Puntos)
	assert.False(t, voto.VotadoEn.IsZero())

	_, err = repo.RegistrarVoto(ctx, sesionID, benID, 13)
	require.NoError(t, err)

	cambiado, err := repo.RegistrarVoto(ctx, sesionID, anaID, 0.5)
	require.NoError(t, err)
	assert.Equal(t, 0.5, cambiado.Puntos)

	total, err := repo.ContarVotosPorRonda(ctx, rondaID)
	require.NoError(t, err)
	assert.Equal(t, 2, total, "cambiar el voto no crea una fila nueva")

	guardado, err := repo.ObtenerVotoUsuario(ctx, rondaID, anaID)
	require.NoError(t, err)
	require.NotNil(t, guardado)
	assert.Equal(t, 0.5, guardado.Puntos)
}

func TestRegistrarVotoFueraDeVotandoNoModificaNada(t *testing.T) {
	for _, estado := range []string{"Revelada", "Cerrada", "Cancelada"} {
		t.Run(estado, func(t *testing.T) {
			db := abrirBaseDePruebas(t)
			repo := NuevoPokerRepository(db)
			proyectoID, historiaID, anaID, benID := prepararPoker(t, db)
			ctx := context.Background()
			sesionID, rondaID := sesionVotandoConRonda(t, repo, proyectoID, historiaID, anaID)

			_, err := repo.RegistrarVoto(ctx, sesionID, anaID, 3)
			require.NoError(t, err)
			require.NoError(t, db.Exec("UPDATE poker_sessions SET estado = ? WHERE id = ?", estado, sesionID).Error)

			_, err = repo.RegistrarVoto(ctx, sesionID, anaID, 8)
			assert.ErrorIs(t, err, domain.ErrSesionPokerNoVotando, "no se puede cambiar un voto")
			_, err = repo.RegistrarVoto(ctx, sesionID, benID, 8)
			assert.ErrorIs(t, err, domain.ErrSesionPokerNoVotando, "no se puede insertar un voto")

			total, err := repo.ContarVotosPorRonda(ctx, rondaID)
			require.NoError(t, err)
			assert.Equal(t, 1, total)
			guardado, err := repo.ObtenerVotoUsuario(ctx, rondaID, anaID)
			require.NoError(t, err)
			require.NotNil(t, guardado)
			assert.Equal(t, 3.0, guardado.Puntos, "el voto original no cambia")
		})
	}
}

func TestRegistrarVotoDeSesionInexistente(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoPokerRepository(db)
	_, _, anaID, _ := prepararPoker(t, db)

	_, err := repo.RegistrarVoto(context.Background(), 99999, anaID, 5)

	assert.ErrorIs(t, err, domain.ErrSesionPokerNoEncontrada)
}

// TestCarreraVotarContraRevelar lanza votos concurrentes mientras otra
// operación revela la sesión con una transición condicionada al estado.
// Invariante: todo voto aceptado quedó guardado y todo voto rechazado recibió
// ErrSesionPokerNoVotando sin insertarse.
func TestCarreraVotarContraRevelar(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoPokerRepository(db)
	proyectoID, historiaID, anaID, benID := prepararPoker(t, db)
	ctx := context.Background()
	sesionID, rondaID := sesionVotandoConRonda(t, repo, proyectoID, historiaID, anaID)

	var wg sync.WaitGroup
	inicio := make(chan struct{})
	resultados := make([]error, 2)
	for idx, usuarioID := range []uint{anaID, benID} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-inicio
			_, resultados[idx] = repo.RegistrarVoto(ctx, sesionID, usuarioID, 5)
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-inicio
		// Simula el revelado de US-16: transición condicionada al estado.
		assert.NoError(t, db.Exec(
			"UPDATE poker_sessions SET estado = 'Revelada' WHERE id = ? AND estado = 'Votando'", sesionID).Error)
	}()
	close(inicio)
	wg.Wait()

	aceptados := 0
	for _, err := range resultados {
		if err == nil {
			aceptados++
			continue
		}
		assert.ErrorIs(t, err, domain.ErrSesionPokerNoVotando)
	}
	total, err := repo.ContarVotosPorRonda(ctx, rondaID)
	require.NoError(t, err)
	assert.Equal(t, aceptados, total, "solo se guardan los votos aceptados")

	// Con la sesión ya revelada, el voto tardío siempre pierde.
	_, err = repo.RegistrarVoto(ctx, sesionID, anaID, 8)
	assert.ErrorIs(t, err, domain.ErrSesionPokerNoVotando)
	totalFinal, err := repo.ContarVotosPorRonda(ctx, rondaID)
	require.NoError(t, err)
	assert.Equal(t, total, totalFinal)
}

// TestVotosSimultaneosDelMismoUsuarioNoSeDuplican verifica que el upsert
// converge en una única fila aunque lleguen varios votos a la vez (E-15).
func TestVotosSimultaneosDelMismoUsuarioNoSeDuplican(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoPokerRepository(db)
	proyectoID, historiaID, anaID, _ := prepararPoker(t, db)
	ctx := context.Background()
	sesionID, rondaID := sesionVotandoConRonda(t, repo, proyectoID, historiaID, anaID)

	var wg sync.WaitGroup
	inicio := make(chan struct{})
	errores := make([]error, 4)
	for idx, puntos := range []float64{1, 2, 3, 5} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-inicio
			_, errores[idx] = repo.RegistrarVoto(ctx, sesionID, anaID, puntos)
		}()
	}
	close(inicio)
	wg.Wait()

	for _, err := range errores {
		assert.NoError(t, err)
	}
	total, err := repo.ContarVotosPorRonda(ctx, rondaID)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
}

func TestRevelarSesionMarcaEstadoYRonda(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoPokerRepository(db)
	proyectoID, historiaID, anaID, benID := prepararPoker(t, db)
	ctx := context.Background()
	sesionID, rondaID := sesionVotandoConRonda(t, repo, proyectoID, historiaID, anaID)

	_, err := repo.RegistrarVoto(ctx, sesionID, anaID, 1)
	require.NoError(t, err)
	_, err = repo.RegistrarVoto(ctx, sesionID, benID, 8)
	require.NoError(t, err)

	require.NoError(t, repo.RevelarSesion(ctx, sesionID))

	sesion, err := repo.ObtenerPorID(ctx, proyectoID, sesionID)
	require.NoError(t, err)
	assert.Equal(t, domain.EstadoSesionRevelada, sesion.Estado)

	ronda, err := repo.ObtenerRonda(ctx, sesionID, 1)
	require.NoError(t, err)
	require.NotNil(t, ronda.ReveladaEn)
	assert.Equal(t, domain.EstadoSesionRevelada, ronda.EstadoRonda())

	total, err := repo.ContarVotosPorRonda(ctx, rondaID)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
}

func TestRevelarSesionSinVotosNoModificaNada(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoPokerRepository(db)
	proyectoID, historiaID, anaID, _ := prepararPoker(t, db)
	ctx := context.Background()
	sesionID, _ := sesionVotandoConRonda(t, repo, proyectoID, historiaID, anaID)

	err := repo.RevelarSesion(ctx, sesionID)
	assert.ErrorIs(t, err, domain.ErrSesionPokerSinVotos)

	sesion, err := repo.ObtenerPorID(ctx, proyectoID, sesionID)
	require.NoError(t, err)
	assert.Equal(t, domain.EstadoSesionVotando, sesion.Estado)
	ronda, err := repo.ObtenerRonda(ctx, sesionID, 1)
	require.NoError(t, err)
	assert.Nil(t, ronda.ReveladaEn)
}

func TestRevelarSesionFueraDeVotandoNoModificaNada(t *testing.T) {
	for _, estado := range []string{"Revelada", "Cerrada", "Cancelada"} {
		t.Run(estado, func(t *testing.T) {
			db := abrirBaseDePruebas(t)
			repo := NuevoPokerRepository(db)
			proyectoID, historiaID, anaID, _ := prepararPoker(t, db)
			ctx := context.Background()
			sesionID, _ := sesionVotandoConRonda(t, repo, proyectoID, historiaID, anaID)
			_, err := repo.RegistrarVoto(ctx, sesionID, anaID, 3)
			require.NoError(t, err)
			require.NoError(t, db.Exec("UPDATE poker_sessions SET estado = ? WHERE id = ?", estado, sesionID).Error)

			err = repo.RevelarSesion(ctx, sesionID)
			assert.ErrorIs(t, err, domain.ErrSesionPokerNoVotando)

			sesion, err := repo.ObtenerPorID(ctx, proyectoID, sesionID)
			require.NoError(t, err)
			assert.Equal(t, domain.EstadoSesionPoker(estado), sesion.Estado)
		})
	}
}

func TestRevelarSesionInexistente(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoPokerRepository(db)
	prepararPoker(t, db)

	err := repo.RevelarSesion(context.Background(), 99999)

	assert.ErrorIs(t, err, domain.ErrSesionPokerNoEncontrada)
}

func TestAbrirNuevaRondaConservaHistorial(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoPokerRepository(db)
	proyectoID, historiaID, anaID, benID := prepararPoker(t, db)
	ctx := context.Background()
	sesionID, rondaID := sesionVotandoConRonda(t, repo, proyectoID, historiaID, anaID)

	_, err := repo.RegistrarVoto(ctx, sesionID, anaID, 3)
	require.NoError(t, err)
	_, err = repo.RegistrarVoto(ctx, sesionID, benID, 5)
	require.NoError(t, err)
	require.NoError(t, repo.RevelarSesion(ctx, sesionID))

	nueva, err := repo.AbrirNuevaRonda(ctx, sesionID)
	require.NoError(t, err)
	assert.Equal(t, 2, nueva.NumeroRonda)
	assert.Nil(t, nueva.ReveladaEn)

	sesion, err := repo.ObtenerPorID(ctx, proyectoID, sesionID)
	require.NoError(t, err)
	assert.Equal(t, 2, sesion.RondaActual)
	assert.Equal(t, domain.EstadoSesionVotando, sesion.Estado)

	anterior, err := repo.ObtenerRonda(ctx, sesionID, 1)
	require.NoError(t, err)
	require.NotNil(t, anterior.ReveladaEn)
	totalAnterior, err := repo.ContarVotosPorRonda(ctx, rondaID)
	require.NoError(t, err)
	assert.Equal(t, 2, totalAnterior)

	rondas, err := repo.ListarRondasPorSesion(ctx, sesionID)
	require.NoError(t, err)
	require.Len(t, rondas, 2)
	assert.Equal(t, 1, rondas[0].NumeroRonda)
	assert.Equal(t, 2, rondas[1].NumeroRonda)

	_, err = repo.RegistrarVoto(ctx, sesionID, anaID, 8)
	require.NoError(t, err)
	totalAnterior, err = repo.ContarVotosPorRonda(ctx, rondaID)
	require.NoError(t, err)
	assert.Equal(t, 2, totalAnterior, "votar en la ronda nueva no toca la anterior")
}

func TestAbrirNuevaRondaFueraDeReveladaNoModificaNada(t *testing.T) {
	for _, estado := range []string{"Votando", "Cerrada", "Cancelada"} {
		t.Run(estado, func(t *testing.T) {
			db := abrirBaseDePruebas(t)
			repo := NuevoPokerRepository(db)
			proyectoID, historiaID, anaID, _ := prepararPoker(t, db)
			ctx := context.Background()
			sesionID, _ := sesionVotandoConRonda(t, repo, proyectoID, historiaID, anaID)
			require.NoError(t, db.Exec("UPDATE poker_sessions SET estado = ? WHERE id = ?", estado, sesionID).Error)

			_, err := repo.AbrirNuevaRonda(ctx, sesionID)
			assert.ErrorIs(t, err, domain.ErrSesionPokerNoRevelada)

			sesion, err := repo.ObtenerPorID(ctx, proyectoID, sesionID)
			require.NoError(t, err)
			assert.Equal(t, 1, sesion.RondaActual)
			rondas, err := repo.ListarRondasPorSesion(ctx, sesionID)
			require.NoError(t, err)
			assert.Len(t, rondas, 1)
		})
	}
}

func TestAbrirNuevaRondaSesionInexistente(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoPokerRepository(db)
	prepararPoker(t, db)

	_, err := repo.AbrirNuevaRonda(context.Background(), 99999)

	assert.ErrorIs(t, err, domain.ErrSesionPokerNoEncontrada)
}

func sesionReveladaConVoto(t *testing.T, repo *pokerRepository, proyectoID, historiaID, creadorID, votanteID uint) (sesionID, rondaID uint) {
	t.Helper()
	ctx := context.Background()
	sesionID, rondaID = sesionVotandoConRonda(t, repo, proyectoID, historiaID, creadorID)
	_, err := repo.RegistrarVoto(ctx, sesionID, votanteID, 3)
	require.NoError(t, err)
	require.NoError(t, repo.RevelarSesion(ctx, sesionID))
	return sesionID, rondaID
}

func TestAcordarSesionPersisteEstimacionYCierra(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoPokerRepository(db)
	historias := NuevoHistoriaRepository(db)
	proyectoID, historiaID, anaID, benID := prepararPoker(t, db)
	ctx := context.Background()
	sesionID, rondaID := sesionReveladaConVoto(t, repo, proyectoID, historiaID, anaID, benID)

	estimacion, err := repo.AcordarSesion(ctx, sesionID, anaID, 5)
	require.NoError(t, err)

	assert.NotZero(t, estimacion.ID)
	assert.Equal(t, historiaID, estimacion.HistoriaID)
	assert.Equal(t, sesionID, estimacion.SesionID)
	assert.Equal(t, rondaID, estimacion.RondaID)
	assert.Equal(t, 5.0, estimacion.Puntos)
	assert.Equal(t, "planning_poker", estimacion.Fuente)
	require.NotNil(t, estimacion.EstimadoPor)
	assert.Equal(t, anaID, *estimacion.EstimadoPor)
	assert.False(t, estimacion.EstimadoEn.IsZero())

	sesion, err := repo.ObtenerPorID(ctx, proyectoID, sesionID)
	require.NoError(t, err)
	assert.Equal(t, domain.EstadoSesionCerrada, sesion.Estado)
	require.NotNil(t, sesion.PuntosAcordados)
	assert.Equal(t, 5.0, *sesion.PuntosAcordados)
	require.NotNil(t, sesion.CerradoEn)
	assert.False(t, sesion.CerradoEn.IsZero())

	historia, err := historias.ObtenerPorID(ctx, proyectoID, historiaID)
	require.NoError(t, err)
	require.NotNil(t, historia.StoryPoints)
	assert.Equal(t, 5.0, *historia.StoryPoints)

	guardada, err := repo.ObtenerStoryEstimatePorSesion(ctx, sesionID)
	require.NoError(t, err)
	require.NotNil(t, guardada)
	assert.Equal(t, estimacion.ID, guardada.ID)
	assert.Equal(t, 5.0, guardada.Puntos)

	totalVotos, err := repo.ContarVotosPorRonda(ctx, rondaID)
	require.NoError(t, err)
	assert.Equal(t, 1, totalVotos, "el acuerdo conserva los votos")
}

func TestAcordarSesionFueraDeReveladaNoModificaNada(t *testing.T) {
	casos := []struct {
		estado string
		error  error
	}{
		{"Votando", domain.ErrSesionPokerNoRevelada},
		{"Cerrada", domain.ErrSesionPokerCerrada},
		{"Cancelada", domain.ErrSesionPokerCerrada},
	}

	for _, caso := range casos {
		t.Run(caso.estado, func(t *testing.T) {
			db := abrirBaseDePruebas(t)
			repo := NuevoPokerRepository(db)
			historias := NuevoHistoriaRepository(db)
			proyectoID, historiaID, anaID, benID := prepararPoker(t, db)
			ctx := context.Background()
			sesionID, _ := sesionVotandoConRonda(t, repo, proyectoID, historiaID, anaID)
			_, err := repo.RegistrarVoto(ctx, sesionID, benID, 3)
			require.NoError(t, err)
			require.NoError(t, db.Exec("UPDATE poker_sessions SET estado = ? WHERE id = ?", caso.estado, sesionID).Error)

			_, err = repo.AcordarSesion(ctx, sesionID, anaID, 5)

			assert.ErrorIs(t, err, caso.error)
			sesion, err := repo.ObtenerPorID(ctx, proyectoID, sesionID)
			require.NoError(t, err)
			assert.Equal(t, domain.EstadoSesionPoker(caso.estado), sesion.Estado)
			assert.Nil(t, sesion.PuntosAcordados)
			assert.Nil(t, sesion.CerradoEn)
			historia, err := historias.ObtenerPorID(ctx, proyectoID, historiaID)
			require.NoError(t, err)
			assert.Nil(t, historia.StoryPoints)
			guardada, err := repo.ObtenerStoryEstimatePorSesion(ctx, sesionID)
			require.NoError(t, err)
			assert.Nil(t, guardada)
		})
	}
}

func TestAcordarSesionInexistente(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoPokerRepository(db)
	prepararPoker(t, db)

	_, err := repo.AcordarSesion(context.Background(), 99999, 1, 5)

	assert.ErrorIs(t, err, domain.ErrSesionPokerNoEncontrada)
}

func forzarFalloEnStoryEstimates(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`
		CREATE OR REPLACE FUNCTION fallar_story_estimates_us17() RETURNS trigger AS $$
		BEGIN
			RAISE EXCEPTION 'fallo forzado en story_estimates';
		END;
		$$ LANGUAGE plpgsql;
		DROP TRIGGER IF EXISTS fallar_story_estimates_us17_trigger ON story_estimates;
		CREATE TRIGGER fallar_story_estimates_us17_trigger
		BEFORE INSERT ON story_estimates
		FOR EACH ROW EXECUTE FUNCTION fallar_story_estimates_us17();
	`).Error)
	t.Cleanup(func() {
		db.Exec("DROP TRIGGER IF EXISTS fallar_story_estimates_us17_trigger ON story_estimates")
		db.Exec("DROP FUNCTION IF EXISTS fallar_story_estimates_us17()")
	})
}

func TestAcordarSesionRevierteTodoSiFallaStoryEstimate(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoPokerRepository(db)
	historias := NuevoHistoriaRepository(db)
	proyectoID, historiaID, anaID, benID := prepararPoker(t, db)
	ctx := context.Background()
	sesionID, _ := sesionReveladaConVoto(t, repo, proyectoID, historiaID, anaID, benID)
	forzarFalloEnStoryEstimates(t, db)

	_, err := repo.AcordarSesion(ctx, sesionID, anaID, 5)

	require.Error(t, err)
	sesion, err := repo.ObtenerPorID(ctx, proyectoID, sesionID)
	require.NoError(t, err)
	assert.Equal(t, domain.EstadoSesionRevelada, sesion.Estado, "la sesión no debe quedar cerrada")
	assert.Nil(t, sesion.PuntosAcordados)
	assert.Nil(t, sesion.CerradoEn)
	historia, err := historias.ObtenerPorID(ctx, proyectoID, historiaID)
	require.NoError(t, err)
	assert.Nil(t, historia.StoryPoints, "story_points no debe cambiar")
	guardada, err := repo.ObtenerStoryEstimatePorSesion(ctx, sesionID)
	require.NoError(t, err)
	assert.Nil(t, guardada, "no debe quedar una estimación parcial")
}

func TestCancelarSesionNoEstimaYConservaVotos(t *testing.T) {
	for _, revelar := range []bool{false, true} {
		nombre := "Votando"
		if revelar {
			nombre = "Revelada"
		}
		t.Run(nombre, func(t *testing.T) {
			db := abrirBaseDePruebas(t)
			repo := NuevoPokerRepository(db)
			historias := NuevoHistoriaRepository(db)
			proyectoID, historiaID, anaID, benID := prepararPoker(t, db)
			ctx := context.Background()
			sesionID, rondaID := sesionVotandoConRonda(t, repo, proyectoID, historiaID, anaID)
			_, err := repo.RegistrarVoto(ctx, sesionID, benID, 5)
			require.NoError(t, err)
			if revelar {
				require.NoError(t, repo.RevelarSesion(ctx, sesionID))
			}

			actualizada, err := repo.CancelarSesion(ctx, sesionID)

			require.NoError(t, err)
			require.NotNil(t, actualizada)
			assert.Equal(t, domain.EstadoSesionCancelada, actualizada.Estado)
			require.NotNil(t, actualizada.CanceladoEn)
			assert.False(t, actualizada.CanceladoEn.IsZero())
			assert.Nil(t, actualizada.PuntosAcordados)
			assert.Nil(t, actualizada.CerradoEn)

			historia, err := historias.ObtenerPorID(ctx, proyectoID, historiaID)
			require.NoError(t, err)
			assert.Nil(t, historia.StoryPoints)

			guardada, err := repo.ObtenerStoryEstimatePorSesion(ctx, sesionID)
			require.NoError(t, err)
			assert.Nil(t, guardada)

			totalVotos, err := repo.ContarVotosPorRonda(ctx, rondaID)
			require.NoError(t, err)
			assert.Equal(t, 1, totalVotos, "cancelar conserva los votos")
			rondas, err := repo.ListarRondasPorSesion(ctx, sesionID)
			require.NoError(t, err)
			assert.Len(t, rondas, 1, "cancelar conserva las rondas")

			nueva := &domain.SesionPoker{ProyectoID: proyectoID, HistoriaID: historiaID, CreadoPor: anaID}
			require.NoError(t, repo.CrearSesion(ctx, nueva), "cancelar libera la historia")
			assert.NotZero(t, nueva.ID)
		})
	}
}

func TestCancelarSesionCerradaOCancelada(t *testing.T) {
	t.Run("Cerrada", func(t *testing.T) {
		db := abrirBaseDePruebas(t)
		repo := NuevoPokerRepository(db)
		proyectoID, historiaID, anaID, benID := prepararPoker(t, db)
		ctx := context.Background()
		sesionID, _ := sesionReveladaConVoto(t, repo, proyectoID, historiaID, anaID, benID)
		_, err := repo.AcordarSesion(ctx, sesionID, anaID, 5)
		require.NoError(t, err)

		_, err = repo.CancelarSesion(ctx, sesionID)

		assert.ErrorIs(t, err, domain.ErrSesionPokerCerrada)
		sesion, err := repo.ObtenerPorID(ctx, proyectoID, sesionID)
		require.NoError(t, err)
		assert.Equal(t, domain.EstadoSesionCerrada, sesion.Estado)
		assert.Nil(t, sesion.CanceladoEn)
	})

	t.Run("Cancelada", func(t *testing.T) {
		db := abrirBaseDePruebas(t)
		repo := NuevoPokerRepository(db)
		proyectoID, historiaID, anaID, benID := prepararPoker(t, db)
		ctx := context.Background()
		sesionID, _ := sesionReveladaConVoto(t, repo, proyectoID, historiaID, anaID, benID)
		_, err := repo.CancelarSesion(ctx, sesionID)
		require.NoError(t, err)

		_, err = repo.CancelarSesion(ctx, sesionID)

		assert.ErrorIs(t, err, domain.ErrSesionPokerCerrada)
	})
}

func TestCancelarSesionInexistente(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoPokerRepository(db)
	prepararPoker(t, db)

	_, err := repo.CancelarSesion(context.Background(), 99999)

	assert.ErrorIs(t, err, domain.ErrSesionPokerNoEncontrada)
}

func TestCarreraAcordarContraCancelar(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoPokerRepository(db)
	historias := NuevoHistoriaRepository(db)
	proyectoID, historiaID, anaID, benID := prepararPoker(t, db)
	ctx := context.Background()
	sesionID, _ := sesionReveladaConVoto(t, repo, proyectoID, historiaID, anaID, benID)

	var wg sync.WaitGroup
	inicio := make(chan struct{})
	var errAcuerdo, errCancelacion error
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-inicio
		_, errAcuerdo = repo.AcordarSesion(ctx, sesionID, anaID, 5)
	}()
	go func() {
		defer wg.Done()
		<-inicio
		_, errCancelacion = repo.CancelarSesion(ctx, sesionID)
	}()
	close(inicio)
	wg.Wait()

	ganadoras := 0
	if errAcuerdo == nil {
		ganadoras++
	} else {
		assert.ErrorIs(t, errAcuerdo, domain.ErrSesionPokerCerrada)
	}
	if errCancelacion == nil {
		ganadoras++
	} else {
		assert.ErrorIs(t, errCancelacion, domain.ErrSesionPokerCerrada)
	}
	assert.Equal(t, 1, ganadoras, "exactamente una operación debe ganar")

	sesion, err := repo.ObtenerPorID(ctx, proyectoID, sesionID)
	require.NoError(t, err)
	historia, err := historias.ObtenerPorID(ctx, proyectoID, historiaID)
	require.NoError(t, err)
	guardada, err := repo.ObtenerStoryEstimatePorSesion(ctx, sesionID)
	require.NoError(t, err)

	if errAcuerdo == nil {
		assert.Equal(t, domain.EstadoSesionCerrada, sesion.Estado)
		require.NotNil(t, historia.StoryPoints)
		assert.Equal(t, 5.0, *historia.StoryPoints)
		require.NotNil(t, guardada)
		assert.Equal(t, 5.0, guardada.Puntos)
		return
	}

	assert.Equal(t, domain.EstadoSesionCancelada, sesion.Estado)
	assert.Nil(t, historia.StoryPoints)
	assert.Nil(t, guardada)
}
