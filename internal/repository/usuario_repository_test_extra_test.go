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
