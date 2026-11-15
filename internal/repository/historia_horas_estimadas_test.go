package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func TestHistoriaPersisteHorasEstimadas(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	historias := NuevoHistoriaRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Con horas estimadas")
	horas := func(v float64) *float64 { return &v }

	historia := &domain.Historia{
		ProyectoID: proyecto.ID,
		Numero:     1,
		Codigo:     domain.FormatearCodigoHistoria(1),
		Titulo:     "Sin horas",
		Prioridad:  domain.PrioridadMedia,
		Estado:     domain.EstadoHistoriaBacklog,
	}
	require.NoError(t, historias.Crear(ctx, historia))

	sinHoras, err := historias.ObtenerPorID(ctx, proyecto.ID, historia.ID)
	require.NoError(t, err)
	assert.Nil(t, sinHoras.HorasEstimadas, "una historia nueva no tiene horas estimadas")

	sinHoras.HorasEstimadas = horas(8)
	sinHoras.StoryPoints = horas(5)
	sinHoras.EsfuerzoReal = horas(3)
	require.NoError(t, historias.Actualizar(ctx, sinHoras))

	recuperada, err := historias.ObtenerPorID(ctx, proyecto.ID, historia.ID)
	require.NoError(t, err)
	require.NotNil(t, recuperada.HorasEstimadas)
	assert.Equal(t, 8.0, *recuperada.HorasEstimadas)
	require.NotNil(t, recuperada.StoryPoints)
	assert.Equal(t, 5.0, *recuperada.StoryPoints, "las horas estimadas no pisan story_points")
	require.NotNil(t, recuperada.EsfuerzoReal)
	assert.Equal(t, 3.0, *recuperada.EsfuerzoReal, "las horas estimadas no pisan esfuerzo_real")

	conHoras := &domain.Historia{
		ProyectoID:     proyecto.ID,
		Numero:         2,
		Codigo:         domain.FormatearCodigoHistoria(2),
		Titulo:         "Con horas en el alta",
		Prioridad:      domain.PrioridadMedia,
		Estado:         domain.EstadoHistoriaBacklog,
		HorasEstimadas: horas(0.25),
	}
	require.NoError(t, historias.Crear(ctx, conHoras))

	recuperada, err = historias.ObtenerPorID(ctx, proyecto.ID, conHoras.ID)
	require.NoError(t, err)
	require.NotNil(t, recuperada.HorasEstimadas)
	assert.Equal(t, 0.25, *recuperada.HorasEstimadas)

	listado, err := historias.ListarPorProyecto(ctx, proyecto.ID)
	require.NoError(t, err)
	encontrado := false
	for _, h := range listado {
		if h.ID == conHoras.ID {
			require.NotNil(t, h.HorasEstimadas)
			assert.Equal(t, 0.25, *h.HorasEstimadas)
			encontrado = true
		}
	}
	assert.True(t, encontrado, "la historia con horas estimadas debe aparecer en el listado")
}

// CA-07 / BDD-12: una base con historias previas a la columna queda con
// horas_estimadas en NULL y sin pérdida de datos al migrar.
func TestMigrarEsquemaAgregaHorasEstimadasNullableSobreFilasExistentes(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	historias := NuevoHistoriaRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Base existente")
	previa := &domain.Historia{
		ProyectoID:  proyecto.ID,
		Numero:      1,
		Codigo:      domain.FormatearCodigoHistoria(1),
		Titulo:      "Historia previa",
		Descripcion: "Creada antes de la columna",
		Prioridad:   domain.PrioridadAlta,
		Estado:      domain.EstadoHistoriaEnProgreso,
	}
	require.NoError(t, historias.Crear(ctx, previa))

	// Se simula el esquema anterior a US-19 quitando la columna.
	require.NoError(t, db.Exec("ALTER TABLE historias DROP COLUMN horas_estimadas").Error)
	require.False(t, db.Migrator().HasColumn("historias", "horas_estimadas"))

	require.NoError(t, MigrarEsquema(db))

	assert.True(t, db.Migrator().HasColumn("historias", "horas_estimadas"), "la migración agrega la columna")

	var fila struct {
		IsNullable string
		ColumnDef  *string
	}
	require.NoError(t, db.Raw(`SELECT is_nullable AS is_nullable, column_default AS column_def
		FROM information_schema.columns
		WHERE table_name = 'historias' AND column_name = 'horas_estimadas'`).Scan(&fila).Error)
	assert.Equal(t, "YES", fila.IsNullable, "la columna es nullable")
	assert.Nil(t, fila.ColumnDef, "la columna no tiene valor por defecto")

	recuperada, err := historias.ObtenerPorID(ctx, proyecto.ID, previa.ID)
	require.NoError(t, err)
	assert.Nil(t, recuperada.HorasEstimadas, "la fila previa queda con horas_estimadas en null")
	assert.Equal(t, "Historia previa", recuperada.Titulo)
	assert.Equal(t, "Creada antes de la columna", recuperada.Descripcion)
	assert.Equal(t, domain.PrioridadAlta, recuperada.Prioridad)
	assert.Equal(t, domain.EstadoHistoriaEnProgreso, recuperada.Estado)
}
