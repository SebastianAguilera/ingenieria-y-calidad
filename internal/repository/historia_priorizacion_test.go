package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func TestHistoriaPersisteValoresDeNegocio(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	historias := NuevoHistoriaRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Con valores de negocio")
	valor := func(v float64) *float64 { return &v }

	historia := &domain.Historia{
		ProyectoID: proyecto.ID,
		Numero:     1,
		Codigo:     domain.FormatearCodigoHistoria(1),
		Titulo:     "Sin valores",
		Prioridad:  domain.PrioridadMedia,
		Estado:     domain.EstadoHistoriaBacklog,
	}
	require.NoError(t, historias.Crear(ctx, historia))

	sinValores, err := historias.ObtenerPorID(ctx, proyecto.ID, historia.ID)
	require.NoError(t, err)
	assert.Nil(t, sinValores.ValorNegocio, "una historia nueva no tiene valor de negocio")
	assert.Nil(t, sinValores.CriticidadTiempo, "una historia nueva no tiene criticidad de tiempo")
	assert.Nil(t, sinValores.ReduccionRiesgo, "una historia nueva no tiene reduccion de riesgo")

	sinValores.StoryPoints = valor(5)
	sinValores.ValorNegocio = valor(8)
	sinValores.CriticidadTiempo = valor(3)
	sinValores.ReduccionRiesgo = valor(2)
	require.NoError(t, historias.Actualizar(ctx, sinValores))

	recuperada, err := historias.ObtenerPorID(ctx, proyecto.ID, historia.ID)
	require.NoError(t, err)
	require.NotNil(t, recuperada.StoryPoints)
	assert.Equal(t, 5.0, *recuperada.StoryPoints)
	require.NotNil(t, recuperada.ValorNegocio)
	assert.Equal(t, 8.0, *recuperada.ValorNegocio)
	require.NotNil(t, recuperada.CriticidadTiempo)
	assert.Equal(t, 3.0, *recuperada.CriticidadTiempo)
	require.NotNil(t, recuperada.ReduccionRiesgo)
	assert.Equal(t, 2.0, *recuperada.ReduccionRiesgo)

	conValores := &domain.Historia{
		ProyectoID:       proyecto.ID,
		Numero:           2,
		Codigo:           domain.FormatearCodigoHistoria(2),
		Titulo:           "Con valores en el alta",
		Prioridad:        domain.PrioridadAlta,
		Estado:           domain.EstadoHistoriaBacklog,
		ValorNegocio:     valor(13),
		CriticidadTiempo: valor(5),
		ReduccionRiesgo:  valor(1),
	}
	require.NoError(t, historias.Crear(ctx, conValores))

	recuperada, err = historias.ObtenerPorID(ctx, proyecto.ID, conValores.ID)
	require.NoError(t, err)
	require.NotNil(t, recuperada.ValorNegocio)
	assert.Equal(t, 13.0, *recuperada.ValorNegocio)
	require.NotNil(t, recuperada.CriticidadTiempo)
	assert.Equal(t, 5.0, *recuperada.CriticidadTiempo)
	require.NotNil(t, recuperada.ReduccionRiesgo)
	assert.Equal(t, 1.0, *recuperada.ReduccionRiesgo)

	listado, err := historias.ListarPorProyecto(ctx, proyecto.ID)
	require.NoError(t, err)
	require.Len(t, listado, 2)
	require.NotNil(t, listado[0].ValorNegocio)
	assert.Equal(t, 8.0, *listado[0].ValorNegocio)
}

// CA-12 / E-10: una base con historias previas a las columnas queda con los tres
// valores en NULL y sin pérdida de datos al migrar.
func TestMigrarEsquemaAgregaValoresDeNegocioNullableSobreFilasExistentes(t *testing.T) {
	db := abrirBaseDePruebas(t)
	proyectos := NuevoProyectoRepository(db)
	historias := NuevoHistoriaRepository(db)
	ctx := ctxBackground()

	proyecto := crearProyectoDePrueba(t, proyectos, "Base existente sin valores")
	previa := &domain.Historia{
		ProyectoID:  proyecto.ID,
		Numero:      1,
		Codigo:      domain.FormatearCodigoHistoria(1),
		Titulo:      "Historia previa",
		Descripcion: "Creada antes de las columnas",
		Prioridad:   domain.PrioridadAlta,
		Estado:      domain.EstadoHistoriaEnProgreso,
	}
	require.NoError(t, historias.Crear(ctx, previa))

	columnas := []string{"valor_negocio", "criticidad_tiempo", "reduccion_riesgo"}
	for _, columna := range columnas {
		require.NoError(t, db.Exec("ALTER TABLE historias DROP COLUMN "+columna).Error)
		require.False(t, db.Migrator().HasColumn("historias", columna))
	}

	require.NoError(t, MigrarEsquema(db))

	for _, columna := range columnas {
		assert.True(t, db.Migrator().HasColumn("historias", columna), "la migración agrega %s", columna)

		var fila struct {
			IsNullable string
			ColumnDef  *string
		}
		require.NoError(t, db.Raw(`SELECT is_nullable AS is_nullable, column_default AS column_def
			FROM information_schema.columns
			WHERE table_name = 'historias' AND column_name = ?`, columna).Scan(&fila).Error)
		assert.Equal(t, "YES", fila.IsNullable, "%s debe ser nullable", columna)
		assert.Nil(t, fila.ColumnDef, "%s no debe tener valor por defecto", columna)
	}

	recuperada, err := historias.ObtenerPorID(ctx, proyecto.ID, previa.ID)
	require.NoError(t, err)
	assert.Nil(t, recuperada.ValorNegocio, "la fila previa queda con valor_negocio en null")
	assert.Nil(t, recuperada.CriticidadTiempo, "la fila previa queda con criticidad_tiempo en null")
	assert.Nil(t, recuperada.ReduccionRiesgo, "la fila previa queda con reduccion_riesgo en null")
	assert.Equal(t, "Historia previa", recuperada.Titulo)
	assert.Equal(t, "Creada antes de las columnas", recuperada.Descripcion)
	assert.Equal(t, domain.PrioridadAlta, recuperada.Prioridad)
	assert.Equal(t, domain.EstadoHistoriaEnProgreso, recuperada.Estado)
}
