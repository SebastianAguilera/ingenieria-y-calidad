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

// Escenarios: docs/features/US-18-defectos-por-sprint.feature (E-01 a E-03, E-04 de repositorio).
// Helpers de este archivo para el listado por Sprint de US-18.

func crearDefectoParaSprint(t *testing.T, defectos domain.DefectoRepository, proyectoID, historiaID, sprintDeteccionID uint, descripcion string) *domain.Defecto {
	t.Helper()
	defecto := &domain.Defecto{
		ProyectoID:        proyectoID,
		HistoriaID:        historiaID,
		SprintDeteccionID: sprintDeteccionID,
		Descripcion:       descripcion,
		Severidad:         domain.SeveridadMinor,
		Estado:            domain.EstadoDefectoAbierto,
	}
	require.NoError(t, defectos.Crear(ctxBackground(), defecto))
	return defecto
}

func resolverDefectoEnSprint(t *testing.T, defectos domain.DefectoRepository, defecto *domain.Defecto, sprintResolucionID uint) {
	t.Helper()
	ahora := time.Now().UTC()
	defecto.Estado = domain.EstadoDefectoResuelto
	defecto.SprintResolucionID = &sprintResolucionID
	defecto.ResueltoEn = &ahora
	require.NoError(t, defectos.Actualizar(ctxBackground(), defecto))
}

func TestDefectoListarPorSprintDeteccion(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	historias := NuevoHistoriaRepository(db)
	sprints := NuevoSprintRepository(db)
	defectos := NuevoDefectoRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Listado por sprint")
	historia := crearHistoriaDePrueba(t, historias, proyecto.ID, 1)
	sprint1 := crearSprintDePrueba(t, sprints, proyecto.ID, "Sprint 1",
		sprintFecha(time.November, 1), sprintFecha(time.November, 15), domain.EstadoSprintActivo)
	sprint2 := crearSprintDePrueba(t, sprints, proyecto.ID, "Sprint 2",
		sprintFecha(time.November, 15), sprintFecha(time.November, 30), domain.EstadoSprintPlanificado)

	// Detectados en Sprint 1.
	d1 := crearDefectoParaSprint(t, defectos, proyecto.ID, historia.ID, sprint1.ID, "detectado en S1 (a)")
	d2 := crearDefectoParaSprint(t, defectos, proyecto.ID, historia.ID, sprint1.ID, "detectado en S1 (b)")
	// Detectado en Sprint 2, pero resuelto en Sprint 1: no debe aparecer por detección de Sprint 1.
	d3 := crearDefectoParaSprint(t, defectos, proyecto.ID, historia.ID, sprint2.ID, "detectado en S2, resuelto en S1")
	resolverDefectoEnSprint(t, defectos, d3, sprint1.ID)

	lista, err := defectos.ListarPorSprint(ctx, proyecto.ID, sprint1.ID, domain.TipoListadoDeteccion)

	require.NoError(t, err, "escenario BDD E-01")
	require.Len(t, lista, 2, "escenario BDD E-01: solo los detectados en el Sprint 1")
	assert.Equal(t, d1.ID, lista[0].ID, "escenario BDD E-01: orden por id ascendente")
	assert.Equal(t, d2.ID, lista[1].ID, "escenario BDD E-01: orden por id ascendente")
	for _, d := range lista {
		assert.Equal(t, sprint1.ID, d.SprintDeteccionID, "escenario BDD E-01")
	}
}

func TestDefectoListarPorSprintResolucion(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	historias := NuevoHistoriaRepository(db)
	sprints := NuevoSprintRepository(db)
	defectos := NuevoDefectoRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Resolucion por sprint")
	historia := crearHistoriaDePrueba(t, historias, proyecto.ID, 1)
	sprint1 := crearSprintDePrueba(t, sprints, proyecto.ID, "Sprint 1",
		sprintFecha(time.November, 1), sprintFecha(time.November, 15), domain.EstadoSprintActivo)
	sprint2 := crearSprintDePrueba(t, sprints, proyecto.ID, "Sprint 2",
		sprintFecha(time.November, 15), sprintFecha(time.November, 30), domain.EstadoSprintPlanificado)

	// Detectado en Sprint 1 y resuelto en Sprint 2.
	resueltoEnS2 := crearDefectoParaSprint(t, defectos, proyecto.ID, historia.ID, sprint1.ID, "resuelto en S2")
	resolverDefectoEnSprint(t, defectos, resueltoEnS2, sprint2.ID)
	// Sin resolver: no tiene sprint de resolución, no debe aparecer en ningún listado por resolución.
	crearDefectoParaSprint(t, defectos, proyecto.ID, historia.ID, sprint2.ID, "abierto")

	lista, err := defectos.ListarPorSprint(ctx, proyecto.ID, sprint2.ID, domain.TipoListadoResolucion)

	require.NoError(t, err, "escenario BDD E-02")
	require.Len(t, lista, 1, "escenario BDD E-02: solo los resueltos en el Sprint 2")
	assert.Equal(t, resueltoEnS2.ID, lista[0].ID, "escenario BDD E-02")
	require.NotNil(t, lista[0].SprintResolucionID)
	assert.Equal(t, sprint2.ID, *lista[0].SprintResolucionID, "escenario BDD E-02")
	assert.Equal(t, sprint1.ID, lista[0].SprintDeteccionID, "el sprint de detección no cambia al listar por resolución")

	// El mismo defecto no aparece al listar por resolución del Sprint 1.
	otro, err := defectos.ListarPorSprint(ctx, proyecto.ID, sprint1.ID, domain.TipoListadoResolucion)
	require.NoError(t, err)
	assert.Empty(t, otro, "escenario BDD E-02")
}

func TestDefectoListarPorSprintSinCoincidenciasDevuelveVacio(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	sprints := NuevoSprintRepository(db)
	defectos := NuevoDefectoRepository(db)

	proyecto := crearProyectoDePrueba(t, proyectos, "Sprint vacio")
	sprint := crearSprintDePrueba(t, sprints, proyecto.ID, "Sprint 1",
		sprintFecha(time.November, 1), sprintFecha(time.November, 15), domain.EstadoSprintActivo)

	lista, err := defectos.ListarPorSprint(ctxBackground(), proyecto.ID, sprint.ID, domain.TipoListadoResolucion)

	require.NoError(t, err, "escenario BDD E-03")
	assert.NotNil(t, lista, "escenario BDD E-03: el listado vacio nunca debe ser nil")
	assert.Empty(t, lista, "escenario BDD E-03")
}

func TestDefectoListarPorSprintTipoInvalidoDevuelveError(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	sprints := NuevoSprintRepository(db)
	defectos := NuevoDefectoRepository(db)

	proyecto := crearProyectoDePrueba(t, proyectos, "Tipo invalido")
	sprint := crearSprintDePrueba(t, sprints, proyecto.ID, "Sprint 1",
		sprintFecha(time.November, 1), sprintFecha(time.November, 15), domain.EstadoSprintActivo)

	for _, tipo := range []domain.TipoListadoDefecto{"", "Deteccion"} {
		_, err := defectos.ListarPorSprint(ctxBackground(), proyecto.ID, sprint.ID, tipo)
		assert.ErrorIs(t, err, domain.ErrTipoListadoDefectoInvalido, "escenario BDD E-04/E-05: tipo %q", tipo)
	}
}

func TestDefectoListarPorSprintNoMezclaProyectos(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	historias := NuevoHistoriaRepository(db)
	sprints := NuevoSprintRepository(db)
	defectos := NuevoDefectoRepository(db)

	origen := crearProyectoDePrueba(t, proyectos, "Origen listado")
	destino := crearProyectoDePrueba(t, proyectos, "Destino listado")
	historiaOrigen := crearHistoriaDePrueba(t, historias, origen.ID, 1)
	sprintOrigen := crearSprintDePrueba(t, sprints, origen.ID, "Sprint 1",
		sprintFecha(time.November, 1), sprintFecha(time.November, 15), domain.EstadoSprintActivo)
	crearDefectoParaSprint(t, defectos, origen.ID, historiaOrigen.ID, sprintOrigen.ID, "solo de origen")

	lista, err := defectos.ListarPorSprint(ctxBackground(), destino.ID, sprintOrigen.ID, domain.TipoListadoDeteccion)

	require.NoError(t, err, "escenario BDD E-07")
	assert.Empty(t, lista, "escenario BDD E-07: un defecto de otro proyecto no aparece")
}

// TestDefectoActualizarPersisteDescripcionYSeveridad cubre la persistencia de
// E-09 y E-10 (docs/features/US-18-defectos-por-sprint.feature) y verifica que
// no se toca el sprint de detección.
func TestDefectoActualizarPersisteDescripcionYSeveridad(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	historias := NuevoHistoriaRepository(db)
	sprints := NuevoSprintRepository(db)
	defectos := NuevoDefectoRepository(db)

	proyecto := crearProyectoDePrueba(t, proyectos, "Edicion de defectos")
	historia := crearHistoriaDePrueba(t, historias, proyecto.ID, 1)
	sprint := crearSprintDePrueba(t, sprints, proyecto.ID, "Sprint 1",
		sprintFecha(time.November, 1), sprintFecha(time.November, 15), domain.EstadoSprintActivo)
	defecto := crearDefectoParaSprint(t, defectos, proyecto.ID, historia.ID, sprint.ID, "Descripcion original")

	defecto.Descripcion = "Descripcion corregida"
	defecto.Severidad = domain.SeveridadBlocker
	require.NoError(t, defectos.Actualizar(ctxBackground(), defecto))

	actualizado, err := defectos.ObtenerPorID(ctxBackground(), proyecto.ID, defecto.ID)
	require.NoError(t, err)
	assert.Equal(t, "Descripcion corregida", actualizado.Descripcion, "escenario BDD E-09")
	assert.Equal(t, domain.SeveridadBlocker, actualizado.Severidad, "escenario BDD E-10")
	assert.Equal(t, sprint.ID, actualizado.SprintDeteccionID, "escenario BDD E-17: el sprint de detección no cambia")
}
