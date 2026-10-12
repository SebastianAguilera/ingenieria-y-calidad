package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func crearProyectoDePrueba(t *testing.T, repo *proyectoRepository, nombre string) *domain.Proyecto {
	t.Helper()
	inicio := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)
	proyecto := &domain.Proyecto{Nombre: nombre, FechaInicio: inicio, Estado: domain.EstadoActivo}
	require.NoError(t, repo.CrearConUsuarios(ctxBackground(), proyecto, nil))
	return proyecto
}

func TestHistoriaRepositorioCicloCompleto(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	historias := NuevoHistoriaRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Con backlog")

	numero, err := historias.SiguienteNumero(ctx, proyecto.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, numero, "la primera historia es la numero 1")

	primera := &domain.Historia{
		ProyectoID: proyecto.ID,
		Numero:     numero,
		Codigo:     domain.FormatearCodigoHistoria(numero),
		Titulo:     "Primera historia",
		Prioridad:  domain.PrioridadAlta,
		Estado:     domain.EstadoHistoriaBacklog,
	}
	require.NoError(t, historias.Crear(ctx, primera))
	assert.NotZero(t, primera.ID)
	assert.False(t, primera.CreadoEn.IsZero())

	numero, err = historias.SiguienteNumero(ctx, proyecto.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, numero)

	segunda := &domain.Historia{
		ProyectoID: proyecto.ID,
		Numero:     numero,
		Codigo:     domain.FormatearCodigoHistoria(numero),
		Titulo:     "Segunda historia",
		Prioridad:  domain.PrioridadMedia,
		Estado:     domain.EstadoHistoriaBacklog,
	}
	require.NoError(t, historias.Crear(ctx, segunda))
	assert.Equal(t, "US-02", segunda.Codigo)

	lista, err := historias.ListarPorProyecto(ctx, proyecto.ID)
	require.NoError(t, err)
	require.Len(t, lista, 2)
	assert.Equal(t, "US-01", lista[0].Codigo, "el listado va en orden de alta")

	recuperada, err := historias.ObtenerPorID(ctx, proyecto.ID, primera.ID)
	require.NoError(t, err)
	assert.Equal(t, "Primera historia", recuperada.Titulo)

	recuperada.Titulo = "Primera historia modificada"
	recuperada.Prioridad = domain.PrioridadBaja
	recuperada.Estado = domain.EstadoHistoriaHecho
	require.NoError(t, historias.Actualizar(ctx, recuperada))

	actualizada, err := historias.ObtenerPorID(ctx, proyecto.ID, primera.ID)
	require.NoError(t, err)
	assert.Equal(t, "Primera historia modificada", actualizada.Titulo)
	assert.Equal(t, domain.PrioridadBaja, actualizada.Prioridad)
	assert.Equal(t, domain.EstadoHistoriaHecho, actualizada.Estado)
	assert.False(t, actualizada.ActualizadoEn.Before(actualizada.CreadoEn))
}

func TestHistoriaCodigoDuplicadoEnElMismoProyecto(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	historias := NuevoHistoriaRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Con colision")

	require.NoError(t, historias.Crear(ctx, &domain.Historia{
		ProyectoID: proyecto.ID, Numero: 1, Codigo: "US-01", Titulo: "Original",
		Prioridad: domain.PrioridadMedia, Estado: domain.EstadoHistoriaBacklog,
	}))

	err := historias.Crear(ctx, &domain.Historia{
		ProyectoID: proyecto.ID, Numero: 1, Codigo: "US-01", Titulo: "Duplicada",
		Prioridad: domain.PrioridadMedia, Estado: domain.EstadoHistoriaBacklog,
	})
	assert.ErrorIs(t, err, domain.ErrCodigoHistoriaDuplicado, "escenario BDD E-14")
}

func TestHistoriaDeOtroProyectoEsInvisible(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	historias := NuevoHistoriaRepository(db)
	ctx := ctxBackground()

	origen := crearProyectoDePrueba(t, proyectos, "Origen")
	destino := crearProyectoDePrueba(t, proyectos, "Destino")

	historia := &domain.Historia{
		ProyectoID: origen.ID, Numero: 1, Codigo: "US-01", Titulo: "De origen",
		Prioridad: domain.PrioridadMedia, Estado: domain.EstadoHistoriaBacklog,
	}
	require.NoError(t, historias.Crear(ctx, historia))

	_, err := historias.ObtenerPorID(ctx, destino.ID, historia.ID)
	assert.ErrorIs(t, err, domain.ErrHistoriaNoEncontrada, "escenario BDD E-32")

	lista, err := historias.ListarPorProyecto(ctx, destino.ID)
	require.NoError(t, err)
	assert.Empty(t, lista)
}

func TestListarHistoriasSinRegistrosDevuelveVacio(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	historias := NuevoHistoriaRepository(db)

	proyecto := crearProyectoDePrueba(t, proyectos, "Sin backlog")

	lista, err := historias.ListarPorProyecto(ctxBackground(), proyecto.ID)
	require.NoError(t, err)
	assert.NotNil(t, lista, "el listado vacio no debe ser nil")
	assert.Empty(t, lista)
}
