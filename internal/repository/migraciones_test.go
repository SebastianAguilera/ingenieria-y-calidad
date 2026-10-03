package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"ingenieria-y-calidad/internal/domain"
)

// abrirBaseDePruebas abre una conexion real y limpia las tablas de US-01 para
// dejar la base en un estado conocido. Se omite cuando la base no esta
// disponible, de modo que go test ./... funcione sin infraestructura.
func abrirBaseDePruebas(t *testing.T) *gorm.DB {
	t.Helper()

	if testing.Short() {
		t.Skip("test de integracion omitido en modo -short")
	}
	if os.Getenv("DB_HOST") == "" {
		t.Skip("DB_HOST no definido: base de datos no disponible para el test de integracion")
	}

	db, err := ConnectDB()
	if err != nil {
		t.Skipf("base de datos no disponible: %v", err)
	}

	require.NoError(t, MigrarEsquema(db))
	limpiarTablas(t, db)
	t.Cleanup(func() { limpiarTablas(t, db) })

	return db
}

func limpiarTablas(t *testing.T, db *gorm.DB) {
	t.Helper()
	for _, tabla := range []string{"proyecto_integrantes", "usuarios", "proyectos"} {
		if err := db.Exec("DROP TABLE IF EXISTS " + tabla + " CASCADE").Error; err != nil {
			t.Logf("no se pudo eliminar la tabla %s: %v", tabla, err)
		}
	}
	require.NoError(t, MigrarEsquema(db))
}

func TestMigrarEsquemaEsIdempotente(t *testing.T) {
	db := abrirBaseDePruebas(t)

	require.NotPanics(t, func() {
		require.NoError(t, MigrarEsquema(db))
	})

	for _, tabla := range []string{"proyectos", "usuarios", "proyecto_integrantes"} {
		assert.True(t, db.Migrator().HasTable(tabla), "falta la tabla %s", tabla)
	}
}

func TestIndiceUnicoDeEmail(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoUsuarioRepository(db)

	require.NoError(t, repo.Crear(ctxBackground(), &domain.Usuario{Nombre: "Ada", Email: "ada@utn.edu.ar"}))
	err := repo.Crear(ctxBackground(), &domain.Usuario{Nombre: "ada", Email: "ada@utn.edu.ar"})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUsuarioYaAsociado)
}

func TestBajaLogicaInvisibleParaElCliente(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	ctx := ctxBackground()

	inicio := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)
	creado := &domain.Proyecto{Nombre: "Proyecto", FechaInicio: inicio, Estado: domain.EstadoActivo}
	require.NoError(t, proyectos.CrearConUsuarios(ctx, creado, nil))
	require.NotZero(t, creado.ID, "el id lo genera la base")

	encontrado, err := proyectos.ObtenerPorID(ctx, creado.ID)
	require.NoError(t, err)
	assert.Equal(t, "Proyecto", encontrado.Nombre)
	assert.Nil(t, encontrado.FechaFin)

	require.NoError(t, proyectos.Eliminar(ctx, creado.ID))

	_, err = proyectos.ObtenerPorID(ctx, creado.ID)
	assert.ErrorIs(t, err, domain.ErrProyectoNoEncontrado, "escenario BDD E-50")

	lista, err := proyectos.Listar(ctx)
	require.NoError(t, err)
	assert.Empty(t, lista, "un proyecto dado de baja no debe aparecer en el listado")

	assert.ErrorIs(t, proyectos.Eliminar(ctx, creado.ID), domain.ErrProyectoNoEncontrado)
}

func TestCrearConIntegrantesEsAtomico(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	usuarios := NuevoUsuarioRepository(db)
	ctx := ctxBackground()

	inicio := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)
	proyecto := &domain.Proyecto{Nombre: "Con equipo", FechaInicio: inicio, Estado: domain.EstadoActivo}
	equipo := []domain.Usuario{
		{Nombre: "Ada Lovelace", Email: "ada@utn.edu.ar"},
		{Nombre: "Alan Turing", Email: "alan@utn.edu.ar"},
	}

	require.NoError(t, proyectos.CrearConUsuarios(ctx, proyecto, equipo))
	assert.NotZero(t, proyecto.ID)
	assert.NotZero(t, equipo[0].ID, "el id de cada usuario lo genera la base")

	detalle, err := proyectos.ObtenerPorID(ctx, proyecto.ID)
	require.NoError(t, err)
	require.Len(t, detalle.Usuarios, 2)

	asociado, err := usuarios.EstaAsociado(ctx, proyecto.ID, equipo[0].ID)
	require.NoError(t, err)
	assert.True(t, asociado)

	email, err := usuarios.ObtenerPorEmail(ctx, "  ADA@UTN.edu.ar ")
	require.NoError(t, err)
	assert.Equal(t, equipo[0].ID, email.ID, "la busqueda normaliza el email")

	require.NoError(t, usuarios.Desvincular(ctx, proyecto.ID, equipo[0].ID))
	asociado, err = usuarios.EstaAsociado(ctx, proyecto.ID, equipo[0].ID)
	require.NoError(t, err)
	assert.False(t, asociado)
}

func TestActualizarProyecto(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	ctx := ctxBackground()

	inicio := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)
	fin := time.Date(2026, 12, 18, 0, 0, 0, 0, time.UTC)
	proyecto := &domain.Proyecto{Nombre: "Original", FechaInicio: inicio, Estado: domain.EstadoActivo}
	require.NoError(t, proyectos.CrearConUsuarios(ctx, proyecto, nil))

	proyecto.Nombre = "Modificado"
	proyecto.FechaFin = &fin
	proyecto.Estado = domain.EstadoCerrado
	require.NoError(t, proyectos.Actualizar(ctx, proyecto))

	detalle, err := proyectos.ObtenerPorID(ctx, proyecto.ID)
	require.NoError(t, err)
	assert.Equal(t, "Modificado", detalle.Nombre)
	require.NotNil(t, detalle.FechaFin)
	assert.Equal(t, fin, *detalle.FechaFin)
	assert.Equal(t, domain.EstadoCerrado, detalle.Estado)
}

// TestLaAltaConIntegrantesRevierteSiFallaUnIntegrante es el unico test que
// prueba de verdad la atomicidad de CL-18 y del criterio correspondiente. El
// resto de la suite solo verifica que la transaccion este declarada, nunca que
// revierta. La falla se provoca con un email duplicado: el indice unico de
// usuarios.rechaza el segundo INSERT y la transaccion debe deshacer tanto el
// proyecto como el primer usuario.
func TestLaAltaConIntegrantesRevierteSiFallaUnIntegrante(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	usuarios := NuevoUsuarioRepository(db)
	ctx := ctxBackground()

	inicio := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)

	proyectosAntes, err := proyectos.Listar(ctx)
	require.NoError(t, err)

	// El equipo trae dos usuarios con el mismo email normalizado. El primero
	// se persiste y el segundo viola el indice unico, lo que aborta la
	// transaccion completa.
	equipo := []domain.Usuario{
		{Nombre: "Primero", Email: "primero@utn.edu.ar"},
		{Nombre: "Segundo", Email: "PRIMERO@utn.edu.ar"},
	}
	proyecto := &domain.Proyecto{
		Nombre: "Debe revertirse", FechaInicio: inicio, Estado: domain.EstadoActivo,
	}

	err = proyectos.CrearConUsuarios(ctx, proyecto, equipo)
	require.Error(t, err, "el alta debe fallar por el email duplicado")
	assert.ErrorIs(t, err, domain.ErrUsuarioYaAsociado)

	// Nada de lo que escribio la transaccion debe haber sobrevivido.
	proyectosDespues, err := proyectos.Listar(ctx)
	require.NoError(t, err)
	assert.Len(t, proyectosDespues, len(proyectosAntes),
		"el proyecto no debe haber quedado persistido")

	_, err = proyectos.ObtenerPorID(ctx, proyecto.ID)
	assert.ErrorIs(t, err, domain.ErrProyectoNoEncontrado,
		"el proyecto revertido no debe ser consultable por id")

	// El primer usuario llego a insertarse antes de la falla: si la
	// transaccion no revirtiera, aqui quedaria un huerfano.
	_, err = usuarios.ObtenerPorEmail(ctx, "primero@utn.edu.ar")
	assert.ErrorIs(t, err, domain.ErrUsuarioNoEncontrado,
		"el primer usuario de la transaccion tampoco debe quedar")
}

// TestMarcasDeAuditoriaSePersisten protege la exposicion de las marcas de tiempo:
// el smoke test end-to-end devolvio "actualizado_en": "0001-01-01T00:00:00Z"
// porque la columna no tenia valor por defecto y el repositorio no la escribia.
func TestMarcasDeAuditoriaSePersisten(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	ctx := ctxBackground()

	inicio := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)
	proyecto := &domain.Proyecto{Nombre: "Auditado", FechaInicio: inicio, Estado: domain.EstadoActivo}
	require.NoError(t, proyectos.CrearConUsuarios(ctx, proyecto, nil))

	assert.False(t, proyecto.CreadoEn.IsZero(), "creado_en debe informarse al crear")
	assert.False(t, proyecto.ActualizadoEn.IsZero(), "actualizado_en debe informarse al crear")

	creadoEn := proyecto.CreadoEn
	actualizadoEn := proyecto.ActualizadoEn
	detalle, err := proyectos.ObtenerPorID(ctx, proyecto.ID)
	require.NoError(t, err)
	assert.False(t, detalle.CreadoEn.IsZero())
	assert.False(t, detalle.ActualizadoEn.IsZero(),
		"la base no debe devolver el valor cero del tiempo")

	// Una actualizacion debe refrescar actualizado_en sin perder creado_en.
	time.Sleep(10 * time.Millisecond)
	proyecto.Nombre = "Auditado v2"
	require.NoError(t, proyectos.Actualizar(ctx, proyecto))

	detalle, err = proyectos.ObtenerPorID(ctx, proyecto.ID)
	require.NoError(t, err)
	assert.WithinDuration(t, creadoEn, detalle.CreadoEn, time.Second,
		"creado_en es inmutable")
	assert.False(t, detalle.ActualizadoEn.Before(actualizadoEn),
		"actualizado_en no debe retroceder")
}

func ctxBackground() context.Context {
	return context.Background()
}
