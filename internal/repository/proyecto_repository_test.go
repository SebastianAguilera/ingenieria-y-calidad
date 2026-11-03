package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func TestProyectoTieneHistorialSinDependencias(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Sin historial")

	tieneHistorial, err := proyectos.TieneHistorial(ctx, proyecto.ID)
	require.NoError(t, err)
	assert.False(t, tieneHistorial)

	tieneHistorial, err = proyectos.TieneHistorial(ctx, 999999)
	require.NoError(t, err)
	assert.False(t, tieneHistorial, "un proyecto inexistente no tiene historial")
}

func TestProyectoTieneHistorialConHistoria(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	historias := NuevoHistoriaRepository(db)
	ctx := ctxBackground()

	conHistoria := crearProyectoDePrueba(t, proyectos, "Con historia")
	sinHistoria := crearProyectoDePrueba(t, proyectos, "Sin historia")
	crearHistoriaDePrueba(t, historias, conHistoria.ID, 1)

	tieneHistorial, err := proyectos.TieneHistorial(ctx, conHistoria.ID)
	require.NoError(t, err)
	assert.True(t, tieneHistorial)

	tieneHistorial, err = proyectos.TieneHistorial(ctx, sinHistoria.ID)
	require.NoError(t, err)
	assert.False(t, tieneHistorial, "las historias de otro proyecto no cuentan")
}

func TestProyectoTieneHistorialConSprint(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	sprints := NuevoSprintRepository(db)
	ctx := ctxBackground()

	conSprint := crearProyectoDePrueba(t, proyectos, "Con sprint")
	sinSprint := crearProyectoDePrueba(t, proyectos, "Sin sprint")
	crearSprintDePrueba(t, sprints, conSprint.ID, "Sprint 1",
		sprintFecha(time.October, 6), sprintFecha(time.October, 20), domain.EstadoSprintPlanificado)

	tieneHistorial, err := proyectos.TieneHistorial(ctx, conSprint.ID)
	require.NoError(t, err)
	assert.True(t, tieneHistorial)

	tieneHistorial, err = proyectos.TieneHistorial(ctx, sinSprint.ID)
	require.NoError(t, err)
	assert.False(t, tieneHistorial, "los sprints de otro proyecto no cuentan")
}
