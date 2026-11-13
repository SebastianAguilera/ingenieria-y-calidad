package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"ingenieria-y-calidad/internal/domain"
)

var inicioRoles = time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)

func proyectoParaRoles(t *testing.T, db *gorm.DB, nombre string) *domain.Proyecto {
	t.Helper()
	proyecto := &domain.Proyecto{Nombre: nombre, FechaInicio: inicioRoles, Estado: domain.EstadoActivo}
	require.NoError(t, NuevoProyectoRepository(db).CrearConUsuarios(ctxBackground(), proyecto, nil))
	return proyecto
}

func usuarioParaRoles(t *testing.T, db *gorm.DB, email string) *domain.Usuario {
	t.Helper()
	usuario := &domain.Usuario{Nombre: "Usuario " + email, Email: email}
	require.NoError(t, NuevoUsuarioRepository(db).Crear(ctxBackground(), usuario))
	return usuario
}

func TestVincularSinRolEntraComoMiembro(t *testing.T) {
	db := abrirBaseDePruebas(t)
	usuarios := NuevoUsuarioRepository(db)
	ctx := ctxBackground()
	proyecto := proyectoParaRoles(t, db, "Sin rol")
	ada := usuarioParaRoles(t, db, "ada@utn.edu.ar")

	require.NoError(t, usuarios.Vincular(ctx, proyecto.ID, ada.ID))

	rol, err := usuarios.ObtenerRol(ctx, proyecto.ID, ada.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.RolMiembro, rol)
}

func TestCrearConUsuariosPersisteElRolDelCreador(t *testing.T) {
	db := abrirBaseDePruebas(t)
	usuarios := NuevoUsuarioRepository(db)
	ctx := ctxBackground()
	ana := usuarioParaRoles(t, db, "ana@utn.edu.ar")
	ben := usuarioParaRoles(t, db, "ben@utn.edu.ar")
	ana.Rol = domain.RolLider

	proyecto := &domain.Proyecto{Nombre: "Con equipo", FechaInicio: inicioRoles, Estado: domain.EstadoActivo}
	equipo := []domain.Usuario{*ana, *ben}
	require.NoError(t, NuevoProyectoRepository(db).CrearConUsuarios(ctx, proyecto, equipo))

	rolAna, err := usuarios.ObtenerRol(ctx, proyecto.ID, ana.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.RolLider, rolAna, "RB-01: el creador queda como LIDER")

	rolBen, err := usuarios.ObtenerRol(ctx, proyecto.ID, ben.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.RolMiembro, rolBen, "sin rol informado entra como MIEMBRO")
}

func TestObtenerRolDeNoIntegrante(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyecto := proyectoParaRoles(t, db, "Ajeno")
	cris := usuarioParaRoles(t, db, "cris@utn.edu.ar")

	_, err := NuevoUsuarioRepository(db).ObtenerRol(ctxBackground(), proyecto.ID, cris.ID)

	assert.ErrorIs(t, err, domain.ErrUsuarioNoAsociado)
}

func TestContarLideresYActualizarRol(t *testing.T) {
	db := abrirBaseDePruebas(t)
	usuarios := NuevoUsuarioRepository(db)
	ctx := ctxBackground()
	proyecto := proyectoParaRoles(t, db, "Lideres")
	ana := usuarioParaRoles(t, db, "ana@utn.edu.ar")
	ben := usuarioParaRoles(t, db, "ben@utn.edu.ar")
	require.NoError(t, usuarios.Vincular(ctx, proyecto.ID, ana.ID, domain.RolLider))
	require.NoError(t, usuarios.Vincular(ctx, proyecto.ID, ben.ID, domain.RolMiembro))

	total, err := usuarios.ContarLideres(ctx, proyecto.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, total)

	require.NoError(t, usuarios.ActualizarRol(ctx, proyecto.ID, ben.ID, domain.RolLider))
	total, err = usuarios.ContarLideres(ctx, proyecto.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, total)

	rol, err := usuarios.ObtenerRol(ctx, proyecto.ID, ben.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.RolLider, rol)
}

func TestActualizarRolDeNoIntegrante(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyecto := proyectoParaRoles(t, db, "Ajeno")
	cris := usuarioParaRoles(t, db, "cris@utn.edu.ar")

	err := NuevoUsuarioRepository(db).ActualizarRol(ctxBackground(), proyecto.ID, cris.ID, domain.RolLider)

	assert.ErrorIs(t, err, domain.ErrUsuarioNoAsociado)
}

// TestListarMiembrosPorProyectoIncluyeRol cubre CA-1 y E-01/E-06: el listado de
// integrantes devuelve id, nombre, email y rol en orden de alta.
func TestListarMiembrosPorProyectoIncluyeRol(t *testing.T) {
	db := abrirBaseDePruebas(t)
	usuarios := NuevoUsuarioRepository(db)
	ctx := ctxBackground()
	proyecto := proyectoParaRoles(t, db, "Equipo")
	ana := usuarioParaRoles(t, db, "ana@utn.edu.ar")
	ben := usuarioParaRoles(t, db, "ben@utn.edu.ar")
	require.NoError(t, usuarios.Vincular(ctx, proyecto.ID, ana.ID, domain.RolLider))
	require.NoError(t, usuarios.Vincular(ctx, proyecto.ID, ben.ID, domain.RolMiembro))

	miembros, err := usuarios.ListarMiembrosPorProyecto(ctx, proyecto.ID)

	require.NoError(t, err)
	require.Len(t, miembros, 2)
	assert.Equal(t, domain.MiembroProyecto{
		ID: ana.ID, Nombre: ana.Nombre, Email: ana.Email, Rol: domain.RolLider,
	}, miembros[0])
	assert.Equal(t, domain.MiembroProyecto{
		ID: ben.ID, Nombre: ben.Nombre, Email: ben.Email, Rol: domain.RolMiembro,
	}, miembros[1])
}

func TestListarMiembrosPorProyectoVacioYAislado(t *testing.T) {
	db := abrirBaseDePruebas(t)
	usuarios := NuevoUsuarioRepository(db)
	ctx := ctxBackground()
	proyecto := proyectoParaRoles(t, db, "Vacio")
	otro := proyectoParaRoles(t, db, "Otro")
	ada := usuarioParaRoles(t, db, "ada@utn.edu.ar")
	require.NoError(t, usuarios.Vincular(ctx, otro.ID, ada.ID, domain.RolLider))

	miembros, err := usuarios.ListarMiembrosPorProyecto(ctx, proyecto.ID)

	require.NoError(t, err)
	assert.Empty(t, miembros, "un proyecto sin equipo devuelve lista vacia (E-09)")
}

// TestMigracionDeRolesDejaUnLiderPorProyecto cubre E-15, E-16 y CA-10.
func TestMigracionDeRolesDejaUnLiderPorProyecto(t *testing.T) {
	db := abrirBaseDePruebas(t)
	ctx := ctxBackground()
	usuarios := NuevoUsuarioRepository(db)

	ana := usuarioParaRoles(t, db, "ana@utn.edu.ar")
	ben := usuarioParaRoles(t, db, "ben@utn.edu.ar")
	cris := usuarioParaRoles(t, db, "cris@utn.edu.ar")

	// Proyecto 1: roles viejos de US-08 y sin lider.
	sinLider := proyectoParaRoles(t, db, "Sin lider")
	for _, v := range []struct {
		u   *domain.Usuario
		rol string
	}{{ben, "Arquitecto de Producto"}, {cris, "Facilitador Ágil"}, {ana, "Constructor de Producto"}} {
		require.NoError(t, db.Exec(
			"INSERT INTO proyecto_usuarios (proyecto_id, usuario_id, rol) VALUES (?, ?, ?)",
			sinLider.ID, v.u.ID, v.rol).Error)
	}

	// Proyecto 2: ya tiene un lider (cris) y no debe cambiar.
	conLider := proyectoParaRoles(t, db, "Con lider")
	require.NoError(t, usuarios.Vincular(ctx, conLider.ID, ana.ID, domain.RolMiembro))
	require.NoError(t, usuarios.Vincular(ctx, conLider.ID, cris.ID, domain.RolLider))

	require.NoError(t, MigrarEsquema(db))

	rol := func(proyectoID, usuarioID uint) domain.RolProyecto {
		r, err := usuarios.ObtenerRol(ctx, proyectoID, usuarioID)
		require.NoError(t, err)
		return r
	}
	assert.Equal(t, domain.RolLider, rol(sinLider.ID, ana.ID), "el integrante de menor id pasa a LIDER")
	assert.Equal(t, domain.RolMiembro, rol(sinLider.ID, ben.ID), "los roles viejos pasan a MIEMBRO")
	assert.Equal(t, domain.RolMiembro, rol(sinLider.ID, cris.ID))
	assert.Equal(t, domain.RolMiembro, rol(conLider.ID, ana.ID), "un proyecto con lider no se toca")
	assert.Equal(t, domain.RolLider, rol(conLider.ID, cris.ID))

	for _, id := range []uint{sinLider.ID, conLider.ID} {
		total, err := usuarios.ContarLideres(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, 1, total, "CA-10: ningun proyecto con equipo queda sin lider")
	}

	// E-16: una segunda corrida no cambia nada.
	require.NoError(t, MigrarEsquema(db))
	assert.Equal(t, domain.RolLider, rol(sinLider.ID, ana.ID))
	assert.Equal(t, domain.RolMiembro, rol(sinLider.ID, ben.ID))
}
