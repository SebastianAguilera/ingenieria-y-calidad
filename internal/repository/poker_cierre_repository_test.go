package repository

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"ingenieria-y-calidad/internal/domain"
	"ingenieria-y-calidad/internal/service"
)

// Tests de integración de US-17: cierre, cancelación, story_estimate y
// unidad de trabajo transaccional contra PostgreSQL.

// sesionReveladaConVotos crea una sesión, registra votos 3 (ana) y 5 (ben) y la
// revela. Deja la historia con 3 story points para verificar que el rollback la
// conserva.
func sesionReveladaConVotos(t *testing.T, db *gorm.DB) (proyectoID, historiaID, anaID, benID, sesionID, rondaID uint) {
	t.Helper()
	ctx := context.Background()
	repo := NuevoPokerRepository(db)
	proyectoID, historiaID, anaID, benID = prepararPoker(t, db)
	fijarStoryPoints(t, db, proyectoID, historiaID, 3)

	sesionID, rondaID = sesionVotandoConRonda(t, repo, proyectoID, historiaID, anaID)
	_, err := repo.RegistrarVoto(ctx, sesionID, anaID, 3)
	require.NoError(t, err)
	_, err = repo.RegistrarVoto(ctx, sesionID, benID, 5)
	require.NoError(t, err)
	require.NoError(t, repo.RevelarSesion(ctx, sesionID))
	return proyectoID, historiaID, anaID, benID, sesionID, rondaID
}

func fijarStoryPoints(t *testing.T, db *gorm.DB, proyectoID, historiaID uint, puntos float64) {
	t.Helper()
	historias := NuevoHistoriaRepository(db)
	historia, err := historias.ObtenerPorID(context.Background(), proyectoID, historiaID)
	require.NoError(t, err)
	historia.StoryPoints = &puntos
	require.NoError(t, historias.Actualizar(context.Background(), historia))
}

func storyPointsDe(t *testing.T, db *gorm.DB, proyectoID, historiaID uint) *float64 {
	t.Helper()
	historia, err := NuevoHistoriaRepository(db).ObtenerPorID(context.Background(), proyectoID, historiaID)
	require.NoError(t, err)
	return historia.StoryPoints
}

func contarEstimaciones(t *testing.T, db *gorm.DB, sesionID uint) int64 {
	t.Helper()
	var total int64
	require.NoError(t, db.Table("story_estimates").Where("sesion_id = ?", sesionID).Count(&total).Error)
	return total
}

// acordarEnTransaccion reproduce lo que hace PokerService.Acordar dentro de la
// unidad de trabajo, con un hook para forzar fallos en el último paso.
func acordarEnTransaccion(ctx context.Context, uow domain.UnidadDeTrabajo, proyectoID, historiaID, sesionID, usuarioID uint, puntos float64, rondaForzada *uint) error {
	return uow.Ejecutar(ctx, func(repos domain.RepositoriosTransaccion) error {
		ronda, err := repos.Poker().CerrarSesion(ctx, sesionID, puntos)
		if err != nil {
			return err
		}
		historia, err := repos.Historias().ObtenerPorID(ctx, proyectoID, historiaID)
		if err != nil {
			return err
		}
		historia.StoryPoints = &puntos
		if err := repos.Historias().Actualizar(ctx, historia); err != nil {
			return err
		}
		rondaID := ronda.ID
		if rondaForzada != nil {
			rondaID = *rondaForzada
		}
		estimacion := domain.NuevaEstimacionPlanningPoker(
			domain.SesionPoker{ID: sesionID, HistoriaID: historiaID}, rondaID, usuarioID, puntos, time.Now().UTC())
		return repos.Estimaciones().Crear(ctx, &estimacion)
	})
}

func TestMigracionAgregaColumnasDeUS17(t *testing.T) {
	db := abrirBaseDePruebas(t)

	assert.True(t, db.Migrator().HasColumn("poker_sessions", "cancelado_en"))
	assert.True(t, db.Migrator().HasColumn("story_estimates", "sesion_id"))
	assert.True(t, db.Migrator().HasColumn("story_estimates", "ronda_id"))
}

func TestCerrarSesionFijaPuntosYEstado(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoPokerRepository(db)
	proyectoID, _, _, _, sesionID, rondaID := sesionReveladaConVotos(t, db)
	ctx := context.Background()

	ronda, err := repo.CerrarSesion(ctx, sesionID, 5)

	require.NoError(t, err)
	assert.Equal(t, rondaID, ronda.ID, "devuelve la ronda actual para el story_estimate")
	sesion, err := repo.ObtenerPorID(ctx, proyectoID, sesionID)
	require.NoError(t, err)
	assert.Equal(t, domain.EstadoSesionCerrada, sesion.Estado)
	require.NotNil(t, sesion.PuntosAcordados)
	assert.Equal(t, 5.0, *sesion.PuntosAcordados)
	assert.NotNil(t, sesion.CerradoEn)
	assert.Nil(t, sesion.CanceladoEn)
}

func TestCerrarSesionFueraDeReveladaNoModificaNada(t *testing.T) {
	casos := []struct {
		estado   string
		esperado error
	}{
		{"Votando", domain.ErrSesionPokerNoRevelada},
		{"Cerrada", domain.ErrSesionPokerCerrada},
		{"Cancelada", domain.ErrSesionPokerCerrada},
	}
	for _, caso := range casos {
		t.Run(caso.estado, func(t *testing.T) {
			db := abrirBaseDePruebas(t)
			repo := NuevoPokerRepository(db)
			proyectoID, historiaID, anaID, _ := prepararPoker(t, db)
			ctx := context.Background()
			sesionID, _ := sesionVotandoConRonda(t, repo, proyectoID, historiaID, anaID)
			require.NoError(t, db.Exec("UPDATE poker_sessions SET estado = ? WHERE id = ?", caso.estado, sesionID).Error)

			_, err := repo.CerrarSesion(ctx, sesionID, 5)

			assert.ErrorIs(t, err, caso.esperado)
			sesion, err := repo.ObtenerPorID(ctx, proyectoID, sesionID)
			require.NoError(t, err)
			assert.Equal(t, domain.EstadoSesionPoker(caso.estado), sesion.Estado)
			assert.Nil(t, sesion.PuntosAcordados)
		})
	}
}

func TestCerrarSesionInexistente(t *testing.T) {
	db := abrirBaseDePruebas(t)

	_, err := NuevoPokerRepository(db).CerrarSesion(context.Background(), 999, 5)

	assert.ErrorIs(t, err, domain.ErrSesionPokerNoEncontrada)
}

func TestCancelarSesionActivaConservaVotosYLiberaLaHistoria(t *testing.T) {
	for _, revelar := range []bool{false, true} {
		nombre := "Votando"
		if revelar {
			nombre = "Revelada"
		}
		t.Run(nombre, func(t *testing.T) {
			db := abrirBaseDePruebas(t)
			repo := NuevoPokerRepository(db)
			proyectoID, historiaID, anaID, benID := prepararPoker(t, db)
			fijarStoryPoints(t, db, proyectoID, historiaID, 3)
			ctx := context.Background()
			sesionID, rondaID := sesionVotandoConRonda(t, repo, proyectoID, historiaID, anaID)
			_, err := repo.RegistrarVoto(ctx, sesionID, benID, 8)
			require.NoError(t, err)
			if revelar {
				require.NoError(t, repo.RevelarSesion(ctx, sesionID))
			}

			require.NoError(t, repo.CancelarSesion(ctx, sesionID))

			sesion, err := repo.ObtenerPorID(ctx, proyectoID, sesionID)
			require.NoError(t, err)
			assert.Equal(t, domain.EstadoSesionCancelada, sesion.Estado)
			assert.NotNil(t, sesion.CanceladoEn, "RB-09 setea cancelado_en")
			assert.Nil(t, sesion.PuntosAcordados)
			assert.Nil(t, sesion.CerradoEn)
			assert.Equal(t, 3.0, *storyPointsDe(t, db, proyectoID, historiaID), "RB-09 no toca la historia")
			assert.Zero(t, contarEstimaciones(t, db, sesionID), "RB-09 no crea story_estimate")

			total, err := repo.ContarVotosPorRonda(ctx, rondaID)
			require.NoError(t, err)
			assert.Equal(t, 1, total, "RB-10 conserva los votos")

			activa, err := repo.ObtenerActivaPorHistoria(ctx, historiaID)
			require.NoError(t, err)
			assert.Nil(t, activa)
			nueva := &domain.SesionPoker{ProyectoID: proyectoID, HistoriaID: historiaID, CreadoPor: anaID}
			assert.NoError(t, repo.CrearSesion(ctx, nueva), "RB-11 la historia queda libre para otra sesión")
		})
	}
}

func TestCancelarSesionFinalizadaNoModificaNada(t *testing.T) {
	for _, estado := range []string{"Cerrada", "Cancelada"} {
		t.Run(estado, func(t *testing.T) {
			db := abrirBaseDePruebas(t)
			repo := NuevoPokerRepository(db)
			proyectoID, historiaID, anaID, _ := prepararPoker(t, db)
			ctx := context.Background()
			sesionID, _ := sesionVotandoConRonda(t, repo, proyectoID, historiaID, anaID)
			require.NoError(t, db.Exec("UPDATE poker_sessions SET estado = ? WHERE id = ?", estado, sesionID).Error)

			err := repo.CancelarSesion(ctx, sesionID)

			assert.ErrorIs(t, err, domain.ErrSesionPokerCerrada)
			sesion, err := repo.ObtenerPorID(ctx, proyectoID, sesionID)
			require.NoError(t, err)
			assert.Equal(t, domain.EstadoSesionPoker(estado), sesion.Estado)
			assert.Nil(t, sesion.CanceladoEn)
		})
	}
}

func TestCancelarSesionInexistente(t *testing.T) {
	db := abrirBaseDePruebas(t)

	err := NuevoPokerRepository(db).CancelarSesion(context.Background(), 999)

	assert.ErrorIs(t, err, domain.ErrSesionPokerNoEncontrada)
}

func TestStoryEstimateCrearYListarPorSesiones(t *testing.T) {
	db := abrirBaseDePruebas(t)
	_, historiaID, anaID, _, sesionID, rondaID := sesionReveladaConVotos(t, db)
	ctx := context.Background()
	repo := NuevoStoryEstimateRepository(db)

	estimacion := domain.NuevaEstimacionPlanningPoker(
		domain.SesionPoker{ID: sesionID, HistoriaID: historiaID}, rondaID, anaID, 0.5, time.Now().UTC())
	require.NoError(t, repo.Crear(ctx, &estimacion))
	assert.NotZero(t, estimacion.ID)

	listadas, err := repo.ListarPorSesiones(ctx, []uint{sesionID, 999})
	require.NoError(t, err)
	require.Len(t, listadas, 1)
	assert.Equal(t, historiaID, listadas[0].HistoriaID)
	assert.Equal(t, sesionID, *listadas[0].SesionID)
	assert.Equal(t, rondaID, *listadas[0].RondaID)
	assert.Equal(t, anaID, *listadas[0].EstimadoPor)
	assert.Equal(t, 0.5, listadas[0].Puntos, "admite 0.5 (double precision)")
	assert.Equal(t, domain.FuenteEstimacionPlanningPoker, listadas[0].Fuente)

	vacias, err := repo.ListarPorSesiones(ctx, nil)
	require.NoError(t, err)
	assert.Empty(t, vacias)
}

// TestUnidadDeTrabajoConfirmaElAcuerdo cubre CA-6: historia, sesión y
// story_estimate quedan confirmados juntos.
func TestUnidadDeTrabajoConfirmaElAcuerdo(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectoID, historiaID, anaID, _, sesionID, _ := sesionReveladaConVotos(t, db)
	ctx := context.Background()

	err := acordarEnTransaccion(ctx, NuevaUnidadDeTrabajo(db), proyectoID, historiaID, sesionID, anaID, 5, nil)

	require.NoError(t, err)
	sesion, err := NuevoPokerRepository(db).ObtenerPorID(ctx, proyectoID, sesionID)
	require.NoError(t, err)
	assert.Equal(t, domain.EstadoSesionCerrada, sesion.Estado)
	assert.Equal(t, 5.0, *storyPointsDe(t, db, proyectoID, historiaID))
	assert.Equal(t, int64(1), contarEstimaciones(t, db, sesionID))
}

// TestUnidadDeTrabajoRevierteSiFallaElStoryEstimate cubre CA-5 / E-08: se fuerza
// un fallo real al insertar el story_estimate (ronda inexistente → violación de
// FK) y se verifica que ni la historia ni la sesión cambiaron.
func TestUnidadDeTrabajoRevierteSiFallaElStoryEstimate(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectoID, historiaID, anaID, _, sesionID, _ := sesionReveladaConVotos(t, db)
	ctx := context.Background()
	rondaInexistente := uint(999999)

	err := acordarEnTransaccion(ctx, NuevaUnidadDeTrabajo(db), proyectoID, historiaID, sesionID, anaID, 5, &rondaInexistente)

	require.Error(t, err)
	sesion, err := NuevoPokerRepository(db).ObtenerPorID(ctx, proyectoID, sesionID)
	require.NoError(t, err)
	assert.Equal(t, domain.EstadoSesionRevelada, sesion.Estado, "la sesión no se cerró")
	assert.Nil(t, sesion.PuntosAcordados)
	assert.Nil(t, sesion.CerradoEn)
	assert.Equal(t, 3.0, *storyPointsDe(t, db, proyectoID, historiaID), "story_points no cambió")
	assert.Zero(t, contarEstimaciones(t, db, sesionID), "no quedó story_estimate parcial")
}

// TestUnidadDeTrabajoRevierteAnteErrorDeNegocio verifica que cualquier error
// devuelto por fn, aunque no venga de la base, revierte todo lo escrito.
func TestUnidadDeTrabajoRevierteAnteErrorDeNegocio(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectoID, historiaID, _, _, sesionID, _ := sesionReveladaConVotos(t, db)
	ctx := context.Background()
	abortar := errors.New("abortar")

	err := NuevaUnidadDeTrabajo(db).Ejecutar(ctx, func(repos domain.RepositoriosTransaccion) error {
		if _, err := repos.Poker().CerrarSesion(ctx, sesionID, 8); err != nil {
			return err
		}
		historia, err := repos.Historias().ObtenerPorID(ctx, proyectoID, historiaID)
		if err != nil {
			return err
		}
		ocho := 8.0
		historia.StoryPoints = &ocho
		if err := repos.Historias().Actualizar(ctx, historia); err != nil {
			return err
		}
		return abortar
	})

	assert.ErrorIs(t, err, abortar)
	sesion, err := NuevoPokerRepository(db).ObtenerPorID(ctx, proyectoID, sesionID)
	require.NoError(t, err)
	assert.Equal(t, domain.EstadoSesionRevelada, sesion.Estado)
	assert.Equal(t, 3.0, *storyPointsDe(t, db, proyectoID, historiaID))
}

// TestAcuerdosSimultaneosProducenUnUnicoCierre cubre RB-07 / E-10.
func TestAcuerdosSimultaneosProducenUnUnicoCierre(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectoID, historiaID, anaID, benID, sesionID, _ := sesionReveladaConVotos(t, db)
	ctx := context.Background()
	uow := NuevaUnidadDeTrabajo(db)

	var wg sync.WaitGroup
	inicio := make(chan struct{})
	resultados := make([]error, 2)
	for idx, caso := range []struct {
		usuarioID uint
		puntos    float64
	}{{anaID, 5}, {benID, 8}} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-inicio
			resultados[idx] = acordarEnTransaccion(ctx, uow, proyectoID, historiaID, sesionID, caso.usuarioID, caso.puntos, nil)
		}()
	}
	close(inicio)
	wg.Wait()

	exitos := 0
	for _, err := range resultados {
		if err == nil {
			exitos++
			continue
		}
		assert.ErrorIs(t, err, domain.ErrSesionPokerCerrada)
	}
	assert.Equal(t, 1, exitos, "solo un acuerdo gana")
	assert.Equal(t, int64(1), contarEstimaciones(t, db, sesionID), "un único story_estimate")

	sesion, err := NuevoPokerRepository(db).ObtenerPorID(ctx, proyectoID, sesionID)
	require.NoError(t, err)
	assert.Equal(t, *sesion.PuntosAcordados, *storyPointsDe(t, db, proyectoID, historiaID),
		"la historia queda con los puntos del acuerdo ganador")
}

// TestFlujoCompletoConServicios integra PokerService y HistoriaService reales
// sobre PostgreSQL: acuerdo, consulta del backlog, cancelación y reapertura.
func TestFlujoCompletoConServicios(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectoID, historiaID, anaID, _, sesionID, _ := sesionReveladaConVotos(t, db)
	ctx := context.Background()

	proyectos := NuevoProyectoRepository(db)
	historias := NuevoHistoriaRepository(db)
	poker := NuevoPokerRepository(db)
	pokerService := service.NewPokerService(proyectos, historias, NuevoUsuarioRepository(db), poker,
		NuevoStoryEstimateRepository(db), NuevaUnidadDeTrabajo(db))
	historiaService := service.NewHistoriaService(proyectos, historias)

	detalle, err := pokerService.Acordar(ctx, proyectoID, sesionID, anaID, 8)
	require.NoError(t, err)
	assert.Equal(t, domain.EstadoSesionCerrada, detalle.Sesion.Estado)
	require.NotNil(t, detalle.Sesion.Estimacion)
	assert.Equal(t, 8.0, detalle.Sesion.Estimacion.Puntos)

	backlog, err := historiaService.Listar(ctx, proyectoID)
	require.NoError(t, err)
	require.Len(t, backlog, 1)
	require.NotNil(t, backlog[0].StoryPoints)
	assert.Equal(t, 8.0, *backlog[0].StoryPoints, "HistoriaService ve los story points acordados")

	_, err = pokerService.Cancelar(ctx, proyectoID, sesionID, anaID)
	assert.ErrorIs(t, err, domain.ErrSesionPokerCerrada)

	// Tras cerrar se puede abrir otra sesión y cancelarla; la historia no cambia.
	nueva, err := pokerService.CrearSesion(ctx, proyectoID, historiaID, anaID)
	require.NoError(t, err)
	cancelada, err := pokerService.Cancelar(ctx, proyectoID, nueva.Sesion.ID, anaID)
	require.NoError(t, err)
	assert.Equal(t, domain.EstadoSesionCancelada, cancelada.Sesion.Estado)
	assert.Equal(t, 8.0, *storyPointsDe(t, db, proyectoID, historiaID))

	historial, err := pokerService.ListarPorHistoria(ctx, proyectoID, historiaID)
	require.NoError(t, err)
	require.Len(t, historial, 2)
	assert.Equal(t, domain.EstadoSesionCancelada, historial[0].Estado)
	assert.NotNil(t, historial[0].CanceladoEn)
	assert.Nil(t, historial[0].Estimacion)
	assert.Equal(t, domain.EstadoSesionCerrada, historial[1].Estado)
	require.NotNil(t, historial[1].Estimacion)
	assert.Equal(t, 8.0, historial[1].Estimacion.Puntos)
}
