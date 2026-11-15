package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

// TestListarPorUsuarioFiltraPorMembresia cubre RB-01 de US-F10: cada persona
// solo ve los proyectos en los que tiene vínculo; los proyectos sin equipo y
// los de los demás no aparecen, y un usuario sin proyectos obtiene la lista
// vacía.
func TestListarPorUsuarioFiltraPorMembresia(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	usuarios := NuevoUsuarioRepository(db)
	ctx := ctxBackground()

	ana := &domain.Usuario{Nombre: "Ana", Email: "ana@utn.edu.ar"}
	ben := &domain.Usuario{Nombre: "Ben", Email: "ben@utn.edu.ar"}
	require.NoError(t, usuarios.Crear(ctx, ana))
	require.NoError(t, usuarios.Crear(ctx, ben))

	proyectoAna := crearProyectoDePrueba(t, proyectos, "De Ana")
	proyectoBen := crearProyectoDePrueba(t, proyectos, "De Ben")
	sinEquipo := crearProyectoDePrueba(t, proyectos, "Sin equipo")
	require.NoError(t, usuarios.Vincular(ctx, proyectoAna.ID, ana.ID))
	require.NoError(t, usuarios.Vincular(ctx, proyectoBen.ID, ben.ID))
	require.NotZero(t, sinEquipo.ID)

	deAna, err := proyectos.ListarPorUsuario(ctx, ana.ID)
	require.NoError(t, err)
	require.Len(t, deAna, 1, "ana solo ve su propio proyecto")
	assert.Equal(t, proyectoAna.ID, deAna[0].ID)

	deBen, err := proyectos.ListarPorUsuario(ctx, ben.ID)
	require.NoError(t, err)
	require.Len(t, deBen, 1, "ben solo ve su propio proyecto")
	assert.Equal(t, proyectoBen.ID, deBen[0].ID)

	sinCuentas, err := proyectos.ListarPorUsuario(ctx, 999999)
	require.NoError(t, err)
	assert.Empty(t, sinCuentas, "sin membresía no hay proyectos que listar")
}
