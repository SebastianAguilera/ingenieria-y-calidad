package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

// TestObtenerProyectoIncluyeElCreador cubre US-F25: el proyecto persiste quien
// lo dio de alta y el detalle devuelve ese usuario, aunque despues cambie el LIDER.
func TestObtenerProyectoIncluyeElCreador(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	usuarios := NuevoUsuarioRepository(db)
	ctx := ctxBackground()

	ana := &domain.Usuario{Nombre: "Ana Ruiz", Email: "ana@utn.edu.ar", Rol: domain.RolLider}
	require.NoError(t, usuarios.Crear(ctx, ana))

	proyecto := &domain.Proyecto{
		Nombre:      "Con creador",
		FechaInicio: time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC),
		Estado:      domain.EstadoActivo,
		CreadoPorID: ana.ID,
	}
	require.NoError(t, proyectos.CrearConUsuarios(ctx, proyecto, []domain.Usuario{*ana}))

	obtenido, err := proyectos.ObtenerPorID(ctx, proyecto.ID)
	require.NoError(t, err)
	assert.Equal(t, ana.ID, obtenido.CreadoPorID)
	require.NotNil(t, obtenido.CreadoPor)
	assert.Equal(t, "Ana Ruiz", obtenido.CreadoPor.Nombre)
}

// TestMigrarCreadorDeProyectoCompletaLosAnteriores verifica el backfill: un
// proyecto previo sin creador recibe como creador al LIDER de menor id, y un
// proyecto que ya tenia creador no cambia. Es idempotente.
func TestMigrarCreadorDeProyectoCompletaLosAnteriores(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	usuarios := NuevoUsuarioRepository(db)
	ctx := ctxBackground()

	lider := &domain.Usuario{Nombre: "Lider", Email: "lider@utn.edu.ar"}
	otro := &domain.Usuario{Nombre: "Otro", Email: "otro@utn.edu.ar"}
	require.NoError(t, usuarios.Crear(ctx, lider))
	require.NoError(t, usuarios.Crear(ctx, otro))

	anterior := crearProyectoDePrueba(t, proyectos, "Anterior")
	require.NoError(t, usuarios.Vincular(ctx, anterior.ID, lider.ID, domain.RolLider))
	require.NoError(t, usuarios.Vincular(ctx, anterior.ID, otro.ID))

	conCreador := crearProyectoDePrueba(t, proyectos, "Con creador")
	require.NoError(t, db.Exec(`UPDATE proyectos SET creado_por_id = ? WHERE id = ?`, otro.ID, conCreador.ID).Error)

	require.NoError(t, migrarCreadorDeProyecto(db))
	require.NoError(t, migrarCreadorDeProyecto(db), "debe ser idempotente")

	obtenido, err := proyectos.ObtenerPorID(ctx, anterior.ID)
	require.NoError(t, err)
	assert.Equal(t, lider.ID, obtenido.CreadoPorID)

	conservado, err := proyectos.ObtenerPorID(ctx, conCreador.ID)
	require.NoError(t, err)
	assert.Equal(t, otro.ID, conservado.CreadoPorID, "un creador ya registrado no se reemplaza")
}
