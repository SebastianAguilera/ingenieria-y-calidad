package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func TestAcceptanceCriterionRepositorioCicloCompleto(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	historias := NuevoHistoriaRepository(db)
	criterios := NuevoAcceptanceCriterionRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Con criterios")
	historia := crearHistoriaDePrueba(t, historias, proyecto.ID, 1)

	posicion, err := criterios.SiguientePosicion(ctx, historia.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, posicion, "escenario BDD BDD-01: el primer criterio usa posición 1")

	primero := &domain.AcceptanceCriterion{HistoriaID: historia.ID, Descripcion: "Primero", Position: 1}
	require.NoError(t, criterios.Crear(ctx, primero))
	assert.NotZero(t, primero.ID)
	assert.False(t, primero.IsMet)

	posicion, err = criterios.SiguientePosicion(ctx, historia.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, posicion, "escenario BDD BDD-02: el segundo criterio usa posición 2")

	segundo := &domain.AcceptanceCriterion{HistoriaID: historia.ID, Descripcion: "Segundo", Position: 2, IsMet: true}
	require.NoError(t, criterios.Crear(ctx, segundo))

	listado, err := criterios.ListarPorHistoria(ctx, historia.ID)
	require.NoError(t, err)
	require.Len(t, listado, 2)
	assert.Equal(t, "Primero", listado[0].Descripcion)
	assert.Equal(t, "Segundo", listado[1].Descripcion, "escenario BDD BDD-03: orden por posición")

	recuperado, err := criterios.ObtenerPorID(ctx, historia.ID, primero.ID)
	require.NoError(t, err)
	assert.Equal(t, "Primero", recuperado.Descripcion)

	_, err = criterios.ObtenerPorID(ctx, historia.ID, 999999)
	assert.ErrorIs(t, err, domain.ErrCriterioNoEncontrado, "escenario BDD BDD-11")

	recuperado.Descripcion = "Primero actualizado"
	recuperado.IsMet = true
	require.NoError(t, criterios.Actualizar(ctx, recuperado))
	assert.True(t, recuperado.IsMet, "escenario BDD BDD-05")

	actualizado, err := criterios.ObtenerPorID(ctx, historia.ID, primero.ID)
	require.NoError(t, err)
	assert.Equal(t, "Primero actualizado", actualizado.Descripcion)
	assert.True(t, actualizado.IsMet)

	require.NoError(t, criterios.Eliminar(ctx, historia.ID, primero.ID), "escenario BDD BDD-07")
	_, err = criterios.ObtenerPorID(ctx, historia.ID, primero.ID)
	assert.ErrorIs(t, err, domain.ErrCriterioNoEncontrado)

	err = criterios.Eliminar(ctx, historia.ID, primero.ID)
	assert.ErrorIs(t, err, domain.ErrCriterioNoEncontrado)
}
