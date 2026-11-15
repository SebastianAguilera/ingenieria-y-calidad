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

func TestHistoriaPersisteStoryPoints(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	historias := NuevoHistoriaRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Con estimaciones")
	puntos := func(v float64) *float64 { return &v }

	historia := &domain.Historia{
		ProyectoID: proyecto.ID,
		Numero:     1,
		Codigo:     domain.FormatearCodigoHistoria(1),
		Titulo:     "Sin estimar",
		Prioridad:  domain.PrioridadMedia,
		Estado:     domain.EstadoHistoriaBacklog,
	}
	require.NoError(t, historias.Crear(ctx, historia))

	sinEstimar, err := historias.ObtenerPorID(ctx, proyecto.ID, historia.ID)
	require.NoError(t, err)
	assert.Nil(t, sinEstimar.StoryPoints, "una historia nueva no tiene estimacion")

	for _, valor := range []float64{5, 0, 0.5, 21} {
		sinEstimar.StoryPoints = puntos(valor)
		require.NoError(t, historias.Actualizar(ctx, sinEstimar))

		recuperada, err := historias.ObtenerPorID(ctx, proyecto.ID, historia.ID)
		require.NoError(t, err)
		require.NotNil(t, recuperada.StoryPoints, "story points %v", valor)
		assert.Equal(t, valor, *recuperada.StoryPoints)
	}

	conEstimacion := &domain.Historia{
		ProyectoID:  proyecto.ID,
		Numero:      2,
		Codigo:      domain.FormatearCodigoHistoria(2),
		Titulo:      "Estimada en el alta",
		Prioridad:   domain.PrioridadMedia,
		Estado:      domain.EstadoHistoriaBacklog,
		StoryPoints: puntos(13),
	}
	require.NoError(t, historias.Crear(ctx, conEstimacion))

	recuperada, err := historias.ObtenerPorID(ctx, proyecto.ID, conEstimacion.ID)
	require.NoError(t, err)
	require.NotNil(t, recuperada.StoryPoints)
	assert.Equal(t, 13.0, *recuperada.StoryPoints)
}

func TestHistoriaPersisteEsfuerzoReal(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	historias := NuevoHistoriaRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Con esfuerzo real")
	horas := func(v float64) *float64 { return &v }

	historia := &domain.Historia{
		ProyectoID: proyecto.ID,
		Numero:     1,
		Codigo:     domain.FormatearCodigoHistoria(1),
		Titulo:     "Sin esfuerzo",
		Prioridad:  domain.PrioridadMedia,
		Estado:     domain.EstadoHistoriaBacklog,
	}
	require.NoError(t, historias.Crear(ctx, historia))

	sinEsfuerzo, err := historias.ObtenerPorID(ctx, proyecto.ID, historia.ID)
	require.NoError(t, err)
	assert.Nil(t, sinEsfuerzo.EsfuerzoReal, "una historia nueva no tiene esfuerzo real")

	sinEsfuerzo.EsfuerzoReal = horas(4.5)
	require.NoError(t, historias.Actualizar(ctx, sinEsfuerzo))

	recuperada, err := historias.ObtenerPorID(ctx, proyecto.ID, historia.ID)
	require.NoError(t, err)
	require.NotNil(t, recuperada.EsfuerzoReal)
	assert.Equal(t, 4.5, *recuperada.EsfuerzoReal)

	conEsfuerzo := &domain.Historia{
		ProyectoID:   proyecto.ID,
		Numero:       2,
		Codigo:       domain.FormatearCodigoHistoria(2),
		Titulo:       "Con esfuerzo en el alta",
		Prioridad:    domain.PrioridadMedia,
		Estado:       domain.EstadoHistoriaBacklog,
		EsfuerzoReal: horas(6),
	}
	require.NoError(t, historias.Crear(ctx, conEsfuerzo))

	recuperada, err = historias.ObtenerPorID(ctx, proyecto.ID, conEsfuerzo.ID)
	require.NoError(t, err)
	require.NotNil(t, recuperada.EsfuerzoReal)
	assert.Equal(t, 6.0, *recuperada.EsfuerzoReal)

	listado, err := historias.ListarPorProyecto(ctx, proyecto.ID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(listado), 2)
	encontrado := false
	for _, h := range listado {
		if h.ID == conEsfuerzo.ID {
			require.NotNil(t, h.EsfuerzoReal)
			assert.Equal(t, 6.0, *h.EsfuerzoReal)
			encontrado = true
			break
		}
	}
	assert.True(t, encontrado, "la historia con esfuerzo real debe aparecer en el listado")
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

func crearHistoriaConTitulo(t *testing.T, repo *historiaRepository, proyectoID uint, numero int, titulo string) *domain.Historia {
	t.Helper()
	historia := &domain.Historia{
		ProyectoID: proyectoID,
		Numero:     numero,
		Codigo:     domain.FormatearCodigoHistoria(numero),
		Titulo:     titulo,
		Prioridad:  domain.PrioridadMedia,
		Estado:     domain.EstadoHistoriaBacklog,
	}
	require.NoError(t, repo.Crear(ctxBackground(), historia))
	return historia
}

func TestHistoriaObtenerPorIDs(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	historias := NuevoHistoriaRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Con historias por lote")
	primera := crearHistoriaConTitulo(t, historias, proyecto.ID, 1, "Primera")
	segunda := crearHistoriaConTitulo(t, historias, proyecto.ID, 2, "Segunda")
	tercera := crearHistoriaConTitulo(t, historias, proyecto.ID, 3, "Tercera")

	otroProyecto := crearProyectoDePrueba(t, proyectos, "Otro proyecto")
	ajena := crearHistoriaConTitulo(t, historias, otroProyecto.ID, 1, "Ajena")

	listado, err := historias.ObtenerPorIDs(ctx, proyecto.ID, []uint{primera.ID, tercera.ID})
	require.NoError(t, err)
	require.Len(t, listado, 2)
	assert.Equal(t, []string{"US-01", "US-03"}, []string{listado[0].Codigo, listado[1].Codigo})

	listado, err = historias.ObtenerPorIDs(ctx, proyecto.ID, []uint{primera.ID, ajena.ID})
	require.NoError(t, err)
	require.Len(t, listado, 1, "no debe devolver historias de otro proyecto")
	assert.Equal(t, primera.ID, listado[0].ID)

	vacio, err := historias.ObtenerPorIDs(ctx, proyecto.ID, nil)
	require.NoError(t, err)
	assert.NotNil(t, vacio)
	assert.Empty(t, vacio)

	_ = segunda
}
