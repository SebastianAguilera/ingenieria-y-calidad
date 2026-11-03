package repository

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func crearHistoriaDePrueba(t *testing.T, repo *historiaRepository, proyectoID uint, numero int) *domain.Historia {
	t.Helper()
	historia := &domain.Historia{
		ProyectoID: proyectoID,
		Numero:     numero,
		Codigo:     domain.FormatearCodigoHistoria(numero),
		Titulo:     fmt.Sprintf("Historia %d", numero),
		Prioridad:  domain.PrioridadMedia,
		Estado:     domain.EstadoHistoriaBacklog,
	}
	require.NoError(t, repo.Crear(ctxBackground(), historia))
	return historia
}

func TestSprintBacklogAsignarPersisteItem(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	sprints := NuevoSprintRepository(db)
	historias := NuevoHistoriaRepository(db)
	backlog := NuevoSprintBacklogRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Con sprint backlog")
	sprint := crearSprintDePrueba(t, sprints, proyecto.ID, "Sprint 1",
		sprintFecha(time.October, 6), sprintFecha(time.October, 20), domain.EstadoSprintActivo)
	historia := crearHistoriaDePrueba(t, historias, proyecto.ID, 1)

	item := &domain.SprintBacklogItem{
		ProyectoID: proyecto.ID,
		SprintID:   sprint.ID,
		HistoriaID: historia.ID,
		Orden:      1,
	}
	require.NoError(t, backlog.Asignar(ctx, item))

	assert.NotZero(t, item.ID)
	assert.False(t, item.CreadoEn.IsZero(), "creado_en debe informarse al asignar")
	assert.False(t, item.ActualizadoEn.IsZero(), "actualizado_en debe informarse al asignar")
	assert.False(t, item.FechaAsignacion.IsZero(), "la fecha de asignacion se completa sola")

	recuperado, err := backlog.ObtenerPorHistoria(ctx, historia.ID)
	require.NoError(t, err)
	require.NotNil(t, recuperado)
	assert.Equal(t, item.ID, recuperado.ID)
	assert.Equal(t, proyecto.ID, recuperado.ProyectoID)
	assert.Equal(t, sprint.ID, recuperado.SprintID)
	assert.Equal(t, historia.ID, recuperado.HistoriaID)
	assert.Equal(t, 1, recuperado.Orden)
	assert.WithinDuration(t, item.FechaAsignacion, recuperado.FechaAsignacion, time.Second)
}

func TestSprintBacklogAsignarRespetaFechaExplicita(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	sprints := NuevoSprintRepository(db)
	historias := NuevoHistoriaRepository(db)
	backlog := NuevoSprintBacklogRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Con fecha explicita")
	sprint := crearSprintDePrueba(t, sprints, proyecto.ID, "Sprint 1",
		sprintFecha(time.October, 6), sprintFecha(time.October, 20), domain.EstadoSprintActivo)
	historia := crearHistoriaDePrueba(t, historias, proyecto.ID, 1)

	fecha := time.Date(2026, 10, 7, 9, 30, 0, 0, time.UTC)
	item := &domain.SprintBacklogItem{
		ProyectoID:      proyecto.ID,
		SprintID:        sprint.ID,
		HistoriaID:      historia.ID,
		Orden:           1,
		FechaAsignacion: fecha,
	}
	require.NoError(t, backlog.Asignar(ctx, item))
	assert.Equal(t, fecha, item.FechaAsignacion)

	recuperado, err := backlog.ObtenerPorHistoria(ctx, historia.ID)
	require.NoError(t, err)
	require.NotNil(t, recuperado)
	assert.True(t, fecha.Equal(recuperado.FechaAsignacion),
		"se conserva el instante aunque cambie la zona horaria: %v", recuperado.FechaAsignacion)
}

func TestSprintBacklogHistoriaDuplicadaDevuelveError(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	sprints := NuevoSprintRepository(db)
	historias := NuevoHistoriaRepository(db)
	backlog := NuevoSprintBacklogRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Con duplicado")
	sprint := crearSprintDePrueba(t, sprints, proyecto.ID, "Sprint 1",
		sprintFecha(time.October, 6), sprintFecha(time.October, 20), domain.EstadoSprintActivo)
	historia := crearHistoriaDePrueba(t, historias, proyecto.ID, 1)

	require.NoError(t, backlog.Asignar(ctx, &domain.SprintBacklogItem{
		ProyectoID: proyecto.ID,
		SprintID:   sprint.ID,
		HistoriaID: historia.ID,
		Orden:      1,
	}))

	err := backlog.Asignar(ctx, &domain.SprintBacklogItem{
		ProyectoID: proyecto.ID,
		SprintID:   sprint.ID,
		HistoriaID: historia.ID,
		Orden:      2,
	})

	assert.ErrorIs(t, err, domain.ErrHistoriaYaAsignada)
}

func TestSprintBacklogExisteAsignacionHistoria(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	sprints := NuevoSprintRepository(db)
	historias := NuevoHistoriaRepository(db)
	backlog := NuevoSprintBacklogRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Con existencia")
	sprint := crearSprintDePrueba(t, sprints, proyecto.ID, "Sprint 1",
		sprintFecha(time.October, 6), sprintFecha(time.October, 20), domain.EstadoSprintActivo)
	historia := crearHistoriaDePrueba(t, historias, proyecto.ID, 1)

	existe, item, err := backlog.ExisteAsignacionHistoria(ctx, historia.ID)
	require.NoError(t, err)
	assert.False(t, existe)
	assert.Nil(t, item)

	require.NoError(t, backlog.Asignar(ctx, &domain.SprintBacklogItem{
		ProyectoID: proyecto.ID,
		SprintID:   sprint.ID,
		HistoriaID: historia.ID,
		Orden:      3,
	}))

	existe, item, err = backlog.ExisteAsignacionHistoria(ctx, historia.ID)
	require.NoError(t, err)
	assert.True(t, existe)
	require.NotNil(t, item)
	assert.Equal(t, sprint.ID, item.SprintID)
	assert.Equal(t, 3, item.Orden)
}

func TestSprintBacklogDesasignarEliminaItem(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	sprints := NuevoSprintRepository(db)
	historias := NuevoHistoriaRepository(db)
	backlog := NuevoSprintBacklogRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Con desasignacion")
	sprint := crearSprintDePrueba(t, sprints, proyecto.ID, "Sprint 1",
		sprintFecha(time.October, 6), sprintFecha(time.October, 20), domain.EstadoSprintActivo)
	historia := crearHistoriaDePrueba(t, historias, proyecto.ID, 1)

	require.NoError(t, backlog.Asignar(ctx, &domain.SprintBacklogItem{
		ProyectoID: proyecto.ID,
		SprintID:   sprint.ID,
		HistoriaID: historia.ID,
		Orden:      1,
	}))

	require.NoError(t, backlog.Desasignar(ctx, proyecto.ID, sprint.ID, historia.ID))

	existe, item, err := backlog.ExisteAsignacionHistoria(ctx, historia.ID)
	require.NoError(t, err)
	assert.False(t, existe)
	assert.Nil(t, item)

	recuperado, err := backlog.ObtenerPorHistoria(ctx, historia.ID)
	require.NoError(t, err)
	assert.Nil(t, recuperado)

	err = backlog.Desasignar(ctx, proyecto.ID, sprint.ID, historia.ID)
	assert.ErrorIs(t, err, domain.ErrHistoriaNoPerteneceASprint)
}

func TestSprintBacklogObtenerPorSprintOrdenaYAisla(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	sprints := NuevoSprintRepository(db)
	historias := NuevoHistoriaRepository(db)
	backlog := NuevoSprintBacklogRepository(db)
	ctx := ctxBackground()

	proyectoA := crearProyectoDePrueba(t, proyectos, "Proyecto A")
	proyectoB := crearProyectoDePrueba(t, proyectos, "Proyecto B")

	sprintA1 := crearSprintDePrueba(t, sprints, proyectoA.ID, "Sprint 1",
		sprintFecha(time.October, 6), sprintFecha(time.October, 20), domain.EstadoSprintActivo)
	sprintA2 := crearSprintDePrueba(t, sprints, proyectoA.ID, "Sprint 2",
		sprintFecha(time.October, 21), sprintFecha(time.November, 4), domain.EstadoSprintActivo)
	sprintB1 := crearSprintDePrueba(t, sprints, proyectoB.ID, "Sprint 1",
		sprintFecha(time.October, 6), sprintFecha(time.October, 20), domain.EstadoSprintActivo)

	historiaA1 := crearHistoriaDePrueba(t, historias, proyectoA.ID, 1)
	historiaA2 := crearHistoriaDePrueba(t, historias, proyectoA.ID, 2)
	historiaA3 := crearHistoriaDePrueba(t, historias, proyectoA.ID, 3)
	historiaA4 := crearHistoriaDePrueba(t, historias, proyectoA.ID, 4)
	historiaB1 := crearHistoriaDePrueba(t, historias, proyectoB.ID, 1)

	// Mismo orden (1): el desempate es por id ascendente, o sea por orden de alta.
	require.NoError(t, backlog.Asignar(ctx, &domain.SprintBacklogItem{
		ProyectoID: proyectoA.ID, SprintID: sprintA1.ID, HistoriaID: historiaA2.ID, Orden: 1,
	}))
	require.NoError(t, backlog.Asignar(ctx, &domain.SprintBacklogItem{
		ProyectoID: proyectoA.ID, SprintID: sprintA1.ID, HistoriaID: historiaA1.ID, Orden: 1,
	}))
	require.NoError(t, backlog.Asignar(ctx, &domain.SprintBacklogItem{
		ProyectoID: proyectoA.ID, SprintID: sprintA1.ID, HistoriaID: historiaA3.ID, Orden: 0,
	}))
	require.NoError(t, backlog.Asignar(ctx, &domain.SprintBacklogItem{
		ProyectoID: proyectoA.ID, SprintID: sprintA2.ID, HistoriaID: historiaA4.ID, Orden: 1,
	}))
	require.NoError(t, backlog.Asignar(ctx, &domain.SprintBacklogItem{
		ProyectoID: proyectoB.ID, SprintID: sprintB1.ID, HistoriaID: historiaB1.ID, Orden: 1,
	}))

	listaA1, err := backlog.ObtenerPorSprint(ctx, proyectoA.ID, sprintA1.ID)
	require.NoError(t, err)
	require.Len(t, listaA1, 3)
	assert.Equal(t, historiaA3.ID, listaA1[0].HistoriaID, "orden 0 va primero")
	assert.Equal(t, historiaA2.ID, listaA1[1].HistoriaID, "a igual orden gana el id menor")
	assert.Equal(t, historiaA1.ID, listaA1[2].HistoriaID)

	listaA2, err := backlog.ObtenerPorSprint(ctx, proyectoA.ID, sprintA2.ID)
	require.NoError(t, err)
	require.Len(t, listaA2, 1)
	assert.Equal(t, historiaA4.ID, listaA2[0].HistoriaID)

	listaB1, err := backlog.ObtenerPorSprint(ctx, proyectoB.ID, sprintB1.ID)
	require.NoError(t, err)
	require.Len(t, listaB1, 1)
	assert.Equal(t, historiaB1.ID, listaB1[0].HistoriaID)

	vacia, err := backlog.ObtenerPorSprint(ctx, proyectoB.ID, sprintA1.ID)
	require.NoError(t, err)
	assert.NotNil(t, vacia, "el listado vacio no debe ser nil")
	assert.Empty(t, vacia)
}

func TestSprintBacklogObtenerMaxOrden(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	sprints := NuevoSprintRepository(db)
	historias := NuevoHistoriaRepository(db)
	backlog := NuevoSprintBacklogRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Con orden maximo")
	sprint1 := crearSprintDePrueba(t, sprints, proyecto.ID, "Sprint 1",
		sprintFecha(time.October, 6), sprintFecha(time.October, 20), domain.EstadoSprintActivo)
	sprint2 := crearSprintDePrueba(t, sprints, proyecto.ID, "Sprint 2",
		sprintFecha(time.October, 21), sprintFecha(time.November, 4), domain.EstadoSprintActivo)

	maximo, err := backlog.ObtenerMaxOrden(ctx, proyecto.ID, sprint1.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, maximo, "sin asignaciones el maximo es 0")

	historia1 := crearHistoriaDePrueba(t, historias, proyecto.ID, 1)
	historia2 := crearHistoriaDePrueba(t, historias, proyecto.ID, 2)
	historia3 := crearHistoriaDePrueba(t, historias, proyecto.ID, 3)

	require.NoError(t, backlog.Asignar(ctx, &domain.SprintBacklogItem{
		ProyectoID: proyecto.ID, SprintID: sprint1.ID, HistoriaID: historia1.ID, Orden: 1,
	}))
	require.NoError(t, backlog.Asignar(ctx, &domain.SprintBacklogItem{
		ProyectoID: proyecto.ID, SprintID: sprint1.ID, HistoriaID: historia2.ID, Orden: 4,
	}))
	require.NoError(t, backlog.Asignar(ctx, &domain.SprintBacklogItem{
		ProyectoID: proyecto.ID, SprintID: sprint2.ID, HistoriaID: historia3.ID, Orden: 10,
	}))

	maximo, err = backlog.ObtenerMaxOrden(ctx, proyecto.ID, sprint1.ID)
	require.NoError(t, err)
	assert.Equal(t, 4, maximo)

	maximo, err = backlog.ObtenerMaxOrden(ctx, proyecto.ID, sprint2.ID)
	require.NoError(t, err)
	assert.Equal(t, 10, maximo)

	maximo, err = backlog.ObtenerMaxOrden(ctx, proyecto.ID, 999999)
	require.NoError(t, err)
	assert.Equal(t, 0, maximo)
}
