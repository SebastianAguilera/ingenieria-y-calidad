package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func TestSprintBacklogRepositorioObtenerPorSprint(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	sprints := NuevoSprintRepository(db)
	historias := NuevoHistoriaRepository(db)
	backlog := NuevoSprintBacklogRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Con sprint backlog")
	sprint := crearSprintDePrueba(t, sprints, proyecto.ID, "Sprint 1",
		sprintFecha(time.October, 6), sprintFecha(time.October, 20), domain.EstadoSprintActivo)
	historiaUno := crearHistoriaConTitulo(t, historias, proyecto.ID, 1, "Primera")
	historiaDos := crearHistoriaConTitulo(t, historias, proyecto.ID, 2, "Segunda")
	historiaTres := crearHistoriaConTitulo(t, historias, proyecto.ID, 3, "Tercera")

	asignar := func(historiaID uint, orden int) *domain.SprintBacklogItem {
		item := &domain.SprintBacklogItem{
			ProyectoID: proyecto.ID,
			SprintID:   sprint.ID,
			HistoriaID: historiaID,
			Orden:      orden,
		}
		require.NoError(t, backlog.Asignar(ctx, item))
		return item
	}

	primera := asignar(historiaUno.ID, 2)
	segunda := asignar(historiaDos.ID, 1)
	tercera := asignar(historiaTres.ID, 1)

	items, err := backlog.ObtenerPorSprint(ctx, proyecto.ID, sprint.ID)
	require.NoError(t, err)
	require.Len(t, items, 3)
	assert.Equal(t, []uint{segunda.ID, tercera.ID, primera.ID},
		[]uint{items[0].ID, items[1].ID, items[2].ID},
		"ordena por orden ASC y desempata por id ASC")

	vacio := crearSprintDePrueba(t, sprints, proyecto.ID, "Sprint 2",
		sprintFecha(time.October, 21), sprintFecha(time.November, 3), domain.EstadoSprintPlanificado)
	sinItems, err := backlog.ObtenerPorSprint(ctx, proyecto.ID, vacio.ID)
	require.NoError(t, err)
	assert.NotNil(t, sinItems)
	assert.Empty(t, sinItems)
}
