package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

// sprintFecha construye una fecha de calendario UTC para los Sprints de
// prueba, en el mismo formato "YYYY-MM-DD" que expone la API.
func sprintFecha(mes time.Month, dia int) time.Time {
	return time.Date(2026, mes, dia, 0, 0, 0, 0, time.UTC)
}

func crearSprintDePrueba(t *testing.T, repo *sprintRepository, proyectoID uint, identificador string, inicio, fin time.Time, estado domain.EstadoSprint) *domain.Sprint {
	t.Helper()
	sprint := &domain.Sprint{
		ProyectoID:    proyectoID,
		Identificador: identificador,
		SprintGoal:    "Goal de " + identificador,
		Estado:        estado,
		FechaInicio:   inicio,
		FechaFin:      fin,
	}
	require.NoError(t, repo.Crear(ctxBackground(), sprint))
	return sprint
}

func TestSprintRepositorioCicloCompleto(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	sprints := NuevoSprintRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Con sprints")

	primero := crearSprintDePrueba(t, sprints, proyecto.ID, "Sprint 1",
		sprintFecha(time.October, 6), sprintFecha(time.October, 20), domain.EstadoSprintPlanificado)
	assert.NotZero(t, primero.ID)
	assert.False(t, primero.CreadoEn.IsZero(), "escenario BDD E-22")
	assert.NotZero(t, primero.ActualizadoEn)

	segundo := crearSprintDePrueba(t, sprints, proyecto.ID, "Sprint 2",
		sprintFecha(time.October, 20), sprintFecha(time.November, 3), domain.EstadoSprintPlanificado)
	require.NotZero(t, segundo.ID)

	recuperado, err := sprints.ObtenerPorID(ctx, proyecto.ID, primero.ID)
	require.NoError(t, err)
	assert.Equal(t, "Sprint 1", recuperado.Identificador)
	assert.Equal(t, domain.EstadoSprintPlanificado, recuperado.Estado)

	lista, err := sprints.ListarPorProyecto(ctx, proyecto.ID)
	require.NoError(t, err)
	require.Len(t, lista, 2, "escenario BDD E-23")
	assert.Equal(t, "Sprint 1", lista[0].Identificador, "el listado va ordenado por fecha_inicio")
	assert.Equal(t, "Sprint 2", lista[1].Identificador)

	recuperado.Estado = domain.EstadoSprintActivo
	require.NoError(t, sprints.ActualizarEstado(ctx, recuperado))

	actualizado, err := sprints.ObtenerPorID(ctx, proyecto.ID, primero.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.EstadoSprintActivo, actualizado.Estado)
	assert.False(t, actualizado.ActualizadoEn.Before(actualizado.CreadoEn), "escenario BDD E-40")

	activo, err := sprints.ObtenerActivo(ctx, proyecto.ID)
	require.NoError(t, err)
	require.NotNil(t, activo, "escenario BDD E-29")
	assert.Equal(t, primero.ID, activo.ID)

	require.NoError(t, sprints.ActualizarEstado(ctx, &domain.Sprint{
		ID: primero.ID, ProyectoID: proyecto.ID, Estado: domain.EstadoSprintCerrado,
	}))

	sinActivo, err := sprints.ObtenerActivo(ctx, proyecto.ID)
	require.NoError(t, err)
	assert.Nil(t, sinActivo, "no debe quedar ningun sprint activo")
}

func TestSprintIdentificadorDuplicadoEnElMismoProyecto(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	sprints := NuevoSprintRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Con colision")

	crearSprintDePrueba(t, sprints, proyecto.ID, "Sprint 1",
		sprintFecha(time.October, 6), sprintFecha(time.October, 20), domain.EstadoSprintPlanificado)

	err := sprints.Crear(ctx, &domain.Sprint{
		ProyectoID:    proyecto.ID,
		Identificador: "Sprint 1",
		SprintGoal:    "Reintento",
		Estado:        domain.EstadoSprintPlanificado,
		FechaInicio:   sprintFecha(time.October, 21),
		FechaFin:      sprintFecha(time.October, 28),
	})
	assert.ErrorIs(t, err, domain.ErrIdentificadorSprintDuplicado, "escenario BDD E-12")
}

func TestSprintMismoIdentificadorEnOtroProyectoEstaPermitido(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	sprints := NuevoSprintRepository(db)

	origen := crearProyectoDePrueba(t, proyectos, "Origen")
	destino := crearProyectoDePrueba(t, proyectos, "Destino")

	crearSprintDePrueba(t, sprints, origen.ID, "Sprint 1",
		sprintFecha(time.October, 6), sprintFecha(time.October, 20), domain.EstadoSprintPlanificado)

	enDestino := crearSprintDePrueba(t, sprints, destino.ID, "Sprint 1",
		sprintFecha(time.October, 6), sprintFecha(time.October, 20), domain.EstadoSprintPlanificado)
	assert.NotZero(t, enDestino.ID, "escenario BDD E-13")
}

func TestSprintDeOtroProyectoEsInvisible(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	sprints := NuevoSprintRepository(db)
	ctx := ctxBackground()

	origen := crearProyectoDePrueba(t, proyectos, "Origen")
	destino := crearProyectoDePrueba(t, proyectos, "Destino")

	sprint := crearSprintDePrueba(t, sprints, origen.ID, "Sprint 1",
		sprintFecha(time.October, 6), sprintFecha(time.October, 20), domain.EstadoSprintPlanificado)

	_, err := sprints.ObtenerPorID(ctx, destino.ID, sprint.ID)
	assert.ErrorIs(t, err, domain.ErrSprintNoEncontrado, "escenario BDD E-35")

	lista, err := sprints.ListarPorProyecto(ctx, destino.ID)
	require.NoError(t, err)
	assert.Empty(t, lista)
}

func TestSprintListarSinRegistrosDevuelveVacio(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	sprints := NuevoSprintRepository(db)

	proyecto := crearProyectoDePrueba(t, proyectos, "Sin sprints")

	lista, err := sprints.ListarPorProyecto(ctxBackground(), proyecto.ID)
	require.NoError(t, err)
	assert.NotNil(t, lista, "el listado vacio no debe ser nil")
	assert.Empty(t, lista, "escenario BDD E-26")
}

func TestSprintHaySolapamiento(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	sprints := NuevoSprintRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Con solapamiento")
	crearSprintDePrueba(t, sprints, proyecto.ID, "Sprint 1",
		sprintFecha(time.October, 6), sprintFecha(time.October, 20), domain.EstadoSprintPlanificado)

	casos := []struct {
		nombre       string
		inicio       time.Time
		fin          time.Time
		solapamiento bool
		escenario    string
	}{
		{"mismo rango", sprintFecha(time.October, 6), sprintFecha(time.October, 20), true, "E-14"},
		{"solapamiento parcial por la derecha", sprintFecha(time.October, 19), sprintFecha(time.November, 3), true, "E-14"},
		{"solapamiento parcial por la izquierda", sprintFecha(time.October, 3), sprintFecha(time.October, 7), true, "E-14"},
		{"contenido completamente", sprintFecha(time.October, 10), sprintFecha(time.October, 12), true, "E-15"},
		{"contiguo despues", sprintFecha(time.October, 20), sprintFecha(time.November, 3), false, "E-16"},
		{"contiguo antes", sprintFecha(time.October, 3), sprintFecha(time.October, 6), false, "E-16"},
		{"distante", sprintFecha(time.October, 21), sprintFecha(time.October, 28), false, "E-13"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			solapa, err := sprints.HaySolapamiento(ctx, proyecto.ID, c.inicio, c.fin)
			require.NoError(t, err)
			assert.Equal(t, c.solapamiento, solapa, "escenario BDD %s", c.escenario)
		})
	}
}

func TestSprintActualizarEstadoInexistenteDevuelveError(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	sprints := NuevoSprintRepository(db)

	proyecto := crearProyectoDePrueba(t, proyectos, "Vacio")

	err := sprints.ActualizarEstado(ctxBackground(), &domain.Sprint{
		ID: 999, ProyectoID: proyecto.ID, Estado: domain.EstadoSprintActivo,
	})
	assert.ErrorIs(t, err, domain.ErrSprintNoEncontrado, "escenario BDD E-34")
}
