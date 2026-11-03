package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func TestObtenerUsuarioPorID(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoUsuarioRepository(db)
	ctx := ctxBackground()

	usuario := &domain.Usuario{Nombre: "Ada Lovelace", Email: "ada@utn.edu.ar"}
	require.NoError(t, repo.Crear(ctx, usuario))
	require.NotZero(t, usuario.ID)

	recuperado, err := repo.ObtenerPorID(ctx, usuario.ID)
	require.NoError(t, err)
	require.NotNil(t, recuperado)
	assert.Equal(t, usuario.ID, recuperado.ID)
	assert.Equal(t, "Ada Lovelace", recuperado.Nombre)
	assert.Equal(t, "ada@utn.edu.ar", recuperado.Email)

	_, err = repo.ObtenerPorID(ctx, 999999)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUsuarioNoEncontrado)
}

func TestListarUsuarios(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoUsuarioRepository(db)
	ctx := ctxBackground()

	require.NoError(t, repo.Crear(ctx, &domain.Usuario{Nombre: "Ada", Email: "ada@utn.edu.ar"}))
	require.NoError(t, repo.Crear(ctx, &domain.Usuario{Nombre: "Alan", Email: "alan@utn.edu.ar"}))

	lista, err := repo.Listar(ctx)
	require.NoError(t, err)
	assert.Len(t, lista, 2)
}

func TestActualizarNombreUsuario(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoUsuarioRepository(db)
	ctx := ctxBackground()

	usuario := &domain.Usuario{Nombre: "Ada", Email: "ada@utn.edu.ar"}
	require.NoError(t, repo.Crear(ctx, usuario))

	require.NoError(t, repo.ActualizarNombre(ctx, usuario.ID, "  Ada Lovelace  "))

	recuperado, err := repo.ObtenerPorID(ctx, usuario.ID)
	require.NoError(t, err)
	assert.Equal(t, "Ada Lovelace", recuperado.Nombre)
}

func TestListarUsuariosPorProyecto(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	usuarios := NuevoUsuarioRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Con equipo")
	otroProyecto := crearProyectoDePrueba(t, proyectos, "Con otro equipo")

	ada := &domain.Usuario{Nombre: "Ada", Email: "ada@utn.edu.ar"}
	require.NoError(t, usuarios.Crear(ctx, ada))
	alan := &domain.Usuario{Nombre: "Alan", Email: "alan@utn.edu.ar"}
	require.NoError(t, usuarios.Crear(ctx, alan))
	grace := &domain.Usuario{Nombre: "Grace", Email: "grace@utn.edu.ar"}
	require.NoError(t, usuarios.Crear(ctx, grace))

	require.NoError(t, usuarios.Vincular(ctx, proyecto.ID, alan.ID))
	require.NoError(t, usuarios.Vincular(ctx, proyecto.ID, ada.ID))
	require.NoError(t, usuarios.Vincular(ctx, otroProyecto.ID, grace.ID))

	equipo, err := usuarios.ListarPorProyecto(ctx, proyecto.ID)
	require.NoError(t, err)
	require.Len(t, equipo, 2)
	assert.Equal(t, ada.ID, equipo[0].ID, "el listado va ordenado por id")
	assert.Equal(t, alan.ID, equipo[1].ID)
	assert.Equal(t, "ada@utn.edu.ar", equipo[0].Email)

	vacio, err := usuarios.ListarPorProyecto(ctx, 999999)
	require.NoError(t, err)
	assert.NotNil(t, vacio, "el listado vacio no debe ser nil")
	assert.Empty(t, vacio)
}
