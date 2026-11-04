package repository

import (
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
		Titulo:     "Historia de prueba",
		Prioridad:  domain.PrioridadMedia,
		Estado:     domain.EstadoHistoriaBacklog,
	}
	require.NoError(t, repo.Crear(ctxBackground(), historia))
	return historia
}

func TestDefectoRepositorioCicloCompleto(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	historias := NuevoHistoriaRepository(db)
	sprints := NuevoSprintRepository(db)
	defectos := NuevoDefectoRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Con defectos")
	historia := crearHistoriaDePrueba(t, historias, proyecto.ID, 1)
	deteccion := crearSprintDePrueba(t, sprints, proyecto.ID, "Sprint 1",
		sprintFecha(time.November, 1), sprintFecha(time.November, 15), domain.EstadoSprintActivo)
	resolucion := crearSprintDePrueba(t, sprints, proyecto.ID, "Sprint 2",
		sprintFecha(time.November, 15), sprintFecha(time.November, 30), domain.EstadoSprintPlanificado)

	defecto := &domain.Defecto{
		ProyectoID:        proyecto.ID,
		HistoriaID:        historia.ID,
		SprintDeteccionID: deteccion.ID,
		Descripcion:       "Error en validación",
		Severidad:         domain.SeveridadMajor,
		Estado:            domain.EstadoDefectoAbierto,
	}
	require.NoError(t, defectos.Crear(ctx, defecto))
	assert.NotZero(t, defecto.ID)
	assert.False(t, defecto.CreadoEn.IsZero(), "escenario BDD BDD-01")
	assert.Nil(t, defecto.ResueltoEn)

	recuperado, err := defectos.ObtenerPorID(ctx, proyecto.ID, defecto.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.EstadoDefectoAbierto, recuperado.Estado)
	assert.Equal(t, domain.SeveridadMajor, recuperado.Severidad)
	assert.Equal(t, historia.ID, recuperado.HistoriaID)
	assert.Equal(t, deteccion.ID, recuperado.SprintDeteccionID)
	assert.Nil(t, recuperado.SprintResolucionID)

	lista, err := defectos.ListarPorProyecto(ctx, proyecto.ID)
	require.NoError(t, err)
	require.Len(t, lista, 1, "escenario BDD BDD-16")
	assert.Equal(t, defecto.ID, lista[0].ID)

	recuperado.Estado = domain.EstadoDefectoResuelto
	recuperado.SprintResolucionID = &resolucion.ID
	ahora := time.Now().UTC()
	recuperado.ResueltoEn = &ahora
	require.NoError(t, defectos.Actualizar(ctx, recuperado))

	actualizado, err := defectos.ObtenerPorID(ctx, proyecto.ID, defecto.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.EstadoDefectoResuelto, actualizado.Estado, "escenario BDD BDD-08")
	require.NotNil(t, actualizado.SprintResolucionID)
	assert.Equal(t, resolucion.ID, *actualizado.SprintResolucionID)
	require.NotNil(t, actualizado.ResueltoEn)

	actualizado.Estado = domain.EstadoDefectoAbierto
	actualizado.SprintResolucionID = nil
	actualizado.ResueltoEn = nil
	require.NoError(t, defectos.Actualizar(ctx, actualizado))

	reabierto, err := defectos.ObtenerPorID(ctx, proyecto.ID, defecto.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.EstadoDefectoAbierto, reabierto.Estado, "escenario BDD BDD-15")
	assert.Nil(t, reabierto.SprintResolucionID)
	assert.Nil(t, reabierto.ResueltoEn)
}

func TestDefectoDeOtroProyectoEsInvisible(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	historias := NuevoHistoriaRepository(db)
	sprints := NuevoSprintRepository(db)
	defectos := NuevoDefectoRepository(db)
	ctx := ctxBackground()

	origen := crearProyectoDePrueba(t, proyectos, "Origen")
	destino := crearProyectoDePrueba(t, proyectos, "Destino")
	historia := crearHistoriaDePrueba(t, historias, origen.ID, 1)
	sprint := crearSprintDePrueba(t, sprints, origen.ID, "Sprint 1",
		sprintFecha(time.November, 1), sprintFecha(time.November, 15), domain.EstadoSprintActivo)

	defecto := &domain.Defecto{
		ProyectoID:        origen.ID,
		HistoriaID:        historia.ID,
		SprintDeteccionID: sprint.ID,
		Descripcion:       "Solo de origen",
		Severidad:         domain.SeveridadMinor,
		Estado:            domain.EstadoDefectoAbierto,
	}
	require.NoError(t, defectos.Crear(ctx, defecto))

	_, err := defectos.ObtenerPorID(ctx, destino.ID, defecto.ID)
	assert.ErrorIs(t, err, domain.ErrDefectoNoEncontrado, "escenario BDD BDD-14")

	lista, err := defectos.ListarPorProyecto(ctx, destino.ID)
	require.NoError(t, err)
	assert.Empty(t, lista)
}

func TestDefectoListarSinRegistrosDevuelveVacio(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	defectos := NuevoDefectoRepository(db)

	proyecto := crearProyectoDePrueba(t, proyectos, "Sin defectos")

	lista, err := defectos.ListarPorProyecto(ctxBackground(), proyecto.ID)
	require.NoError(t, err)
	assert.NotNil(t, lista, "el listado vacio no debe ser nil")
	assert.Empty(t, lista, "escenario BDD BDD-17")
}

func TestDefectoActualizarInexistenteDevuelveError(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	defectos := NuevoDefectoRepository(db)

	proyecto := crearProyectoDePrueba(t, proyectos, "Vacio")

	err := defectos.Actualizar(ctxBackground(), &domain.Defecto{
		ID: 999, ProyectoID: proyecto.ID, Estado: domain.EstadoDefectoEnProgreso,
	})
	assert.ErrorIs(t, err, domain.ErrDefectoNoEncontrado, "escenario BDD BDD-14")
}
