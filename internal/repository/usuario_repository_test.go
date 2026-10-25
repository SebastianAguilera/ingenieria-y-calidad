package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func TestVincularUsuarioConRolEsPersistido(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	usuarios := NuevoUsuarioRepository(db)
	ctx := ctxBackground()

	inicio := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)
	proyecto := &domain.Proyecto{Nombre: "Proyecto con roles", FechaInicio: inicio, Estado: domain.EstadoActivo}
	require.NoError(t, proyectos.CrearConUsuarios(ctx, proyecto, nil))

	usuario := &domain.Usuario{Nombre: "Ada Lovelace", Email: "ada@utn.edu.ar"}
	require.NoError(t, usuarios.Crear(ctx, usuario))

	require.NoError(t, usuarios.Vincular(ctx, proyecto.ID, usuario.ID, domain.RolArquitectoDeProducto))

	var rolGuardado string
	err := db.WithContext(ctx).
		Table("proyecto_usuarios").
		Select("rol").
		Where("proyecto_id = ? AND usuario_id = ?", proyecto.ID, usuario.ID).
		Scan(&rolGuardado).Error
	require.NoError(t, err)
	assert.Equal(t, string(domain.RolArquitectoDeProducto), rolGuardado)
}
